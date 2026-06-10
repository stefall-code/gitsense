package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/gitsense/gitsense/backend/internal/service"
)

// BuilderWorker 构建 Graph 边的 Worker
// 定时全量 + 手动触发 + 时间窗口优化
type BuilderWorker struct {
	graphStore *Store
	repoStore  interface {
		GetAllWithEmbedding(ctx context.Context, since time.Time, limit int) ([]repoRow, error)
	}
	classifier *service.EcosystemClassifier
	interval   time.Duration
	stopCh     chan struct{}
}

// repoRow 最小 repo 数据（避免依赖 model 包）
type repoRow struct {
	ID          int64
	FullName    string
	Topics      []string
	Description string
	Language    string
	Stars       int
}

// NewBuilderWorker 创建 Graph Builder Worker
func NewBuilderWorker(graphStore *Store, classifier *service.EcosystemClassifier, interval time.Duration) *BuilderWorker {
	if interval == 0 {
		interval = 6 * time.Hour
	}
	return &BuilderWorker{
		graphStore: graphStore,
		classifier: classifier,
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动定时构建
func (w *BuilderWorker) Start(ctx context.Context) {
	log.Printf("[graph-worker] started (interval=%s)", w.interval)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[graph-worker] stopped")
			return
		case <-w.stopCh:
			log.Printf("[graph-worker] stopped")
			return
		case <-ticker.C:
			if _, err := w.BuildGraph(ctx, false); err != nil {
				log.Printf("[graph-worker] build error: %v", err)
			}
		}
	}
}

// Stop 停止 worker
func (w *BuilderWorker) Stop() {
	close(w.stopCh)
}

// BuildGraph 构建 Graph（全量或增量）
// fullRebuild=true: 清空所有边重新构建
// fullRebuild=false: 仅重建最近 7 天更新的 repo
func (w *BuilderWorker) BuildGraph(ctx context.Context, fullRebuild bool) (*BuildGraphResponse, error) {
	log.Printf("[graph-worker] building graph (full=%v)", fullRebuild)

	pool := w.graphStore.Pool()

	// Step 1: 获取所有有 embedding 的 repo
	since := time.Time{}
	if !fullRebuild {
		since = time.Now().AddDate(0, 0, -7) // 最近 7 天
	}

	var repos []repoRow
	query := `
		SELECT id, full_name, topics, description, language, stars
		FROM repositories
		WHERE embedding_status = 'done'
	`
	args := []interface{}{}
	if !fullRebuild && !since.IsZero() {
		query += ` AND updated_at >= $1`
		args = append(args, since)
	}
	query += ` ORDER BY stars DESC`

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query repos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r repoRow
		var topicsJSON []byte
		if err := rows.Scan(&r.ID, &r.FullName, &topicsJSON, &r.Description, &r.Language, &r.Stars); err != nil {
			continue
		}
		_ = json.Unmarshal(topicsJSON, &r.Topics)
		repos = append(repos, r)
	}

	log.Printf("[graph-worker] found %d repos to process", len(repos))

	if fullRebuild {
		_ = w.graphStore.TruncateRepoEdges(ctx)
		_ = w.graphStore.TruncateTopicEdges(ctx)
		_ = w.graphStore.TruncateEcosystemMap(ctx)
	}

	// Step 2: 构建 repo_edges (SIMILAR_TO)
	// 对每个 repo，用 pgvector 检索 top 10 相似，score > 0.75 才建边
	repoEdgeCount := 0
	for _, repo := range repos {
		edges, err := w.buildRepoEdges(ctx, repo, pool)
		if err != nil {
			log.Printf("[graph-worker] build repo edges for %s: %v", repo.FullName, err)
			continue
		}
		for _, edge := range edges {
			if err := w.graphStore.UpsertRepoEdge(ctx, edge); err != nil {
				log.Printf("[graph-worker] upsert repo edge: %v", err)
				continue
			}
			repoEdgeCount++
		}
	}

	// Step 3: 构建 topic_edges (CO_OCCUR)
	topicEdgeCount, err := w.buildTopicEdges(ctx, repos)
	if err != nil {
		log.Printf("[graph-worker] build topic edges: %v", err)
	}

	// Step 4: 构建 ecosystem_map
	ecoCount := 0
	for _, repo := range repos {
		eco := w.classifier.Classify(repo.Topics, repo.Description)
		if eco == "" {
			continue
		}
		entry := EcosystemMapEntry{
			RepoID:     repo.ID,
			Ecosystem:  eco,
			Confidence: 1.0,
		}
		if err := w.graphStore.UpsertEcosystemMap(ctx, entry); err != nil {
			continue
		}
		ecoCount++
	}

	log.Printf("[graph-worker] build complete: repo_edges=%d, topic_edges=%d, eco_mappings=%d",
		repoEdgeCount, topicEdgeCount, ecoCount)

	return &BuildGraphResponse{
		RepoEdges:   repoEdgeCount,
		TopicEdges:  topicEdgeCount,
		EcoMappings: ecoCount,
	}, nil
}

// buildRepoEdges 为单个 repo 构建 SIMILAR_TO 边
func (w *BuilderWorker) buildRepoEdges(ctx context.Context, repo repoRow, pool interface{}) ([]RepoEdge, error) {
	return w.buildRepoEdgesSimple(ctx, repo)
}

// buildRepoEdgesSimple 简化版 repo 边构建
func (w *BuilderWorker) buildRepoEdgesSimple(ctx context.Context, repo repoRow) ([]RepoEdge, error) {
	pool := w.graphStore.Pool()

	// 查询 top 20 相似 repo（Phase 9.6: 从 10 扩大到 20，配合 0.60 阈值）
	rows, err := pool.Query(ctx, `
		SELECT r.full_name,
		       1 - (r.embedding <=> (SELECT embedding FROM repositories WHERE full_name = $1)) AS emb_sim,
		       r.topics, r.language
		FROM repositories r
		WHERE r.full_name != $1 AND r.embedding_status = 'done'
		ORDER BY r.embedding <=> (SELECT embedding FROM repositories WHERE full_name = $1)
		LIMIT 20
	`, repo.FullName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []RepoEdge
	for rows.Next() {
		var dstName string
		var embSim float64
		var topicsJSON []byte
		var dstLang string

		if err := rows.Scan(&dstName, &embSim, &topicsJSON, &dstLang); err != nil {
			continue
		}

		// 计算 topic 相似度
		var dstTopics []string
		_ = json.Unmarshal(topicsJSON, &dstTopics)
		topicSim := jaccardSim(repo.Topics, dstTopics)

		// 融合评分
		score := 0.7*embSim + 0.3*topicSim

		// 阈值过滤：score >= 0.60 才建边（Phase 9.6 调整，原 0.75）
		if score < 0.60 {
			continue
		}

		edges = append(edges, RepoEdge{
			SrcRepo:         repo.FullName,
			DstRepo:         dstName,
			WeightEmbedding: embSim,
			WeightTopic:     topicSim,
			Score:           score,
		})
	}

	// 按 score 排序
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Score > edges[j].Score
	})

	return edges, nil
}

// buildTopicEdges 构建 Topic CO_OCCUR 边
// 同一 repo 内 topic pair +1 weight
// 双阈值：weight >= 2 → candidate, weight >= 3 → strong
func (w *BuilderWorker) buildTopicEdges(ctx context.Context, repos []repoRow) (int, error) {
	// 统计 topic pair 共现次数
	coOccur := make(map[[2]string]int)
	for _, repo := range repos {
		topics := repo.Topics
		if len(topics) < 2 {
			continue
		}
		// 排序保证 (a,b) 和 (b,a) 是同一条边
		sort.Strings(topics)
		for i := 0; i < len(topics); i++ {
			for j := i + 1; j < len(topics); j++ {
				if topics[i] != topics[j] {
					key := [2]string{topics[i], topics[j]}
					coOccur[key]++
				}
			}
		}
	}

	// 写入 topic_edges，双阈值过滤
	count := 0
	for pair, weight := range coOccur {
		if weight < 2 {
			continue // 低于 candidate 阈值，跳过
		}

		strength := "candidate"
		if weight >= 3 {
			strength = "strong"
		}

		edge := TopicEdge{
			TopicA:   pair[0],
			TopicB:   pair[1],
			Weight:   weight,
			Strength: strength,
		}

		if err := w.graphStore.UpsertTopicEdge(ctx, edge); err != nil {
			log.Printf("[graph-worker] upsert topic edge: %v", err)
			continue
		}
		count++
	}

	return count, nil
}

// jaccardSim Jaccard 相似度
func jaccardSim(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	setA := make(map[string]bool, len(a))
	for _, s := range a {
		setA[s] = true
	}
	intersection := 0
	for _, s := range b {
		if setA[s] {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
