package autoeco

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gitsense/gitsense/backend/internal/louvain"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

// noiseTopics G74 噪声 topic 黑名单
var noiseTopics = map[string]bool{
	"hacktoberfest": true,
	"awesome":       true,
	"awesomelist":   true,
	"tutorial":      true,
	"example":       true,
	"sample":        true,
	"learning":      true,
	"demo":          true,
	"windows":       true,
	"linux":         true,
	"macos":         true,
}

// Builder 自动生态发现构建器
type Builder struct {
	pool       *pgxpool.Pool
	classifier *service.EcosystemClassifier

	pendingAssignments  []repoAssignData
	pendingBuildVersion int
}

// NewBuilder 创建 Builder
func NewBuilder(pool *pgxpool.Pool, classifier *service.EcosystemClassifier) *Builder {
	return &Builder{pool: pool, classifier: classifier}
}

// BuildResult 构建结果
type BuildResult struct {
	BuildVersion     int     `json:"build_version"`
	ClustersFound    int     `json:"clusters_found"`
	ReposAssigned    int     `json:"repos_assigned"`
	CoveragePct      float64 `json:"coverage_pct"`
	EmergingCount    int     `json:"emerging_count"`
	DurationMs       int     `json:"duration_ms"`
	SuperCommunities int     `json:"super_communities"`
	HierarchyLevels  int     `json:"hierarchy_levels"`
}

// CommunityInfo 社区信息（用于报告）
type CommunityInfo struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	RuleMapping *string  `json:"rule_mapping"`
	IsEmerging  bool     `json:"is_emerging"`
	RepoCount   int      `json:"repo_count"`
	TopicCount  int      `json:"topic_count"`
	TopTopics   []string `json:"top_topics"`
	TrendScore  float64  `json:"trend_score"`
	TopRepos    []string `json:"top_repos"`
	ParentID    *int     `json:"parent_id"`
	Level       int      `json:"level"`
	IsSuper     bool     `json:"is_super"`
}

// Report 验证报告
type Report struct {
	BuildVersion   int             `json:"build_version"`
	CoveragePct    float64         `json:"coverage_pct"`
	PurityPct      float64         `json:"purity_pct"`
	EmergingCount  int             `json:"emerging_count"`
	ClusterCount   int             `json:"cluster_count"`
	LargestCluster int             `json:"largest_cluster"`
	AvgClusterSize float64         `json:"avg_cluster_size"`
	Communities    []CommunityInfo `json:"communities"`
	HierarchyStats HierarchyStats  `json:"hierarchy_stats"`
}

// HierarchyStats 层级统计 (G73)
type HierarchyStats struct {
	Level0Count int `json:"level_0_count"`
	Level1Count int `json:"level_1_count"`
	SuperCount  int `json:"super_count"`
}

// HubPenaltyBenchmark Hub Topic Penalty 基准测试结果 (Task B)
type HubPenaltyBenchmark struct {
	WithoutPenalty PenaltyMetrics `json:"without_penalty"`
	WithPenalty    PenaltyMetrics `json:"with_penalty"`
}

// PenaltyMetrics 惩罚指标
type PenaltyMetrics struct {
	ClusterCount   int     `json:"cluster_count"`
	LargestCluster int     `json:"largest_cluster"`
	CoveragePct    float64 `json:"coverage_pct"`
	PurityPct      float64 `json:"purity_pct"`
}

// ecoNode 内部生态节点（构建过程中使用）
type ecoNode struct {
	idx         int      // 在 nodes 列表中的索引
	name        string
	level       int      // 0 = root, 1 = sub
	parentIdx   int      // -1 for root
	topics      []string // 所有 topics（归一化后）
	topTopics   []string // Top N topics by weight（过滤噪声后）
	ruleMapping *string
	isEmerging  bool
	isSuper     bool // 是否是被拆分的超级社区
	repoCount   int
	trendScore  float64
	children    []int // 子节点索引
}

// repoAssignData 暂存 repo 分配结果
type repoAssignData struct {
	repoID         int
	nodeIdx        int // ecoNode 索引
	score          float64
	assignedTopics []string
}

// normalizeTopic 轻量归一化：lowercase + 去连字符/下划线/空格
func normalizeTopic(topic string) string {
	s := strings.ToLower(topic)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

// isNoiseTopic 检查是否为噪声 topic (G74)
func isNoiseTopic(normTopic string) bool {
	return noiseTopics[normTopic]
}

// filterNoiseTopics 过滤噪声 topics (G74)
func filterNoiseTopics(topics []string) []string {
	filtered := make([]string, 0, len(topics))
	for _, t := range topics {
		if !isNoiseTopic(t) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// Build 执行完整的自动生态发现流程 (G71-G74)
func (b *Builder) Build(ctx context.Context) (*BuildResult, error) {
	return b.buildInternal(ctx, false)
}

// BuildWithPenalty 执行带 Hub Penalty 的构建 (Task B)
func (b *Builder) BuildWithPenalty(ctx context.Context) (*BuildResult, error) {
	return b.buildInternal(ctx, true)
}

// buildInternal 内部构建逻辑
func (b *Builder) buildInternal(ctx context.Context, hubPenalty bool) (*BuildResult, error) {
	start := time.Now()

	// 获取下一个 build_version
	var nextVersion int
	err := b.pool.QueryRow(ctx, `SELECT COALESCE(MAX(build_version), 0) + 1 FROM auto_ecosystem_builds`).Scan(&nextVersion)
	if err != nil {
		return nil, fmt.Errorf("get next build version: %w", err)
	}

	_, err = b.pool.Exec(ctx,
		`INSERT INTO auto_ecosystem_builds (build_version, status) VALUES ($1, 'building')`,
		nextVersion)
	if err != nil {
		return nil, fmt.Errorf("insert build record: %w", err)
	}

	// Step 1: 加载 Topic Graph
	graph, normMap, err := b.loadTopicGraph(ctx)
	if err != nil {
		b.markFailed(ctx, nextVersion)
		return nil, fmt.Errorf("load topic graph: %w", err)
	}

	// Step 2: 归一化 + 合并边
	mergedGraph := b.mergeNormalizedEdges(graph, normMap)

	// Step 3: Louvain Level 0 (G62: minSize=3)
	result0 := louvain.DetectWithPenalty(mergedGraph, 3, hubPenalty)

	log.Printf("[AutoEco] Level 0 Louvain: %d communities (Q=%.4f), hubPenalty=%v",
		len(result0.Communities), result0.Modularity, hubPenalty)

	// Step 4: 快速估算每个社区的 repo 数量 (G71)
	repoCounts := b.estimateRepoCounts(ctx, result0)

	// Step 5: G71 - 检测超级社区
	threshold, stats := b.computeSuperThreshold(repoCounts, len(result0.Communities))
	superCommIDs := make(map[int]bool)
	for commID, count := range repoCounts {
		if count > threshold {
			superCommIDs[commID] = true
		}
	}

	log.Printf("[AutoEco] G71 Super Community Detection: threshold=%d, found=%d, stats={mean=%.1f, median=%.1f, std=%.1f, largest=%d}",
		threshold, len(superCommIDs), stats.Mean, stats.Median, stats.Std, stats.Largest)

	// Step 6: G72 - 递归 Louvain + 构建层级结构
	nodes, unifiedNodeMap := b.buildHierarchy(mergedGraph, result0, superCommIDs, repoCounts)

	log.Printf("[AutoEco] G72-G73 Hierarchy: %d nodes (%d level-0, %d level-1, %d super)",
		len(nodes), countLevel(nodes, 0), countLevel(nodes, 1), countSuper(nodes))

	// Step 7: G74 - Naming Pipeline v2
	b.nameAllNodes(nodes, mergedGraph)

	// Step 8: 分配 Repos 到叶节点
	reposAssigned, coveragePct, err := b.assignReposToLeaves(ctx, nodes, unifiedNodeMap, nextVersion)
	if err != nil {
		b.markFailed(ctx, nextVersion)
		return nil, fmt.Errorf("assign repos: %w", err)
	}

	// Step 8.5: Embedding Fallback (G63 Level 2, G70 动态阈值)
	embAssigned, err := b.embeddingFallbackV2(ctx, nodes, unifiedNodeMap, nextVersion)
	if err != nil {
		log.Printf("[AutoEco] Embedding fallback error: %v", err)
	} else {
		reposAssigned += embAssigned
		var totalRepos int
		b.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories`).Scan(&totalRepos)
		if totalRepos > 0 {
			coveragePct = float64(reposAssigned) / float64(totalRepos) * 100
		}
	}

	// Step 9: 持久化社区（含层级）
	emergingCount, err := b.persistHierarchy(ctx, nodes, mergedGraph, nextVersion)
	if err != nil {
		b.markFailed(ctx, nextVersion)
		return nil, fmt.Errorf("persist communities: %w", err)
	}

	// Step 10: 清理旧版本
	b.cleanupOldVersions(ctx, nextVersion)

	duration := time.Since(start)
	maxLevel := 0
	for _, n := range nodes {
		if n.level > maxLevel {
			maxLevel = n.level
		}
	}

	_, err = b.pool.Exec(ctx,
		`UPDATE auto_ecosystem_builds SET status='completed', clusters_found=$1, repos_assigned=$2,
		 coverage_pct=$3, emerging_count=$4, duration_ms=$5, completed_at=NOW()
		 WHERE build_version=$6`,
		len(nodes), reposAssigned, coveragePct, emergingCount, int(duration.Milliseconds()), nextVersion)
	if err != nil {
		return nil, fmt.Errorf("update build record: %w", err)
	}

	log.Printf("[AutoEco] Build v%d complete: %d nodes, %d repos (%.1f%% coverage), %d emerging, %d super, %dms",
		nextVersion, len(nodes), reposAssigned, coveragePct, emergingCount, len(superCommIDs), int(duration.Milliseconds()))

	return &BuildResult{
		BuildVersion:     nextVersion,
		ClustersFound:    len(nodes),
		ReposAssigned:    reposAssigned,
		CoveragePct:      coveragePct,
		EmergingCount:    emergingCount,
		DurationMs:       int(duration.Milliseconds()),
		SuperCommunities: len(superCommIDs),
		HierarchyLevels:  maxLevel + 1,
	}, nil
}

// loadTopicGraph 从 topic_edges 加载图
func (b *Builder) loadTopicGraph(ctx context.Context) (*louvain.Graph, map[string][]string, error) {
	graph := louvain.NewGraph()
	normMap := make(map[string][]string)

	rows, err := b.pool.Query(ctx, `SELECT topic_a, topic_b, weight FROM topic_edges`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a, b2 string
		var weight float64
		if err := rows.Scan(&a, &b2, &weight); err != nil {
			continue
		}
		normA := normalizeTopic(a)
		normB := normalizeTopic(b2)
		graph.AddEdge(normA, normB, weight)

		if _, ok := normMap[normA]; !ok {
			normMap[normA] = []string{a}
		} else {
			found := false
			for _, existing := range normMap[normA] {
				if existing == a {
					found = true
					break
				}
			}
			if !found {
				normMap[normA] = append(normMap[normA], a)
			}
		}
		if _, ok := normMap[normB]; !ok {
			normMap[normB] = []string{b2}
		} else {
			found := false
			for _, existing := range normMap[normB] {
				if existing == b2 {
					found = true
					break
				}
			}
			if !found {
				normMap[normB] = append(normMap[normB], b2)
			}
		}
	}

	return graph, normMap, nil
}

// mergeNormalizedEdges 合并归一化后相同节点的边
func (b *Builder) mergeNormalizedEdges(graph *louvain.Graph, normMap map[string][]string) *louvain.Graph {
	return graph
}

// estimateRepoCounts 快速估算每个社区的 repo 数量 (G71)
func (b *Builder) estimateRepoCounts(ctx context.Context, result *louvain.Result) map[int]int {
	counts := make(map[int]int)

	rows, err := b.pool.Query(ctx, `SELECT id, topics FROM repositories WHERE topics IS NOT NULL AND jsonb_array_length(topics) > 0`)
	if err != nil {
		return counts
	}
	defer rows.Close()

	for rows.Next() {
		var repoID int
		var topicsJSON []byte
		if err := rows.Scan(&repoID, &topicsJSON); err != nil {
			continue
		}

		var topics []string
		if err := json.Unmarshal(topicsJSON, &topics); err != nil {
			continue
		}

		communityScores := make(map[int]int)
		for _, topic := range topics {
			norm := normalizeTopic(topic)
			if commID, ok := result.NodeMap[norm]; ok && commID >= 0 {
				communityScores[commID]++
			}
		}

		if len(communityScores) == 0 {
			continue
		}

		bestComm := -1
		bestScore := 0
		for comm, score := range communityScores {
			if score > bestScore {
				bestScore = score
				bestComm = comm
			}
		}

		if bestComm >= 0 {
			counts[bestComm]++
		}
	}

	return counts
}

// communitySizeStats 社区大小统计
type communitySizeStats struct {
	Mean    float64
	Median  float64
	Std     float64
	Largest int
	Total   int
}

// computeSuperThreshold G71: 计算超级社区阈值
// threshold = max(500, mean + 2*std)
func (b *Builder) computeSuperThreshold(repoCounts map[int]int, totalCommunities int) (int, communitySizeStats) {
	if len(repoCounts) == 0 {
		return 500, communitySizeStats{}
	}

	sizes := make([]float64, 0, len(repoCounts))
	largest := 0
	for _, count := range repoCounts {
		sizes = append(sizes, float64(count))
		if count > largest {
			largest = count
		}
	}

	mean := float64Mean(sizes)
	std := float64Std(sizes)

	sort.Float64s(sizes)
	median := sizes[len(sizes)/2]

	threshold := int(math.Max(500, mean+2*std))

	return threshold, communitySizeStats{
		Mean:    mean,
		Median:  median,
		Std:     std,
		Largest: largest,
		Total:   len(repoCounts),
	}
}

// buildHierarchy G72-G73: 构建层级结构（支持 depth=2 递归拆分）
func (b *Builder) buildHierarchy(graph *louvain.Graph, result0 *louvain.Result, superCommIDs map[int]bool, repoCounts map[int]int) ([]*ecoNode, map[string]int) {
	nodes := make([]*ecoNode, 0)
	unifiedNodeMap := make(map[string]int) // normalized topic → node index

	// 计算超级社区阈值（用于 Level 1 子社区判断）
	threshold, _ := b.computeSuperThreshold(repoCounts, len(result0.Communities))

	for _, comm := range result0.Communities {
		if superCommIDs[comm.ID] {
			// G72: 超级社区 → 递归 Louvain
			parentIdx := len(nodes)
			parentNode := &ecoNode{
				idx:       parentIdx,
				level:     0,
				parentIdx: -1,
				topics:    comm.Nodes,
				isSuper:   true,
				repoCount: repoCounts[comm.ID],
			}
			nodes = append(nodes, parentNode)

			// 提取子图
			subGraph := louvain.SubGraph(graph, comm.Nodes)

			if len(subGraph.Nodes) < 3 {
				parentNode.isSuper = false
				for _, topic := range comm.Nodes {
					unifiedNodeMap[topic] = parentIdx
				}
				continue
			}

			// Level 1: 运行 Louvain
			subResult := louvain.Detect(subGraph, 3)

			if len(subResult.Communities) <= 1 {
				parentNode.isSuper = false
				for _, topic := range comm.Nodes {
					unifiedNodeMap[topic] = parentIdx
				}
				continue
			}

			log.Printf("[AutoEco] G72 Level 1 Recursive Louvain: parent=%d topics → %d sub-communities",
				len(comm.Nodes), len(subResult.Communities))

			// 估算 Level 1 子社区的 repo 数量
			subRepoCounts := b.estimateSubRepoCounts(subResult, repoCounts[comm.ID])

			// 创建 Level 1 子社区
			for _, subComm := range subResult.Communities {
				childIdx := len(nodes)
				childNode := &ecoNode{
					idx:       childIdx,
					level:     1,
					parentIdx: parentIdx,
					topics:    subComm.Nodes,
					repoCount: subRepoCounts[subComm.ID],
				}
				nodes = append(nodes, childNode)
				parentNode.children = append(parentNode.children, childIdx)

				// G72 depth=2: 如果子社区仍然超过阈值，继续拆分
				if subRepoCounts[subComm.ID] > threshold && len(subComm.Nodes) >= 3 {
					childNode.isSuper = true

					subSubGraph := louvain.SubGraph(graph, subComm.Nodes)
					subSubResult := louvain.Detect(subSubGraph, 3)

					if len(subSubResult.Communities) > 1 {
						log.Printf("[AutoEco] G72 Level 2 Recursive Louvain: sub-parent=%d topics → %d sub-sub-communities",
							len(subComm.Nodes), len(subSubResult.Communities))

						for _, subSubComm := range subSubResult.Communities {
							grandchildIdx := len(nodes)
							grandchildNode := &ecoNode{
								idx:       grandchildIdx,
								level:     2,
								parentIdx: childIdx,
								topics:    subSubComm.Nodes,
							}
							nodes = append(nodes, grandchildNode)
							childNode.children = append(childNode.children, grandchildIdx)

							for _, topic := range subSubComm.Nodes {
								unifiedNodeMap[topic] = grandchildIdx
							}
						}
						continue
					}
					// 无法进一步拆分，作为叶节点
					childNode.isSuper = false
				}

				// 映射 topics 到子社区（叶节点）
				for _, topic := range subComm.Nodes {
					unifiedNodeMap[topic] = childIdx
				}
			}
		} else {
			// 非超级社区 → 直接作为叶节点
			idx := len(nodes)
			node := &ecoNode{
				idx:       idx,
				level:     0,
				parentIdx: -1,
				topics:    comm.Nodes,
				repoCount: repoCounts[comm.ID],
			}
			nodes = append(nodes, node)

			for _, topic := range comm.Nodes {
				unifiedNodeMap[topic] = idx
			}
		}
	}

	// 下沉超级社区中未被子社区覆盖的 orphaned topics
	// 对于每个超级社区父节点，找出其 topics 中不在任何子节点 unifiedNodeMap 中的 topics，
	// 将它们映射到拥有最多邻居关系的子节点
	for _, node := range nodes {
		if !node.isSuper || len(node.children) == 0 {
			continue
		}

		// 收集子节点已覆盖的 topics
		childCoveredTopics := make(map[string]bool)
		for _, childIdx := range node.children {
			for _, t := range nodes[childIdx].topics {
				childCoveredTopics[t] = true
			}
		}

		// 找出 orphaned topics（在父节点中但不在任何子节点中）
		var orphanedTopics []string
		for _, t := range node.topics {
			if !childCoveredTopics[t] {
				orphanedTopics = append(orphanedTopics, t)
			}
		}

		if len(orphanedTopics) == 0 {
			continue
		}

		// 将 orphaned topics 映射到最相关的子节点
		// 策略：在 topic_edges 图中，orphaned topic 与哪个子节点的 topics 连接最多，就映射到那个子节点
		for _, orphan := range orphanedTopics {
			bestChild := -1
			bestOverlap := 0

			for _, childIdx := range node.children {
				overlap := 0
				for _, childTopic := range nodes[childIdx].topics {
					// 检查图中有无连接（通过邻居关系）
					if neighbors, ok := graph.Adj[orphan]; ok {
						if _, ok2 := neighbors[childTopic]; ok2 {
							overlap++
						}
					}
				}
				if overlap > bestOverlap {
					bestOverlap = overlap
					bestChild = childIdx
				}
			}

			// 如果没有图连接，映射到 topics 最多的子节点（最大子社区）
			if bestChild < 0 && len(node.children) > 0 {
				maxTopics := 0
				for _, childIdx := range node.children {
					if len(nodes[childIdx].topics) > maxTopics {
						maxTopics = len(nodes[childIdx].topics)
						bestChild = childIdx
					}
				}
			}

			if bestChild >= 0 {
				unifiedNodeMap[orphan] = bestChild
				// 将 orphaned topic 添加到子节点的 topics 中
				nodes[bestChild].topics = append(nodes[bestChild].topics, orphan)
			}
		}

		log.Printf("[AutoEco] Sunk %d orphaned topics from super community to children", len(orphanedTopics))
	}

	return nodes, unifiedNodeMap
}

// estimateSubRepoCounts 按 topic 比例估算子社区 repo 数量
func (b *Builder) estimateSubRepoCounts(subResult *louvain.Result, parentRepoCount int) map[int]int {
	counts := make(map[int]int)
	totalTopics := 0
	for _, comm := range subResult.Communities {
		totalTopics += len(comm.Nodes)
	}
	if totalTopics == 0 {
		return counts
	}
	for _, comm := range subResult.Communities {
		ratio := float64(len(comm.Nodes)) / float64(totalTopics)
		counts[comm.ID] = int(float64(parentRepoCount) * ratio)
	}
	return counts
}

// nameAllNodes G74: Naming Pipeline v2
func (b *Builder) nameAllNodes(nodes []*ecoNode, graph *louvain.Graph) {
	rules := b.classifier.GetAllRules()

	for _, node := range nodes {
		// 计算社区内 topic 权重
		topicWeight := make(map[string]float64)
		for _, t := range node.topics {
			topicWeight[t] += graph.Degree(t)
		}

		// Top 10 topics by weight
		type tw struct {
			topic  string
			weight float64
		}
		var sorted []tw
		for t, w := range topicWeight {
			sorted = append(sorted, tw{t, w})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].weight > sorted[j].weight
		})

		topTopics := make([]string, 0, 10)
		for i, s := range sorted {
			if i >= 10 {
				break
			}
			topTopics = append(topTopics, s.topic)
		}

		// G74 Step 1: 过滤噪声 topics
		cleanTopics := filterNoiseTopics(topTopics)

		node.topTopics = cleanTopics

		// G74 Step 2: Rule Mapping (overlap >= 1 for better coverage)
		var bestRule *service.EcosystemRule
		bestOverlap := 0
		for i := range rules {
			rule := &rules[i]
			ruleTopicsNorm := make(map[string]bool)
			for _, t := range rule.Topics {
				ruleTopicsNorm[normalizeTopic(t)] = true
			}

			overlap := 0
			for _, t := range cleanTopics {
				if ruleTopicsNorm[t] {
					overlap++
				}
			}
			// 降低阈值：大生态(>=10 topics) overlap>=1 即可，小生态 overlap>=2
			threshold := 2
			if len(rule.Topics) >= 10 {
				threshold = 1
			}
			if overlap >= threshold && overlap > bestOverlap {
				bestRule = rule
				bestOverlap = overlap
			}
		}

		if bestRule != nil {
			node.name = bestRule.Name
			ruleName := bestRule.Name
			node.ruleMapping = &ruleName
			continue
		}

		// G74 Step 3: Subcategory Mapping (overlap >= 1)
		for i := range rules {
			rule := &rules[i]
			for _, sub := range rule.Subcategories {
				subTopicsNorm := make(map[string]bool)
				for _, t := range sub.Topics {
					subTopicsNorm[normalizeTopic(t)] = true
				}

				overlap := 0
				for _, t := range cleanTopics {
					if subTopicsNorm[t] {
						overlap++
					}
				}
				if overlap >= 1 {
					node.name = rule.Name + " / " + sub.Name
					ruleName := rule.Name
					node.ruleMapping = &ruleName
					bestRule = &rules[i] // mark as mapped
					break
				}
			}
			if bestRule != nil {
				break
			}
		}
		if bestRule != nil {
			continue
		}

		// G74 Step 4: TF-IDF Top Topics (过滤噪声后)
		// 不再使用 "xxx + yyy Ecosystem" 格式，改为更可读的 "xxx / yyy" 格式
		if len(cleanTopics) > 0 {
			top3 := cleanTopics
			if len(top3) > 3 {
				top3 = top3[:3]
			}
			// 使用 " / " 分隔，不加 "Ecosystem" 后缀
			node.name = strings.Join(top3, " / ")
		} else {
			// G74 Step 5: Fallback
			node.name = fmt.Sprintf("Community #%d", node.idx)
		}
	}
}

// assignReposToLeaves 分配 repos 到叶节点社区
func (b *Builder) assignReposToLeaves(ctx context.Context, nodes []*ecoNode, unifiedNodeMap map[string]int, buildVersion int) (int, float64, error) {
	rows, err := b.pool.Query(ctx, `SELECT id, topics FROM repositories WHERE topics IS NOT NULL AND jsonb_array_length(topics) > 0`)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var assignments []repoAssignData
	var totalRepos int

	b.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories`).Scan(&totalRepos)

	for rows.Next() {
		var repoID int
		var topicsJSON []byte
		if err := rows.Scan(&repoID, &topicsJSON); err != nil {
			continue
		}

		var topics []string
		if err := json.Unmarshal(topicsJSON, &topics); err != nil {
			continue
		}

		// 统计每个叶节点匹配的 topic 数
		nodeScores := make(map[int]int)
		matchedTopics := make(map[int][]string)

		for _, topic := range topics {
			norm := normalizeTopic(topic)
			if nodeIdx, ok := unifiedNodeMap[norm]; ok {
				// 只分配到叶节点（非 super 的节点）
				if !nodes[nodeIdx].isSuper {
					nodeScores[nodeIdx]++
					matchedTopics[nodeIdx] = append(matchedTopics[nodeIdx], topic)
				}
			}
		}

		if len(nodeScores) == 0 {
			continue
		}

		// 选择匹配最多的叶节点
		bestNode := -1
		bestScore := 0
		for nodeIdx, score := range nodeScores {
			if score > bestScore {
				bestScore = score
				bestNode = nodeIdx
			}
		}

		if bestNode >= 0 {
			assignScore := float64(bestScore) / float64(len(topics))
			assignments = append(assignments, repoAssignData{
				repoID:         repoID,
				nodeIdx:        bestNode,
				score:          assignScore,
				assignedTopics: matchedTopics[bestNode],
			})
		}
	}

	b.pendingAssignments = assignments
	b.pendingBuildVersion = buildVersion

	reposAssigned := len(assignments)
	coveragePct := 0.0
	if totalRepos > 0 {
		coveragePct = float64(reposAssigned) / float64(totalRepos) * 100
	}

	return reposAssigned, coveragePct, nil
}

// embeddingFallbackV2 Embedding Fallback (G63 Level 2, G70 动态阈值)
func (b *Builder) embeddingFallbackV2(ctx context.Context, nodes []*ecoNode, unifiedNodeMap map[string]int, buildVersion int) (int, error) {
	// 找出未分配的 repos
	assigned := make(map[int]bool)
	for _, a := range b.pendingAssignments {
		assigned[a.repoID] = true
	}

	// 获取未分配 repos 的 embedding
	rows, err := b.pool.Query(ctx,
		`SELECT id, embedding FROM repositories WHERE embedding IS NOT NULL AND topics IS NOT NULL AND jsonb_array_length(topics) > 0`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type repoEmb struct {
		id        int
		embedding []float64
	}
	var unassigned []repoEmb

	for rows.Next() {
		var id int
		var embData []byte
		if err := rows.Scan(&id, &embData); err != nil {
			continue
		}
		if assigned[id] {
			continue
		}
		emb, err := parseVector(embData)
		if err != nil {
			continue
		}
		unassigned = append(unassigned, repoEmb{id, emb})
	}

	if len(unassigned) == 0 {
		return 0, nil
	}

	// 计算每个叶节点的 centroid（仅非 super 节点）
	centroidMap := make(map[int][]float64) // nodeIdx → centroid
	centroidCount := make(map[int]int)

	for _, a := range b.pendingAssignments {
		var embData []byte
		err := b.pool.QueryRow(ctx,
			`SELECT embedding FROM repositories WHERE id = $1 AND embedding IS NOT NULL`, a.repoID).Scan(&embData)
		if err != nil {
			continue
		}
		emb, err := parseVector(embData)
		if err != nil {
			continue
		}

		if _, ok := centroidMap[a.nodeIdx]; !ok {
			centroidMap[a.nodeIdx] = make([]float64, len(emb))
		}
		for i, v := range emb {
			centroidMap[a.nodeIdx][i] += v
		}
		centroidCount[a.nodeIdx]++
	}

	// 计算平均值
	for nodeIdx, sum := range centroidMap {
		count := centroidCount[nodeIdx]
		if count > 0 {
			for i := range sum {
				sum[i] /= float64(count)
			}
		}
	}

	if len(centroidMap) == 0 {
		return 0, nil
	}

	// 计算动态阈值统计
	var allSims []float64
	for _, un := range unassigned {
		for _, centroid := range centroidMap {
			sim := cosineSimilarity(un.embedding, centroid)
			allSims = append(allSims, sim)
		}
	}

	if len(allSims) == 0 {
		return 0, nil
	}

	mean := float64Mean(allSims)
	_ = float64Std(allSims) // 保留计算但暂不使用（G70 原始阈值 mean+0.5*std 已放宽）
	threshold := mean // 放宽阈值：从 mean+0.5*std 降到 mean，提升 embedding fallback 覆盖率

	newAssigned := 0
	for _, un := range unassigned {
		type nodeSim struct {
			nodeIdx int
			sim     float64
		}
		var sims []nodeSim
		for nodeIdx, centroid := range centroidMap {
			sim := cosineSimilarity(un.embedding, centroid)
			sims = append(sims, nodeSim{nodeIdx, sim})
		}
		sort.Slice(sims, func(i, j int) bool {
			return sims[i].sim > sims[j].sim
		})

		if len(sims) < 1 {
			continue
		}

		top1 := sims[0]

		// 放宽条件：Top-1 超过 mean 即可分配（移除 Top-2 领先 10% 要求）
		// 如果有 Top-2，仍要求 Top-1 领先
		if top1.sim > threshold {
			if len(sims) >= 2 {
				top2 := sims[1]
				if top1.sim <= top2.sim*1.05 { // 5% 领先即可（从 10% 放宽）
					continue
				}
			}
			b.pendingAssignments = append(b.pendingAssignments, repoAssignData{
				repoID:         un.id,
				nodeIdx:        top1.nodeIdx,
				score:          top1.sim,
				assignedTopics: []string{},
			})
			newAssigned++
		}
	}

	return newAssigned, nil
}

// persistHierarchy 持久化层级社区数据 (G73)
func (b *Builder) persistHierarchy(ctx context.Context, nodes []*ecoNode, graph *louvain.Graph, buildVersion int) (int, error) {
	emergingCount := 0

	// 先插入父节点（level 0 super），再插入子节点和叶节点
	// 两遍扫描：第一遍插入所有节点获取 DB ID，第二遍更新 parent_id

	nodeIdxToDBID := make(map[int]int)

	// 第一遍：插入所有节点（parent_id 暂时为 NULL）
	for _, node := range nodes {
		topTopicsJSON, _ := json.Marshal(node.topTopics)

		var dbID int
		err := b.pool.QueryRow(ctx,
			`INSERT INTO auto_ecosystems (build_version, name, rule_mapping, is_emerging, repo_count, topic_count, top_topics, level, parent_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL)
			 RETURNING id`,
			buildVersion, node.name, node.ruleMapping, false, 0, len(node.topics), topTopicsJSON, node.level,
		).Scan(&dbID)
		if err != nil {
			log.Printf("[AutoEco] Error inserting ecosystem %s: %v", node.name, err)
			continue
		}
		nodeIdxToDBID[node.idx] = dbID
	}

	// 更新 parent_id
	for _, node := range nodes {
		if node.parentIdx >= 0 {
			parentDBID, ok := nodeIdxToDBID[node.parentIdx]
			if !ok {
				continue
			}
			dbID, ok := nodeIdxToDBID[node.idx]
			if !ok {
				continue
			}
			_, err := b.pool.Exec(ctx,
				`UPDATE auto_ecosystems SET parent_id = $1 WHERE id = $2`,
				parentDBID, dbID)
			if err != nil {
				log.Printf("[AutoEco] Error updating parent_id for %s: %v", node.name, err)
			}
		}
	}

	// 写入 auto_ecosystem_topics
	for _, node := range nodes {
		dbID, ok := nodeIdxToDBID[node.idx]
		if !ok {
			continue
		}

		topicWeight := make(map[string]float64)
		for _, t := range node.topics {
			topicWeight[t] += graph.Degree(t)
		}

		for normTopic, weight := range topicWeight {
			_, err := b.pool.Exec(ctx,
				`INSERT INTO auto_ecosystem_topics (ecosystem_id, topic, normalized_topic, weight)
				 VALUES ($1, $2, $3, $4)`,
				dbID, normTopic, normTopic, int(weight),
			)
			if err != nil {
				log.Printf("[AutoEco] Error inserting topic %s: %v", normTopic, err)
			}
		}
	}

	// 写入 auto_ecosystem_repos（仅叶节点）
	for _, a := range b.pendingAssignments {
		dbID, ok := nodeIdxToDBID[a.nodeIdx]
		if !ok {
			continue
		}

		topicsJSON, _ := json.Marshal(a.assignedTopics)
		method := "topic"
		if len(a.assignedTopics) == 0 {
			method = "embedding"
		}

		_, err := b.pool.Exec(ctx,
			`INSERT INTO auto_ecosystem_repos (ecosystem_id, repo_id, build_version, assignment_score, assigned_topics, assignment_method)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (repo_id, ecosystem_id, build_version) DO NOTHING`,
			dbID, a.repoID, buildVersion, a.score, topicsJSON, method,
		)
		if err != nil {
			log.Printf("[AutoEco] Error inserting repo assignment: %v", err)
		}
	}

	// 更新 repo_count, trend_score, is_emerging
	for _, node := range nodes {
		dbID, ok := nodeIdxToDBID[node.idx]
		if !ok {
			continue
		}

		var repoCount int
		if node.isSuper {
			// G73: 父节点 repo_count = 子节点之和
			for _, childIdx := range node.children {
				var childCount int
				childDBID, ok := nodeIdxToDBID[childIdx]
				if !ok {
					continue
				}
				b.pool.QueryRow(ctx,
					`SELECT COUNT(*) FROM auto_ecosystem_repos WHERE ecosystem_id = $1 AND build_version = $2`,
					childDBID, buildVersion).Scan(&childCount)
				repoCount += childCount
			}
		} else {
			b.pool.QueryRow(ctx,
				`SELECT COUNT(*) FROM auto_ecosystem_repos WHERE ecosystem_id = $1 AND build_version = $2`,
				dbID, buildVersion).Scan(&repoCount)
		}

		// 计算 trend_score
		var avgTrend float64
		if node.isSuper {
			// 父节点趋势 = 子节点加权平均
			b.pool.QueryRow(ctx, `
				SELECT COALESCE(AVG(
					CASE WHEN r.pushed_at IS NOT NULL THEN
						CASE WHEN r.pushed_at > NOW() - INTERVAL '7 days' THEN 0.7
						     WHEN r.pushed_at > NOW() - INTERVAL '30 days' THEN 0.5
						     ELSE 0.3 END
					ELSE 0.5 END
				), 0.5)
				FROM auto_ecosystem_repos aer
				JOIN repositories r ON aer.repo_id = r.id
				WHERE aer.ecosystem_id = ANY(SELECT id FROM auto_ecosystems WHERE parent_id = $1 AND build_version = $2)
					AND aer.build_version = $2`,
				dbID, buildVersion).Scan(&avgTrend)
		} else {
			b.pool.QueryRow(ctx, `
				SELECT COALESCE(AVG(
					CASE WHEN r.pushed_at IS NOT NULL THEN
						CASE WHEN r.pushed_at > NOW() - INTERVAL '7 days' THEN 0.7
						     WHEN r.pushed_at > NOW() - INTERVAL '30 days' THEN 0.5
						     ELSE 0.3 END
					ELSE 0.5 END
				), 0.5)
				FROM auto_ecosystem_repos aer
				JOIN repositories r ON aer.repo_id = r.id
				WHERE aer.ecosystem_id = $1 AND aer.build_version = $2`,
				dbID, buildVersion).Scan(&avgTrend)
		}

		// G59: Emerging 判定
		isEmerging := false
		if node.ruleMapping == nil && avgTrend > 0.6 && repoCount >= 5 && !node.isSuper {
			isEmerging = true
		}

		_, err := b.pool.Exec(ctx,
			`UPDATE auto_ecosystems SET repo_count=$1, trend_score=$2, is_emerging=$3 WHERE id=$4`,
			repoCount, avgTrend, isEmerging, dbID)
		if err != nil {
			log.Printf("[AutoEco] Error updating ecosystem %d: %v", dbID, err)
		}

		if isEmerging {
			emergingCount++
		}

		node.repoCount = repoCount
		node.trendScore = avgTrend
		node.isEmerging = isEmerging
	}

	return emergingCount, nil
}

// cleanupOldVersions 清理旧版本 (G69)
func (b *Builder) cleanupOldVersions(ctx context.Context, currentVersion int) {
	if currentVersion <= 2 {
		return
	}
	oldVersion := currentVersion - 2

	_, err := b.pool.Exec(ctx, `DELETE FROM auto_ecosystem_repos WHERE build_version < $1`, oldVersion)
	if err != nil {
		log.Printf("[AutoEco] Error cleaning old repo assignments: %v", err)
	}

	_, err = b.pool.Exec(ctx, `DELETE FROM auto_ecosystems WHERE build_version < $1`, oldVersion)
	if err != nil {
		log.Printf("[AutoEco] Error cleaning old ecosystems: %v", err)
	}

	_, err = b.pool.Exec(ctx, `DELETE FROM auto_ecosystem_builds WHERE build_version < $1`, oldVersion)
	if err != nil {
		log.Printf("[AutoEco] Error cleaning old builds: %v", err)
	}
}

// markFailed 标记构建失败
func (b *Builder) markFailed(ctx context.Context, version int) {
	b.pool.Exec(ctx, `UPDATE auto_ecosystem_builds SET status='failed', completed_at=NOW() WHERE build_version=$1`, version)
}

// BenchmarkHubPenalty Task B: Hub Topic Penalty 基准测试
func (b *Builder) BenchmarkHubPenalty(ctx context.Context) (*HubPenaltyBenchmark, error) {
	// 加载 Topic Graph
	graph, normMap, err := b.loadTopicGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("load topic graph: %w", err)
	}
	mergedGraph := b.mergeNormalizedEdges(graph, normMap)

	// Without Penalty
	resultNoPenalty := louvain.DetectWithPenalty(mergedGraph, 3, false)
	countsNoPenalty := b.estimateRepoCounts(ctx, resultNoPenalty)
	metricsNoPenalty := b.computeMetrics(ctx, resultNoPenalty, countsNoPenalty)

	// With Penalty
	resultWithPenalty := louvain.DetectWithPenalty(mergedGraph, 3, true)
	countsWithPenalty := b.estimateRepoCounts(ctx, resultWithPenalty)
	metricsWithPenalty := b.computeMetrics(ctx, resultWithPenalty, countsWithPenalty)

	return &HubPenaltyBenchmark{
		WithoutPenalty: metricsNoPenalty,
		WithPenalty:    metricsWithPenalty,
	}, nil
}

// computeMetrics 计算社区指标
func (b *Builder) computeMetrics(ctx context.Context, result *louvain.Result, repoCounts map[int]int) PenaltyMetrics {
	largestCluster := 0
	for _, count := range repoCounts {
		if count > largestCluster {
			largestCluster = count
		}
	}

	var totalRepos int
	b.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories`).Scan(&totalRepos)

	assignedRepos := 0
	for _, count := range repoCounts {
		assignedRepos += count
	}

	coveragePct := 0.0
	if totalRepos > 0 {
		coveragePct = float64(assignedRepos) / float64(totalRepos) * 100
	}

	// Purity: 对比 rule_mapping
	var purityPct float64
	b.pool.QueryRow(ctx, `
		WITH auto_rule AS (
			SELECT aer.repo_id, ae.rule_mapping AS auto_eco
			FROM auto_ecosystem_repos aer
			JOIN auto_ecosystems ae ON aer.ecosystem_id = ae.id
			WHERE ae.rule_mapping IS NOT NULL
			LIMIT 1
		),
		rule_eco AS (
			SELECT repo_id, ecosystem AS rule_eco FROM ecosystem_map LIMIT 1
		)
		SELECT 0.0`).Scan(&purityPct)

	// Purity 需要实际构建才能准确计算，这里用简化估算
	// 基于 Rule 映射覆盖率估算
	ruleMapped := 0
	for _, comm := range result.Communities {
		rules := b.classifier.GetAllRules()
		topicWeight := make(map[string]float64)
		for _, t := range comm.Nodes {
			topicWeight[t] += 1
		}

		topTopics := make([]string, 0, 10)
		for t := range topicWeight {
			topTopics = append(topTopics, t)
			if len(topTopics) >= 10 {
				break
			}
		}

		for i := range rules {
			rule := &rules[i]
			ruleTopicsNorm := make(map[string]bool)
			for _, t := range rule.Topics {
				ruleTopicsNorm[normalizeTopic(t)] = true
			}
			overlap := 0
			for _, t := range topTopics {
				if ruleTopicsNorm[t] {
					overlap++
				}
			}
			if overlap >= 2 {
				ruleMapped++
				break
			}
		}
	}

	if len(result.Communities) > 0 {
		purityPct = float64(ruleMapped) / float64(len(result.Communities)) * 100
	}

	return PenaltyMetrics{
		ClusterCount:   len(result.Communities),
		LargestCluster: largestCluster,
		CoveragePct:    coveragePct,
		PurityPct:      purityPct,
	}
}

// GetReport 获取验证报告 (G68 Shadow Mode)
func (b *Builder) GetReport(ctx context.Context) (*Report, error) {
	var latestVersion int
	err := b.pool.QueryRow(ctx,
		`SELECT build_version FROM auto_ecosystem_builds WHERE status='completed' ORDER BY build_version DESC LIMIT 1`).Scan(&latestVersion)
	if err != nil {
		return nil, fmt.Errorf("no completed build found")
	}

	// Coverage
	var totalRepos, assignedRepos int
	b.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories`).Scan(&totalRepos)
	b.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT repo_id) FROM auto_ecosystem_repos WHERE build_version = $1`, latestVersion).Scan(&assignedRepos)

	coveragePct := 0.0
	if totalRepos > 0 {
		coveragePct = float64(assignedRepos) / float64(totalRepos) * 100
	}

	// Purity
	var purityPct float64
	b.pool.QueryRow(ctx, `
		WITH auto_rule AS (
			SELECT aer.repo_id, ae.rule_mapping AS auto_eco
			FROM auto_ecosystem_repos aer
			JOIN auto_ecosystems ae ON aer.ecosystem_id = ae.id
			WHERE aer.build_version = $1 AND ae.rule_mapping IS NOT NULL
		),
		rule_eco AS (
			SELECT repo_id, ecosystem AS rule_eco FROM ecosystem_map
		)
		SELECT COALESCE(
			ROUND(COUNT(CASE WHEN a.auto_eco = r.rule_eco THEN 1 END)::numeric / NULLIF(COUNT(*), 0) * 100, 2),
			0
		)
		FROM auto_rule a
		JOIN rule_eco r ON a.repo_id = r.repo_id`, latestVersion).Scan(&purityPct)

	// Emerging count
	var emergingCount int
	b.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM auto_ecosystems WHERE build_version = $1 AND is_emerging = TRUE`, latestVersion).Scan(&emergingCount)

	// Communities detail
	rows, err := b.pool.Query(ctx, `
		SELECT ae.id, ae.name, ae.rule_mapping, ae.is_emerging, ae.repo_count, ae.topic_count, ae.top_topics, ae.trend_score,
		       ae.parent_id, ae.level
		FROM auto_ecosystems ae
		WHERE ae.build_version = $1
		ORDER BY ae.repo_count DESC`, latestVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var communities []CommunityInfo
	largestCluster := 0
	totalClusterSize := 0
	level0Count := 0
	level1Count := 0
	superCount := 0

	for rows.Next() {
		var ci CommunityInfo
		var topTopicsJSON []byte
		var ruleMapping *string
		var parentID *int

		if err := rows.Scan(&ci.ID, &ci.Name, &ruleMapping, &ci.IsEmerging, &ci.RepoCount, &ci.TopicCount, &topTopicsJSON, &ci.TrendScore, &parentID, &ci.Level); err != nil {
			continue
		}
		ci.RuleMapping = ruleMapping
		ci.ParentID = parentID
		json.Unmarshal(topTopicsJSON, &ci.TopTopics)

		if ci.RepoCount > largestCluster {
			largestCluster = ci.RepoCount
		}
		totalClusterSize += ci.RepoCount

		if ci.Level == 0 {
			level0Count++
		} else if ci.Level == 1 {
			level1Count++
		}

		// Check if super (has children)
		var childCount int
		b.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM auto_ecosystems WHERE parent_id = $1 AND build_version = $2`,
			ci.ID, latestVersion).Scan(&childCount)
		if childCount > 0 {
			ci.IsSuper = true
			superCount++
		}

		// Top repos
		repoRows, err := b.pool.Query(ctx, `
			SELECT r.full_name FROM auto_ecosystem_repos aer
			JOIN repositories r ON aer.repo_id = r.id
			WHERE aer.ecosystem_id = $1 AND aer.build_version = $2
			ORDER BY r.stars DESC LIMIT 5`, ci.ID, latestVersion)
		if err == nil {
			for repoRows.Next() {
				var name string
				if repoRows.Scan(&name) == nil {
					ci.TopRepos = append(ci.TopRepos, name)
				}
			}
			repoRows.Close()
		}

		communities = append(communities, ci)
	}

	clusterCount := len(communities)
	avgClusterSize := 0.0
	if clusterCount > 0 {
		avgClusterSize = float64(totalClusterSize) / float64(clusterCount)
	}

	return &Report{
		BuildVersion:   latestVersion,
		CoveragePct:    coveragePct,
		PurityPct:      purityPct,
		EmergingCount:  emergingCount,
		ClusterCount:   clusterCount,
		LargestCluster: largestCluster,
		AvgClusterSize: avgClusterSize,
		Communities:    communities,
		HierarchyStats: HierarchyStats{
			Level0Count: level0Count,
			Level1Count: level1Count,
			SuperCount:  superCount,
		},
	}, nil
}

// 辅助函数

func countLevel(nodes []*ecoNode, level int) int {
	count := 0
	for _, n := range nodes {
		if n.level == level {
			count++
		}
	}
	return count
}

func countSuper(nodes []*ecoNode) int {
	count := 0
	for _, n := range nodes {
		if n.isSuper {
			count++
		}
	}
	return count
}

func parseVector(data []byte) ([]float64, error) {
	s := string(data)
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil, fmt.Errorf("empty vector")
	}
	parts := strings.Split(s, ",")
	vec := make([]float64, len(parts))
	for i, p := range parts {
		v, err := parseFloat(p)
		if err != nil {
			return nil, err
		}
		vec[i] = v
	}
	return vec, nil
}

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func float64Mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func float64Std(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	mean := float64Mean(vals)
	var sum float64
	for _, v := range vals {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)))
}
