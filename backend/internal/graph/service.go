package graph

import (
	"context"
	"fmt"

	"github.com/gitsense/gitsense/backend/internal/service"
)

// Service Graph 查询服务
type Service struct {
	graphStore *Store
	classifier *service.EcosystemClassifier
}

// NewService 创建 Graph Service
func NewService(graphStore *Store, classifier *service.EcosystemClassifier) *Service {
	return &Service{graphStore: graphStore, classifier: classifier}
}

// GetRepoGraph 获取 Repo 的图邻域
func (s *Service) GetRepoGraph(ctx context.Context, fullName string) (*RepoGraphResponse, error) {
	// 获取相似边
	edges, err := s.graphStore.GetRepoEdges(ctx, fullName, 20)
	if err != nil {
		return nil, fmt.Errorf("get repo edges: %w", err)
	}

	// 获取生态
	pool := s.graphStore.Pool()
	var topics []string
	var language string
	var eco string

	_ = pool.QueryRow(ctx, `
		SELECT topics, language FROM repositories WHERE full_name = $1
	`, fullName).Scan(&topics, &language)

	// 从 ecosystem_map 获取生态
	var repoID int64
	_ = pool.QueryRow(ctx, `SELECT id FROM repositories WHERE full_name = $1`, fullName).Scan(&repoID)
	if repoID > 0 {
		eco, _ = s.graphStore.GetRepoEcosystem(ctx, repoID)
	}

	// 构建节点列表
	nodes := []GraphNode{
		{ID: fullName, Type: NodeRepo, Label: fullName},
	}
	for _, e := range edges {
		peer := e.DstRepo
		if peer == fullName {
			peer = e.SrcRepo
		}
		nodes = append(nodes, GraphNode{ID: peer, Type: NodeRepo, Label: peer})
	}
	for _, t := range topics {
		nodes = append(nodes, GraphNode{ID: "topic:" + t, Type: NodeTopic, Label: t})
	}
	if language != "" {
		nodes = append(nodes, GraphNode{ID: "lang:" + language, Type: NodeLanguage, Label: language})
	}
	if eco != "" {
		nodes = append(nodes, GraphNode{ID: "eco:" + eco, Type: NodeEcosystem, Label: eco})
	}

	return &RepoGraphResponse{
		Repo:      fullName,
		SimilarTo: edges,
		Topics:    topics,
		Language:  language,
		Ecosystem: eco,
		Nodes:     nodes,
	}, nil
}

// GetEcosystemGraph 获取生态图
func (s *Service) GetEcosystemGraph(ctx context.Context, ecosystem string) (*EcosystemGraphResponse, error) {
	// 获取生态内 repos
	repos, err := s.graphStore.GetEcosystemRepos(ctx, ecosystem, 50)
	if err != nil {
		return nil, fmt.Errorf("get ecosystem repos: %w", err)
	}

	// 获取生态内 topics
	topics, err := s.graphStore.GetEcosystemTopics(ctx, ecosystem)
	if err != nil {
		topics = nil
	}

	// 获取 topic 边
	topicEdges, _ := s.graphStore.GetEcosystemTopicEdges(ctx, topics, 50)

	// 构建节点
	var nodes []GraphNode
	nodes = append(nodes, GraphNode{ID: "eco:" + ecosystem, Type: NodeEcosystem, Label: ecosystem})
	for _, r := range repos {
		nodes = append(nodes, GraphNode{ID: r, Type: NodeRepo, Label: r})
	}
	for _, t := range topics {
		nodes = append(nodes, GraphNode{ID: "topic:" + t, Type: NodeTopic, Label: t})
	}

	return &EcosystemGraphResponse{
		Ecosystem:    ecosystem,
		Repos:        repos,
		TopicCluster: topicEdges,
		Nodes:        nodes,
	}, nil
}

// FindPaths 2-hop 路径查询
func (s *Service) FindPaths(ctx context.Context, from, to string, maxHops, limit int) (*PathResponse, error) {
	if maxHops <= 0 {
		maxHops = 2
	}
	if maxHops > 2 {
		maxHops = 2
	}
	if limit <= 0 {
		limit = 50
	}

	paths, err := s.graphStore.FindPaths(ctx, from, to, maxHops, limit)
	if err != nil {
		return nil, fmt.Errorf("find paths: %w", err)
	}

	return &PathResponse{
		From:  from,
		To:    to,
		Paths: paths,
		Hops:  maxHops,
	}, nil
}

// GetGraphExplanation 为推荐生成 Graph 解释（实时计算，不存储）
func (s *Service) GetGraphExplanation(ctx context.Context, repoA, repoB string, repoATopics, repoBTopics []string) (*GraphExplanation, error) {
	explanation := &GraphExplanation{}

	// 1. 相似路径：查找 A → B 的 1-hop 或 2-hop 路径
	paths, err := s.graphStore.FindPaths(ctx, repoA, repoB, 2, 5)
	if err == nil && len(paths) > 0 {
		// 取最短路径
		path := paths[0]
		pathNodes := []string{repoA}
		for _, edge := range path {
			if edge.SrcRepo != pathNodes[len(pathNodes)-1] {
				pathNodes = append(pathNodes, edge.SrcRepo)
			}
			pathNodes = append(pathNodes, edge.DstRepo)
		}
		explanation.SimilarityPath = pathNodes
	}

	// 2. 共享生态
	ecoA, _ := s.graphStore.GetRepoEcosystem(ctx, 0) // 需要 repoID，简化处理
	_ = ecoA
	// 从 ecosystem_map 查询
	pool := s.graphStore.Pool()
	var sharedEco string
	_ = pool.QueryRow(ctx, `
		SELECT em1.ecosystem
		FROM ecosystem_map em1
		JOIN ecosystem_map em2 ON em1.ecosystem = em2.ecosystem
		JOIN repositories r1 ON r1.id = em1.repo_id AND r1.full_name = $1
		JOIN repositories r2 ON r2.id = em2.repo_id AND r2.full_name = $2
		LIMIT 1
	`, repoA, repoB).Scan(&sharedEco)
	explanation.SharedEcosystem = sharedEco

	// 3. Topic 桥接路径：通过 topic_edges 找 A 的 topic → B 的 topic 的连接
	topicBridge := s.findTopicBridge(ctx, repoATopics, repoBTopics)
	explanation.TopicBridge = topicBridge

	// 4. graph_score（实时计算，仅用于 explanation）
	graphScore := s.computeGraphScore(ctx, repoA, repoB, repoATopics, repoBTopics, sharedEco)
	explanation.GraphScore = graphScore

	return explanation, nil
}

// findTopicBridge 通过 topic_edges 找 topic 桥接路径
func (s *Service) findTopicBridge(ctx context.Context, topicsA, topicsB []string) []string {
	if len(topicsA) == 0 || len(topicsB) == 0 {
		return nil
	}

	setB := make(map[string]bool, len(topicsB))
	for _, t := range topicsB {
		setB[t] = true
	}

	// 查找 A 的 topic 通过 topic_edges 连接到 B 的 topic
	for _, tA := range topicsA {
		edges, err := s.graphStore.GetTopicEdges(ctx, tA, 10)
		if err != nil {
			continue
		}
		for _, edge := range edges {
			peer := edge.TopicB
			if peer == tA {
				peer = edge.TopicA
			}
			if setB[peer] {
				return []string{tA, peer}
			}
		}
	}

	return nil
}

// computeGraphScore 实时计算 graph_score（仅用于 explanation）
// graph_score = 0.5 * embedding_similarity + 0.3 * shared_ecosystem + 0.2 * topic_distance
func (s *Service) computeGraphScore(ctx context.Context, repoA, repoB string, topicsA, topicsB []string, sharedEco string) float64 {
	var score float64

	// embedding_similarity: 从 repo_edges 获取
	edges, err := s.graphStore.GetRepoEdges(ctx, repoA, 50)
	if err == nil {
		for _, e := range edges {
			peer := e.DstRepo
			if peer == repoA {
				peer = e.SrcRepo
			}
			if peer == repoB {
				score += 0.5 * e.WeightEmbedding
				break
			}
		}
	}

	// shared_ecosystem
	if sharedEco != "" {
		score += 0.3
	}

	// topic_distance: Jaccard 相似度
	if len(topicsA) > 0 && len(topicsB) > 0 {
		setA := make(map[string]bool, len(topicsA))
		for _, t := range topicsA {
			setA[t] = true
		}
		intersection := 0
		for _, t := range topicsB {
			if setA[t] {
				intersection++
			}
		}
		union := len(topicsA) + len(topicsB) - intersection
		if union > 0 {
			score += 0.2 * float64(intersection) / float64(union)
		}
	}

	return score
}
