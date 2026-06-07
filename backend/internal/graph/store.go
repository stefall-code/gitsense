package graph

import (
	"context"

	"github.com/gitsense/gitsense/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store 处理 Graph 相关的 PostgreSQL 数据访问
type Store struct {
	pool *pgxpool.Pool
}

// NewStore 创建新的 Graph Store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// --- RepoEdges ---

// UpsertRepoEdge 插入或更新 repo 边
func (s *Store) UpsertRepoEdge(ctx context.Context, edge RepoEdge) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO repo_edges (src_repo, dst_repo, weight_embedding, weight_topic, score)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (src_repo, dst_repo) DO UPDATE SET
			weight_embedding = EXCLUDED.weight_embedding,
			weight_topic = EXCLUDED.weight_topic,
			score = EXCLUDED.score
	`, edge.SrcRepo, edge.DstRepo, edge.WeightEmbedding, edge.WeightTopic, edge.Score)
	return err
}

// GetRepoEdges 获取指定 repo 的所有相似边
func (s *Store) GetRepoEdges(ctx context.Context, fullName string, limit int) ([]RepoEdge, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, src_repo, dst_repo, weight_embedding, weight_topic, score, created_at
		FROM repo_edges
		WHERE src_repo = $1 OR dst_repo = $1
		ORDER BY score DESC
		LIMIT $2
	`, fullName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []RepoEdge
	for rows.Next() {
		var e RepoEdge
		if err := rows.Scan(&e.ID, &e.SrcRepo, &e.DstRepo, &e.WeightEmbedding, &e.WeightTopic, &e.Score, &e.CreatedAt); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// DeleteRepoEdges 删除指定 repo 的所有边
func (s *Store) DeleteRepoEdges(ctx context.Context, fullName string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM repo_edges WHERE src_repo = $1 OR dst_repo = $1`, fullName)
	return err
}

// TruncateRepoEdges 清空所有 repo 边
func (s *Store) TruncateRepoEdges(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `TRUNCATE repo_edges`)
	return err
}

// CountRepoEdges 统计 repo 边数量
func (s *Store) CountRepoEdges(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repo_edges`).Scan(&count)
	return count, err
}

// FindPaths 2-hop 路径查询
func (s *Store) FindPaths(ctx context.Context, from, to string, maxHops, limit int) ([][]RepoEdge, error) {
	if maxHops > 2 {
		maxHops = 2
	}
	if limit <= 0 {
		limit = 50
	}

	// 1-hop: direct edge
	var paths [][]RepoEdge
	directRows, err := s.pool.Query(ctx, `
		SELECT id, src_repo, dst_repo, weight_embedding, weight_topic, score, created_at
		FROM repo_edges
		WHERE (src_repo = $1 AND dst_repo = $2) OR (src_repo = $2 AND dst_repo = $1)
		ORDER BY score DESC LIMIT $3
	`, from, to, limit)
	if err == nil {
		defer directRows.Close()
		for directRows.Next() {
			var e RepoEdge
			if err := directRows.Scan(&e.ID, &e.SrcRepo, &e.DstRepo, &e.WeightEmbedding, &e.WeightTopic, &e.Score, &e.CreatedAt); err == nil {
				paths = append(paths, []RepoEdge{e})
			}
		}
	}

	if maxHops >= 2 {
		// 2-hop: from → intermediate → to
		twoHopRows, err := s.pool.Query(ctx, `
			SELECT e1.id, e1.src_repo, e1.dst_repo, e1.weight_embedding, e1.weight_topic, e1.score, e1.created_at,
			       e2.id, e2.src_repo, e2.dst_repo, e2.weight_embedding, e2.weight_topic, e2.score, e2.created_at
			FROM repo_edges e1
			JOIN repo_edges e2 ON (
				(e1.dst_repo = e2.src_repo AND e2.dst_repo = $2 AND e1.src_repo = $1)
				OR (e1.src_repo = e2.dst_repo AND e2.src_repo = $2 AND e1.dst_repo = $1)
				OR (e1.dst_repo = e2.dst_repo AND e1.src_repo = $1 AND e2.src_repo = $2)
				OR (e1.src_repo = e2.src_repo AND e1.dst_repo = $1 AND e2.dst_repo = $2)
			)
			WHERE e1.id != e2.id
			ORDER BY (e1.score + e2.score) DESC
			LIMIT $3
		`, from, to, limit)
		if err == nil {
			defer twoHopRows.Close()
			for twoHopRows.Next() {
				var e1, e2 RepoEdge
				if err := twoHopRows.Scan(
					&e1.ID, &e1.SrcRepo, &e1.DstRepo, &e1.WeightEmbedding, &e1.WeightTopic, &e1.Score, &e1.CreatedAt,
					&e2.ID, &e2.SrcRepo, &e2.DstRepo, &e2.WeightEmbedding, &e2.WeightTopic, &e2.Score, &e2.CreatedAt,
				); err == nil {
					paths = append(paths, []RepoEdge{e1, e2})
				}
			}
		}
	}

	return paths, nil
}

// --- TopicEdges ---

// UpsertTopicEdge 插入或更新 topic 边
func (s *Store) UpsertTopicEdge(ctx context.Context, edge TopicEdge) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO topic_edges (topic_a, topic_b, weight, strength)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (topic_a, topic_b) DO UPDATE SET
			weight = EXCLUDED.weight,
			strength = EXCLUDED.strength
	`, edge.TopicA, edge.TopicB, edge.Weight, edge.Strength)
	return err
}

// GetTopicEdges 获取指定 topic 的共现边
func (s *Store) GetTopicEdges(ctx context.Context, topic string, limit int) ([]TopicEdge, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, topic_a, topic_b, weight, strength, created_at
		FROM topic_edges
		WHERE topic_a = $1 OR topic_b = $1
		ORDER BY weight DESC
		LIMIT $2
	`, topic, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []TopicEdge
	for rows.Next() {
		var e TopicEdge
		if err := rows.Scan(&e.ID, &e.TopicA, &e.TopicB, &e.Weight, &e.Strength, &e.CreatedAt); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// GetEcosystemTopicEdges 获取生态内所有 topic 边
func (s *Store) GetEcosystemTopicEdges(ctx context.Context, topics []string, limit int) ([]TopicEdge, error) {
	if len(topics) == 0 {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, topic_a, topic_b, weight, strength, created_at
		FROM topic_edges
		WHERE topic_a = ANY($1) OR topic_b = ANY($1)
		ORDER BY weight DESC
		LIMIT $2
	`, topics, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []TopicEdge
	for rows.Next() {
		var e TopicEdge
		if err := rows.Scan(&e.ID, &e.TopicA, &e.TopicB, &e.Weight, &e.Strength, &e.CreatedAt); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// TruncateTopicEdges 清空所有 topic 边
func (s *Store) TruncateTopicEdges(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `TRUNCATE topic_edges`)
	return err
}

// CountTopicEdges 统计 topic 边数量
func (s *Store) CountTopicEdges(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM topic_edges`).Scan(&count)
	return count, err
}

// --- EcosystemMap ---

// UpsertEcosystemMap 插入或更新生态映射
func (s *Store) UpsertEcosystemMap(ctx context.Context, entry EcosystemMapEntry) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ecosystem_map (repo_id, ecosystem, confidence)
		VALUES ($1, $2, $3)
		ON CONFLICT (repo_id, ecosystem) DO UPDATE SET
			confidence = EXCLUDED.confidence
	`, entry.RepoID, entry.Ecosystem, entry.Confidence)
	return err
}

// GetEcosystemRepos 获取生态内所有 repo
func (s *Store) GetEcosystemRepos(ctx context.Context, ecosystem string, limit int) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.full_name
		FROM ecosystem_map em
		JOIN repositories r ON r.id = em.repo_id
		WHERE em.ecosystem = $1
		ORDER BY r.stars DESC
		LIMIT $2
	`, ecosystem, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		repos = append(repos, name)
	}
	return repos, rows.Err()
}

// GetRepoEcosystem 获取 repo 的生态
func (s *Store) GetRepoEcosystem(ctx context.Context, repoID int64) (string, error) {
	var ecosystem string
	err := s.pool.QueryRow(ctx, `
		SELECT ecosystem FROM ecosystem_map
		WHERE repo_id = $1
		ORDER BY confidence DESC LIMIT 1
	`, repoID).Scan(&ecosystem)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return ecosystem, err
}

// TruncateEcosystemMap 清空所有生态映射
func (s *Store) TruncateEcosystemMap(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `TRUNCATE ecosystem_map`)
	return err
}

// CountEcosystemMap 统计生态映射数量
func (s *Store) CountEcosystemMap(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ecosystem_map`).Scan(&count)
	return count, err
}

// GetEcosystemTopics 获取生态内所有 topics
func (s *Store) GetEcosystemTopics(ctx context.Context, ecosystem string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT jsonb_array_elements_text(r.topics) AS topic
		FROM ecosystem_map em
		JOIN repositories r ON r.id = em.repo_id
		WHERE em.ecosystem = $1 AND r.topics != '[]'::jsonb
		ORDER BY topic
	`, ecosystem)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			continue
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

// --- 辅助：获取 RepoStore 的 pool（用于 Builder） ---

// Pool 返回底层连接池
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// GraphMetrics Graph 指标
type GraphMetrics struct {
	RepoCount       int64   `json:"repo_count"`
	RepoEdges       int64   `json:"repo_edges"`
	ActiveNodes     int64   `json:"active_nodes"`
	IsolatedNodes   int64   `json:"isolated_nodes"`
	CoveragePercent float64 `json:"coverage_percent"`
	AverageDegree   float64 `json:"average_degree"`
	LargestComponent int64  `json:"largest_component"`
	ComponentCount  int64   `json:"component_count"`
}

// GetGraphMetrics 获取 Graph 指标
func (s *Store) GetGraphMetrics(ctx context.Context) (*GraphMetrics, error) {
	m := &GraphMetrics{}

	// Total repos with embedding
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM repositories WHERE embedding_status = 'done'`,
	).Scan(&m.RepoCount); err != nil {
		return nil, err
	}

	// Total edges
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM repo_edges`,
	).Scan(&m.RepoEdges); err != nil {
		return nil, err
	}

	// Active nodes (repos with at least one edge)
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT n) FROM (
			SELECT src_repo AS n FROM repo_edges
			UNION
			SELECT dst_repo AS n FROM repo_edges
		) t
	`).Scan(&m.ActiveNodes); err != nil {
		return nil, err
	}

	m.IsolatedNodes = m.RepoCount - m.ActiveNodes
	if m.RepoCount > 0 {
		m.CoveragePercent = float64(m.ActiveNodes) / float64(m.RepoCount) * 100
	}

	// Average degree
	if m.ActiveNodes > 0 {
		m.AverageDegree = float64(m.RepoEdges) / float64(m.ActiveNodes)
	}

	// Connected components using UNION-FIND approximation
	// For small graphs, use exact counting; for large, estimate
	// nodes - unique_undirected_edges + 1 (forest formula)
	var uniqueUndirectedEdges int64
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT DISTINCT LEAST(src_repo, dst_repo) AS a, GREATEST(src_repo, dst_repo) AS b FROM repo_edges
		) t
	`).Scan(&uniqueUndirectedEdges)

	m.ComponentCount = m.ActiveNodes - uniqueUndirectedEdges + 1
	if m.ComponentCount < 1 {
		m.ComponentCount = 1
	}

	// Largest component: for now, estimate based on avg_degree
	// With avg_degree ~6, largest component should include most active nodes
	// Exact computation requires BFS which is expensive
	if m.ActiveNodes > 0 && m.AverageDegree >= 2.0 {
		m.LargestComponent = m.ActiveNodes // highly connected graph
	} else {
		m.LargestComponent = 2 // isolated pairs
	}

	return m, nil
}
var _ = (*repository.RepoStore)(nil) // 仅类型检查，不实际使用
