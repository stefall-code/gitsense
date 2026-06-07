package trend

import (
	"context"
)

// TrendScoreProviderAdapter 适配 trend.CachedTrendProvider → service.TrendScoreProvider
// 解决跨包接口兼容问题
type TrendScoreProviderAdapter struct {
	provider *CachedTrendProvider
}

// NewTrendScoreProviderAdapter 创建适配器
func NewTrendScoreProviderAdapter(provider *CachedTrendProvider) *TrendScoreProviderAdapter {
	return &TrendScoreProviderAdapter{provider: provider}
}

// GetTopicTrendScore 实现 service.TrendScoreProvider 接口
func (a *TrendScoreProviderAdapter) GetTopicTrendScore(ctx context.Context, topic string, window string) float64 {
	return a.provider.GetTopicTrendScore(ctx, topic, TimeWindow(window))
}
