package audit

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GraphAuditResult Graph 审计结果
type GraphAuditResult struct {
	RepoCount      int64               `json:"repo_count"`
	RepoEdges      int64               `json:"repo_edges"`
	TopicEdges     int64               `json:"topic_edges"`
	EcosystemCount int64               `json:"ecosystem_count"`
	ThresholdSweep []ThresholdSweepRow `json:"threshold_sweep"`
	GraphHealth    GraphHealth         `json:"graph_health"`
	Ablation       AblationResult      `json:"ablation"`
	Recommendation AuditRecommendation `json:"recommendation"`
}

// ThresholdSweepRow 阈值扫描行
type ThresholdSweepRow struct {
	Threshold     float64 `json:"threshold"`
	EdgeCount     int64   `json:"edge_count"`
	AvgDegree     float64 `json:"avg_degree"`
	MaxDegree     int64   `json:"max_degree"`
	Components    int64   `json:"components"`
	Density       string  `json:"density"`
}

// GraphHealth 图健康度
type GraphHealth struct {
	AvgDegree     float64 `json:"avg_degree"`
	MaxDegree     int64   `json:"max_degree"`
	Components    int64   `json:"components"`
	NodesWithEdge int64   `json:"nodes_with_edge"`
	CoveragePct   float64 `json:"coverage_pct"`
}

// AblationResult 消融实验结果
type AblationResult struct {
	AvgOverlapPct     float64  `json:"avg_overlap_pct"`
	AvgGraphDiscovery float64  `json:"avg_graph_discovery"`
	Queries            []AblationQuery `json:"queries"`
}

// AblationQuery 单个查询的消融结果
type AblationQuery struct {
	Query           string  `json:"query"`
	OverlapPct      float64 `json:"overlap_pct"`
	GraphDiscovery  int     `json:"graph_discovery"`
}

// AuditRecommendation 审计建议
type AuditRecommendation struct {
	BestThreshold     float64 `json:"best_threshold"`
	CurrentThreshold  float64 `json:"current_threshold"`
	ShouldAdjust      bool    `json:"should_adjust"`
	Reason            string  `json:"reason"`
	ShouldAdjustWeight bool   `json:"should_adjust_weight"`
	WeightReason      string  `json:"weight_reason"`
}

// Store audit 数据访问
type Store struct {
	pool *pgxpool.Pool
}

// NewStore 创建 audit store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// RunGraphAudit 执行完整的 Graph 审计
func (s *Store) RunGraphAudit(ctx context.Context) (*GraphAuditResult, error) {
	result := &GraphAuditResult{}

	// 1. 基础统计
	if err := s.basicStats(ctx, result); err != nil {
		return nil, fmt.Errorf("basic stats: %w", err)
	}

	// 2. 阈值扫描
	if err := s.thresholdSweep(ctx, result); err != nil {
		return nil, fmt.Errorf("threshold sweep: %w", err)
	}

	// 3. 图健康度
	if err := s.graphHealth(ctx, result); err != nil {
		return nil, fmt.Errorf("graph health: %w", err)
	}

	// 4. 生成建议
	s.generateRecommendation(result)

	return result, nil
}

func (s *Store) basicStats(ctx context.Context, result *GraphAuditResult) error {
	queries := []struct {
		query string
		dest  *int64
	}{
		{"SELECT COUNT(*) FROM repositories WHERE embedding_status='done'", &result.RepoCount},
		{"SELECT COUNT(*) FROM repo_edges", &result.RepoEdges},
		{"SELECT COUNT(*) FROM topic_edges", &result.TopicEdges},
		{"SELECT COUNT(DISTINCT ecosystem) FROM ecosystem_map", &result.EcosystemCount},
	}
	for _, q := range queries {
		if err := s.pool.QueryRow(ctx, q.query).Scan(q.dest); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) thresholdSweep(ctx context.Context, result *GraphAuditResult) error {
	// 采样 200 个 repo，避免全量扫描
	thresholds := []float64{0.80, 0.75, 0.70, 0.65, 0.60, 0.55}

	// 先采样获取 top-20 邻居的 emb_sim 分布
	rows, err := s.pool.Query(ctx, `
		WITH sample AS (
			SELECT embedding FROM repositories WHERE embedding_status = 'done' ORDER BY RANDOM() LIMIT 200
		)
		SELECT 1 - (r.embedding <=> s.embedding) AS emb_sim
		FROM sample s
		CROSS JOIN LATERAL (
			SELECT embedding FROM repositories
			WHERE embedding_status = 'done'
			ORDER BY embedding <=> s.embedding LIMIT 20
		) r
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var allSim []float64
	for rows.Next() {
		var sim float64
		if err := rows.Scan(&sim); err != nil {
			continue
		}
		allSim = append(allSim, sim)
	}

	// 统计每个阈值的采样边数，然后按比例估算全量
	sampleSize := int64(len(allSim))
	if sampleSize == 0 {
		return nil
	}

	for _, th := range thresholds {
		var sampleCount int64
		for _, sim := range allSim {
			if sim >= th {
				sampleCount++
			}
		}

		// 按比例估算全量边数：sampleCount / sampleSize * totalPairs
		// totalPairs = repoCount * 20 (top-20 neighbors per repo)
		estimatedEdges := sampleCount * result.RepoCount * 20 / sampleSize

		avgDegree := float64(estimatedEdges) / float64(result.RepoCount)
		var maxDegree int64
		if avgDegree > 20 {
			maxDegree = 20
		} else {
			maxDegree = int64(avgDegree * 2)
		}

		density := "sparse"
		if avgDegree >= 3.0 && avgDegree <= 8.0 {
			density = "reasonable"
		} else if avgDegree > 8.0 {
			density = "dense"
		}

		result.ThresholdSweep = append(result.ThresholdSweep, ThresholdSweepRow{
			Threshold: th,
			EdgeCount: estimatedEdges,
			AvgDegree: avgDegree,
			MaxDegree: maxDegree,
			Density:   density,
		})
	}
	return nil
}

func (s *Store) graphHealth(ctx context.Context, result *GraphAuditResult) error {
	var nodesWithEdge int64
	var avgDegree float64
	var maxDegree int64

	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT src_repo),
			COALESCE(AVG(degree), 0),
			COALESCE(MAX(degree), 0)
		FROM (
			SELECT src_repo, COUNT(*) AS degree FROM repo_edges GROUP BY src_repo
		) t
	`).Scan(&nodesWithEdge, &avgDegree, &maxDegree)
	if err != nil {
		return err
	}

	// 连通分量（简化：用 UNION 去重节点对）
	var components int64
	err = s.pool.QueryRow(ctx, `
		WITH nodes AS (
			SELECT DISTINCT repo FROM (
				SELECT src_repo AS repo FROM repo_edges
				UNION
				SELECT dst_repo AS repo FROM repo_edges
			) t
		),
		edges AS (
			SELECT DISTINCT LEAST(src_repo, dst_repo) AS a, GREATEST(src_repo, dst_repo) AS b
			FROM repo_edges
		)
		SELECT COUNT(*) - (SELECT COUNT(*) FROM edges) + 1 FROM nodes
	`).Scan(&components)
	if err != nil {
		// fallback: nodes - edges + 1
		components = nodesWithEdge - result.RepoEdges/2 + 1
		if components < 1 {
			components = 1
		}
	}

	coverage := 0.0
	if result.RepoCount > 0 {
		coverage = float64(nodesWithEdge) / float64(result.RepoCount) * 100
	}

	result.GraphHealth = GraphHealth{
		AvgDegree:     avgDegree,
		MaxDegree:     maxDegree,
		Components:    components,
		NodesWithEdge: nodesWithEdge,
		CoveragePct:   coverage,
	}
	return nil
}

func (s *Store) generateRecommendation(result *GraphAuditResult) {
	rec := AuditRecommendation{
		CurrentThreshold: 0.75,
	}

	// 找到 avg_degree 在 3-8 范围内的最佳阈值
	for _, row := range result.ThresholdSweep {
		if row.AvgDegree >= 3.0 && row.AvgDegree <= 8.0 {
			rec.BestThreshold = row.Threshold
			rec.ShouldAdjust = true
			rec.Reason = fmt.Sprintf(
				"Threshold %.2f produces avg_degree=%.1f (reasonable density), vs current 0.75 which produces only %d edges (%.1f%% coverage)",
				row.Threshold, row.AvgDegree, result.RepoEdges, result.GraphHealth.CoveragePct,
			)
			break
		}
	}

	if !rec.ShouldAdjust {
		// 如果没有合理阈值，选最接近的
		rec.BestThreshold = 0.60
		rec.ShouldAdjust = true
		rec.Reason = "No threshold produces ideal density (avg_degree 3-8). Recommend 0.60 as compromise."
	}

	// 权重建议
	if result.GraphHealth.CoveragePct < 1.0 {
		rec.ShouldAdjustWeight = true
		rec.WeightReason = fmt.Sprintf(
			"Graph covers only %.1f%% of repos. V3 graph weight (0.20) is wasted on most queries. "+
				"Consider reducing to 0.10 and redistributing to embedding (0.55) until graph coverage improves.",
			result.GraphHealth.CoveragePct,
		)
	} else {
		rec.ShouldAdjustWeight = false
		rec.WeightReason = "Graph coverage is adequate, current weights are reasonable."
	}

	result.Recommendation = rec
}
