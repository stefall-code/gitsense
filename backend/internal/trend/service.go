package trend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service Trend 查询服务
type Service struct {
	provider *CachedTrendProvider
	pool     *pgxpool.Pool
}

// NewService 创建 Trend Service
func NewService(provider *CachedTrendProvider) *Service {
	return &Service{
		provider: provider,
		pool:     provider.pool,
	}
}

// GetTopicTrends 获取 Topic 趋势列表
func (s *Service) GetTopicTrends(ctx context.Context, window TimeWindow, limit int) ([]TopicTrend, error) {
	if limit <= 0 {
		limit = 20
	}

	// 获取所有 topic（从 repositories 聚合）
	topics, err := s.getTopTopics(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("get top topics: %w", err)
	}

	// 计算每个 topic 的趋势
	var trends []TopicTrend
	for _, topic := range topics {
		trend := s.provider.computeTopicTrend(ctx, topic, window)
		if trend != nil {
			trends = append(trends, *trend)
		}
	}

	// 按 trend_score 降序排序
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].TrendScore > trends[j].TrendScore
	})

	if len(trends) > limit {
		trends = trends[:limit]
	}

	return trends, nil
}

// GetEcosystemTrends 获取生态趋势列表
func (s *Service) GetEcosystemTrends(ctx context.Context, window TimeWindow, limit int) ([]EcosystemTrend, error) {
	if limit <= 0 {
		limit = 20
	}

	// 获取所有 ecosystem
	ecosystems, err := s.getEcosystems(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ecosystems: %w", err)
	}

	var trends []EcosystemTrend
	for _, eco := range ecosystems {
		trend := s.provider.computeEcosystemTrend(ctx, eco, window)
		if trend != nil {
			trends = append(trends, *trend)
		}
	}

	sort.Slice(trends, func(i, j int) bool {
		return trends[i].TrendScore > trends[j].TrendScore
	})

	if len(trends) > limit {
		trends = trends[:limit]
	}

	return trends, nil
}

// GetOverview 获取趋势总览
func (s *Service) GetOverview(ctx context.Context, window TimeWindow) (*TrendOverview, error) {
	topicTrends, err := s.GetTopicTrends(ctx, window, 10)
	if err != nil {
		log.Printf("[trend] get topic trends for overview: %v", err)
		topicTrends = nil
	}

	ecoTrends, err := s.GetEcosystemTrends(ctx, window, 10)
	if err != nil {
		log.Printf("[trend] get ecosystem trends for overview: %v", err)
		ecoTrends = nil
	}

	// 只保留 rising 的
	var risingTopics []TopicTrend
	for _, t := range topicTrends {
		if t.Status == TrendRising {
			risingTopics = append(risingTopics, t)
		}
	}

	var risingEcosystems []EcosystemTrend
	for _, t := range ecoTrends {
		if t.Status == TrendRising {
			risingEcosystems = append(risingEcosystems, t)
		}
	}

	return &TrendOverview{
		TopRisingTopics:     risingTopics,
		TopRisingEcosystems: risingEcosystems,
		Window:              window,
		UpdatedAt:           time.Now(),
	}, nil
}

// GetTopicTrend 获取单个 topic 的趋势
func (s *Service) GetTopicTrend(ctx context.Context, topic string, window TimeWindow) (*TopicTrend, error) {
	trend := s.provider.computeTopicTrend(ctx, topic, window)
	if trend == nil {
		return nil, fmt.Errorf("topic trend not available: %s", topic)
	}
	return trend, nil
}

// GetEcosystemTrend 获取单个 ecosystem 的趋势
func (s *Service) GetEcosystemTrend(ctx context.Context, ecosystem string, window TimeWindow) (*EcosystemTrend, error) {
	trend := s.provider.computeEcosystemTrend(ctx, ecosystem, window)
	if trend == nil {
		return nil, fmt.Errorf("ecosystem trend not available: %s", ecosystem)
	}
	return trend, nil
}

// GetTopicTrendScore 获取单个 topic 的趋势分数（便捷方法）
func (s *Service) GetTopicTrendScore(ctx context.Context, topic string, window TimeWindow) float64 {
	return s.provider.GetTopicTrendScore(ctx, topic, window)
}

// GetEcosystemTrendScore 获取单个 ecosystem 的趋势分数（便捷方法）
func (s *Service) GetEcosystemTrendScore(ctx context.Context, ecosystem string, window TimeWindow) float64 {
	return s.provider.GetEcosystemTrendScore(ctx, ecosystem, window)
}

// --- 辅助查询 ---

// getTopTopics 获取出现频率最高的 topics
func (s *Service) getTopTopics(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT topic, COUNT(*) as cnt
		FROM (
			SELECT jsonb_array_elements_text(topics) AS topic
			FROM repositories
			WHERE topics != '[]'::jsonb
		) sub
		GROUP BY topic
		ORDER BY cnt DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topic string
		var cnt int
		if err := rows.Scan(&topic, &cnt); err != nil {
			continue
		}
		topics = append(topics, topic)
	}
	return topics, rows.Err()
}

// getEcosystems 获取所有 ecosystem
func (s *Service) getEcosystems(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ecosystem FROM ecosystem_map ORDER BY ecosystem
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ecosystems []string
	for rows.Next() {
		var eco string
		if err := rows.Scan(&eco); err != nil {
			continue
		}
		ecosystems = append(ecosystems, eco)
	}
	return ecosystems, rows.Err()
}

// RefreshCache 刷新 Redis 缓存（供 Worker 调用）
func (s *Service) RefreshCache(ctx context.Context, window TimeWindow) error {
	topics, err := s.getTopTopics(ctx, 100)
	if err != nil {
		return fmt.Errorf("get top topics: %w", err)
	}

	for _, topic := range topics {
		trend := s.provider.computeTopicTrend(ctx, topic, window)
		if trend != nil && s.provider.redis != nil {
			key := fmt.Sprintf("trend:topic:%s:%s", window, topic)
			data, _ := json.Marshal(trend)
			_ = s.provider.redis.Set(ctx, key, string(data), s.provider.ttl)
		}
	}

	ecosystems, err := s.getEcosystems(ctx)
	if err != nil {
		return fmt.Errorf("get ecosystems: %w", err)
	}

	for _, eco := range ecosystems {
		trend := s.provider.computeEcosystemTrend(ctx, eco, window)
		if trend != nil && s.provider.redis != nil {
			key := fmt.Sprintf("trend:ecosystem:%s:%s", window, eco)
			data, _ := json.Marshal(trend)
			_ = s.provider.redis.Set(ctx, key, string(data), s.provider.ttl)
		}
	}

	log.Printf("[trend] cache refreshed: %d topics, %d ecosystems", len(topics), len(ecosystems))
	return nil
}
