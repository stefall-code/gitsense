package trend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/gitsense/gitsense/backend/internal/cache"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RedisClient 最小 Redis 客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) (string, cache.CacheStatus, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}

// CachedTrendProvider Redis 缓存 + DB fallback 的趋势数据提供者
type CachedTrendProvider struct {
	redis RedisClient
	pool  *pgxpool.Pool
	ttl   time.Duration
}

// NewCachedTrendProvider 创建缓存趋势提供者
func NewCachedTrendProvider(redis RedisClient, pool *pgxpool.Pool, ttl time.Duration) *CachedTrendProvider {
	if ttl == 0 {
		ttl = 12 * time.Hour
	}
	return &CachedTrendProvider{redis: redis, pool: pool, ttl: ttl}
}

// GetTopicTrendScore 获取 topic 趋势分数（Redis 优先，miss 时 fallback 实时计算）
func (p *CachedTrendProvider) GetTopicTrendScore(ctx context.Context, topic string, window TimeWindow) float64 {
	key := fmt.Sprintf("trend:topic:%s:%s", window, topic)

	if p.redis != nil {
		val, _, err := p.redis.Get(ctx, key)
		if err == nil && val != "" {
			var t TopicTrend
			if err := json.Unmarshal([]byte(val), &t); err == nil {
				return t.TrendScore
			}
		}
	}

	trend := p.computeTopicTrend(ctx, topic, window)
	if trend != nil && p.redis != nil {
		data, _ := json.Marshal(trend)
		_ = p.redis.Set(ctx, key, string(data), p.ttl)
	}
	if trend != nil {
		return trend.TrendScore
	}
	return 0
}

// GetEcosystemTrendScore 获取 ecosystem 趋势分数
func (p *CachedTrendProvider) GetEcosystemTrendScore(ctx context.Context, ecosystem string, window TimeWindow) float64 {
	key := fmt.Sprintf("trend:ecosystem:%s:%s", window, ecosystem)

	if p.redis != nil {
		val, _, err := p.redis.Get(ctx, key)
		if err == nil && val != "" {
			var t EcosystemTrend
			if err := json.Unmarshal([]byte(val), &t); err == nil {
				return t.TrendScore
			}
		}
	}

	trend := p.computeEcosystemTrend(ctx, ecosystem, window)
	if trend != nil && p.redis != nil {
		data, _ := json.Marshal(trend)
		_ = p.redis.Set(ctx, key, string(data), p.ttl)
	}
	if trend != nil {
		return trend.TrendScore
	}
	return 0
}

// computeTopicTrend 实时计算 topic 趋势
func (p *CachedTrendProvider) computeTopicTrend(ctx context.Context, topic string, window TimeWindow) *TopicTrend {
	windowDur := WindowToDuration(window)
	now := time.Now()
	since := now.Add(-windowDur)
	prevSince := since.Add(-windowDur)

	count7d := p.countReposWithTopic(ctx, topic, since, now)
	countPrev7d := p.countReposWithTopic(ctx, topic, prevSince, since)

	growthRate := float64(count7d+1) / float64(countPrev7d+1)
	trendScore := ComputeTrendScore(growthRate)

	return &TopicTrend{
		Topic:       topic,
		GrowthRate:  growthRate,
		TrendScore:  trendScore,
		Status:      ClassifyTrend(trendScore),
		Count7d:     count7d,
		CountPrev7d: countPrev7d,
		Window:      window,
		UpdatedAt:   now,
	}
}

// computeEcosystemTrend 实时计算 ecosystem 趋势
func (p *CachedTrendProvider) computeEcosystemTrend(ctx context.Context, ecosystem string, window TimeWindow) *EcosystemTrend {
	windowDur := WindowToDuration(window)
	now := time.Now()
	since := now.Add(-windowDur)
	prevSince := since.Add(-windowDur)

	count7d := p.countReposInEcosystem(ctx, ecosystem, since, now)
	countPrev7d := p.countReposInEcosystem(ctx, ecosystem, prevSince, since)

	growthRate := float64(count7d+1) / float64(countPrev7d+1)
	trendScore := ComputeTrendScore(growthRate)

	return &EcosystemTrend{
		Ecosystem:   ecosystem,
		GrowthRate:  growthRate,
		TrendScore:  trendScore,
		Status:      ClassifyTrend(trendScore),
		Count7d:     count7d,
		CountPrev7d: countPrev7d,
		Window:      window,
		UpdatedAt:   now,
	}
}

// countReposWithTopic 统计指定时间窗口内包含某 topic 的 repo 数量
func (p *CachedTrendProvider) countReposWithTopic(ctx context.Context, topic string, since, until time.Time) int {
	var count int
	err := p.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM repositories
		WHERE topics @> $1::jsonb
		  AND last_activity_at >= $2
		  AND last_activity_at < $3
	`, fmt.Sprintf(`["%s"]`, topic), since, until).Scan(&count)
	if err != nil {
		log.Printf("[trend] count repos with topic %s: %v", topic, err)
		return 0
	}
	return count
}

// countReposInEcosystem 统计指定时间窗口内属于某 ecosystem 的 repo 数量
func (p *CachedTrendProvider) countReposInEcosystem(ctx context.Context, ecosystem string, since, until time.Time) int {
	var count int
	err := p.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ecosystem_map em
		JOIN repositories r ON r.id = em.repo_id
		WHERE em.ecosystem = $1
		  AND r.last_activity_at >= $2
		  AND r.last_activity_at < $3
	`, ecosystem, since, until).Scan(&count)
	if err != nil {
		log.Printf("[trend] count repos in ecosystem %s: %v", ecosystem, err)
		return 0
	}
	return count
}

// --- 核心计算函数（公开，供 Worker 和 V3 Strategy 使用）---

// ComputeTrendScore 计算趋势分数
// growth_rate = (count_7d + 1) / (count_prev_7d + 1)
// trend_score = tanh(log(growth_rate))，输出 [-1, 1]
func ComputeTrendScore(growthRate float64) float64 {
	if growthRate <= 0 {
		return -1
	}
	return math.Tanh(math.Log(growthRate))
}

// ClassifyTrend 根据 trend_score 分类
// > 0.2 = rising, [-0.2, 0.2] = stable, < -0.2 = declining
func ClassifyTrend(score float64) TrendStatus {
	if score > 0.2 {
		return TrendRising
	}
	if score < -0.2 {
		return TrendDeclining
	}
	return TrendStable
}

// WindowToDuration 将 TimeWindow 转为 time.Duration
func WindowToDuration(w TimeWindow) time.Duration {
	switch w {
	case Window30d:
		return 30 * 24 * time.Hour
	default:
		return 7 * 24 * time.Hour
	}
}
