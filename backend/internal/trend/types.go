package trend

import (
	"context"
	"time"
)

// TrendStatus 趋势状态
type TrendStatus string

const (
	TrendRising    TrendStatus = "rising"
	TrendStable    TrendStatus = "stable"
	TrendDeclining TrendStatus = "declining"
)

// TimeWindow 时间窗口
type TimeWindow string

const (
	Window7d  TimeWindow = "7d"
	Window30d TimeWindow = "30d"
)

// TopicTrend Topic 趋势数据
type TopicTrend struct {
	Topic       string      `json:"topic"`
	GrowthRate  float64     `json:"growth_rate"`
	TrendScore  float64     `json:"trend_score"`
	Status      TrendStatus `json:"status"`
	Count7d     int         `json:"count_7d"`
	CountPrev7d int         `json:"count_prev_7d"`
	Window      TimeWindow  `json:"window"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// EcosystemTrend 生态趋势数据
type EcosystemTrend struct {
	Ecosystem   string      `json:"ecosystem"`
	GrowthRate  float64     `json:"growth_rate"`
	TrendScore  float64     `json:"trend_score"`
	Status      TrendStatus `json:"status"`
	Count7d     int         `json:"count_7d"`
	CountPrev7d int         `json:"count_prev_7d"`
	Window      TimeWindow  `json:"window"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// TrendOverview 趋势总览
type TrendOverview struct {
	TopRisingTopics     []TopicTrend     `json:"top_rising_topics"`
	TopRisingEcosystems []EcosystemTrend `json:"top_rising_ecosystems"`
	Window              TimeWindow       `json:"window"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

// TrendProvider 趋势数据提供接口
type TrendProvider interface {
	// GetTopicTrendScore 获取单个 topic 的趋势分数（归一化到 [-1, 1]）
	GetTopicTrendScore(ctx context.Context, topic string, window TimeWindow) float64
	// GetEcosystemTrendScore 获取单个 ecosystem 的趋势分数
	GetEcosystemTrendScore(ctx context.Context, ecosystem string, window TimeWindow) float64
}
