package service

import (
	"context"
	"math"
)

// EmbeddingProvider 定义 Embedding 生成接口
// 不绑定具体实现，支持 OpenAI / 本地模型 / 自定义服务
type EmbeddingProvider interface {
	// Generate 生成单条文本的向量表示
	Generate(ctx context.Context, text string) ([]float32, error)

	// GenerateBatch 批量生成向量（性能关键路径）
	GenerateBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimensions 返回向量维度
	Dimensions() int
}

// SimilarityStrategy 定义相似度计算策略接口
// 默认实现加权评分，可扩展其他策略
type SimilarityStrategy interface {
	// Calculate 计算两个仓库的相似度评分和推荐原因
	Calculate(ctx context.Context, target RepoFeatures, candidate RepoFeatures) (score float64, reason RecommendationReason, err error)
	// Name 返回策略名称
	Name() string
}

// RepoFeatures 表示用于相似度计算的仓库特征
type RepoFeatures struct {
	Embedding []float32
	Topics    []string
	Language  string
	Stars     int
}

// RecommendationReason 表示推荐原因（service 层定义，与 model 层解耦）
type RecommendationReason struct {
	EmbeddingSimilarity float64
	TopicSimilarity     float64
	LanguageMatch       bool
	PopularityScore     float64
	CommonTopics        []string
}

// ScoreBreakdown debug 模式下的评分分解
type ScoreBreakdown struct {
	EmbeddingScore   float64 `json:"embedding_score"`
	TopicScore       float64 `json:"topic_score"`
	LanguageScore    float64 `json:"language_score"`
	PopularityScore  float64 `json:"popularity_score"`
	TrendScore       float64 `json:"trend_score,omitempty"`
	FinalScore       float64 `json:"final_score"`
}

// --- V1 Strategy (0.7/0.2/0.1) 保留，不覆盖 ---

// WeightedSimilarityStrategy V1 加权相似度策略
// score = 0.7 * embedding_similarity + 0.2 * topic_similarity + 0.1 * language_similarity
type WeightedSimilarityStrategy struct{}

func NewWeightedSimilarityStrategy() *WeightedSimilarityStrategy {
	return &WeightedSimilarityStrategy{}
}

func (s *WeightedSimilarityStrategy) Name() string { return "weighted_v1" }

func (s *WeightedSimilarityStrategy) Calculate(ctx context.Context, target RepoFeatures, candidate RepoFeatures) (float64, RecommendationReason, error) {
	embSim := cosineSimilarity(target.Embedding, candidate.Embedding)
	topicSim := jaccardSimilarity(target.Topics, candidate.Topics)
	langMatch := target.Language != "" && target.Language == candidate.Language
	commonTopics := intersection(target.Topics, candidate.Topics)

	var langScore float64
	if langMatch {
		langScore = 1.0
	}

	score := 0.7*embSim + 0.2*topicSim + 0.1*langScore

	reason := RecommendationReason{
		EmbeddingSimilarity: embSim,
		TopicSimilarity:     topicSim,
		LanguageMatch:       langMatch,
		CommonTopics:        commonTopics,
	}

	return score, reason, nil
}

// --- V2 Strategy (0.6/0.2/0.1/0.1) 新增 ---

// GlobalMaxStars 全局最大 stars，用于 popularity 归一化
// 启动时预计算或使用固定值，保证同一 repo 在不同查询中 score 一致
var GlobalMaxStars float64 = 5000000 // ~top repo on GitHub

// WeightedSimilarityStrategyV2 多因子加权相似度策略
// score = 0.6 * embedding_similarity + 0.2 * topic_similarity + 0.1 * language_match + 0.1 * popularity
type WeightedSimilarityStrategyV2 struct{}

func NewWeightedSimilarityStrategyV2() *WeightedSimilarityStrategyV2 {
	return &WeightedSimilarityStrategyV2{}
}

func (s *WeightedSimilarityStrategyV2) Name() string { return "weighted_v2" }

func (s *WeightedSimilarityStrategyV2) Calculate(ctx context.Context, target RepoFeatures, candidate RepoFeatures) (float64, RecommendationReason, error) {
	embSim := cosineSimilarity(target.Embedding, candidate.Embedding)
	topicSim := jaccardSimilarity(target.Topics, candidate.Topics)
	langMatch := target.Language != "" && target.Language == candidate.Language
	popScore := normalizePopularity(candidate.Stars)
	commonTopics := intersection(target.Topics, candidate.Topics)

	var langScore float64
	if langMatch {
		langScore = 1.0
	}

	score := 0.6*embSim + 0.2*topicSim + 0.1*langScore + 0.1*popScore

	reason := RecommendationReason{
		EmbeddingSimilarity: embSim,
		TopicSimilarity:     topicSim,
		LanguageMatch:       langMatch,
		PopularityScore:     popScore,
		CommonTopics:        commonTopics,
	}

	return score, reason, nil
}

// CalculateWithBreakdown V2 专用：返回评分分解（debug 模式）
func (s *WeightedSimilarityStrategyV2) CalculateWithBreakdown(ctx context.Context, target RepoFeatures, candidate RepoFeatures) (ScoreBreakdown, RecommendationReason, error) {
	embSim := cosineSimilarity(target.Embedding, candidate.Embedding)
	topicSim := jaccardSimilarity(target.Topics, candidate.Topics)
	langMatch := target.Language != "" && target.Language == candidate.Language
	popScore := normalizePopularity(candidate.Stars)
	commonTopics := intersection(target.Topics, candidate.Topics)

	var langScore float64
	if langMatch {
		langScore = 1.0
	}

	embWeighted := 0.6 * embSim
	topicWeighted := 0.2 * topicSim
	langWeighted := 0.1 * langScore
	popWeighted := 0.1 * popScore
	finalScore := embWeighted + topicWeighted + langWeighted + popWeighted

	reason := RecommendationReason{
		EmbeddingSimilarity: embSim,
		TopicSimilarity:     topicSim,
		LanguageMatch:       langMatch,
		PopularityScore:     popScore,
		CommonTopics:        commonTopics,
	}

	breakdown := ScoreBreakdown{
		EmbeddingScore:  embWeighted,
		TopicScore:      topicWeighted,
		LanguageScore:   langWeighted,
		PopularityScore: popWeighted,
		FinalScore:      finalScore,
	}

	return breakdown, reason, nil
}

// normalizePopularity 使用全局 max stars 归一化
// popularity_score = log(stars + 1) / log(global_max_stars + 1)
func normalizePopularity(stars int) float64 {
	if stars <= 0 {
		return 0
	}
	return math.Log(float64(stars)+1) / math.Log(GlobalMaxStars+1)
}

// --- 数学工具函数 ---

func cosineSimilarity(a, b []float32) float64 {
	if a == nil || b == nil || len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	setA := make(map[string]bool, len(a))
	for _, s := range a {
		setA[s] = true
	}
	intersectionCount := 0
	for _, s := range b {
		if setA[s] {
			intersectionCount++
		}
	}
	unionCount := len(a) + len(b) - intersectionCount
	if unionCount == 0 {
		return 0
	}
	return float64(intersectionCount) / float64(unionCount)
}

func intersection(a, b []string) []string {
	setA := make(map[string]bool, len(a))
	for _, s := range a {
		setA[s] = true
	}
	var result []string
	for _, s := range b {
		if setA[s] {
			result = append(result, s)
		}
	}
	return result
}

// topicOverlapCount 返回两个 topic 列表的交集数量
func topicOverlapCount(a, b []string) int {
	setA := make(map[string]bool, len(a))
	for _, s := range a {
		setA[s] = true
	}
	count := 0
	for _, s := range b {
		if setA[s] {
			count++
		}
	}
	return count
}

// minTopicOverlap 根据目标仓库 topics 数量返回最小重叠阈值
// 分层阈值：>=3 topics 时 overlap>=2，2 topics 时 overlap>=1，1 topic 时 fallback
func minTopicOverlap(targetTopicCount int) int {
	if targetTopicCount >= 3 {
		return 2
	}
	return 1
}

// --- V3 Strategy (0.45/0.20/0.10/0.10/0.15) Trend-aware ---

// TrendScoreProvider 趋势分数提供接口（由 trend 包注入）
type TrendScoreProvider interface {
	GetTopicTrendScore(ctx context.Context, topic string, window string) float64
}

// WeightedSimilarityStrategyV3 Trend-aware 多因子加权策略
// score = 0.45*embedding + 0.20*graph_signal + 0.10*topic + 0.10*popularity + 0.15*trend
// graph_signal 和 trend 由外部注入，默认 0
type WeightedSimilarityStrategyV3 struct {
	trendProvider TrendScoreProvider
	enabled       bool // feature flag: ENABLE_TREND_RANKING
}

// NewWeightedSimilarityStrategyV3 创建 V3 策略
func NewWeightedSimilarityStrategyV3(trendProvider TrendScoreProvider, enabled bool) *WeightedSimilarityStrategyV3 {
	return &WeightedSimilarityStrategyV3{
		trendProvider: trendProvider,
		enabled:       enabled,
	}
}

func (s *WeightedSimilarityStrategyV3) Name() string { return "weighted_v3" }

func (s *WeightedSimilarityStrategyV3) Calculate(ctx context.Context, target RepoFeatures, candidate RepoFeatures) (float64, RecommendationReason, error) {
	embSim := cosineSimilarity(target.Embedding, candidate.Embedding)
	topicSim := jaccardSimilarity(target.Topics, candidate.Topics)
	langMatch := target.Language != "" && target.Language == candidate.Language
	popScore := normalizePopularity(candidate.Stars)
	commonTopics := intersection(target.Topics, candidate.Topics)

	var langScore float64
	if langMatch {
		langScore = 1.0
	}
	_ = langScore // reserved for future use

	// Trend score（归一化到 [0, 1]）
	var trendScore float64
	if s.enabled && s.trendProvider != nil && len(candidate.Topics) > 0 {
		totalTrend := float64(0)
		for _, t := range candidate.Topics {
			totalTrend += s.trendProvider.GetTopicTrendScore(ctx, t, "7d")
		}
		// 平均 + 偏移到 [0, 1]（原始 [-1, 1]）
		avgTrend := totalTrend / float64(len(candidate.Topics))
		trendScore = (avgTrend + 1) / 2 // [-1,1] → [0,1]
	}

	// V3 权重：0.45 emb + 0.20 graph(暂用 emb 近似) + 0.10 topic + 0.10 pop + 0.15 trend
	// graph_signal 在 ranking pipeline 中单独处理，这里用 0.20 作为预留
	graphSignal := embSim * 0.8 // 简化：用 embedding 相似度近似 graph signal
	if !s.enabled {
		// 未启用 trend 时，trend 权重归零，重新分配
		score := 0.55*embSim + 0.20*graphSignal + 0.10*topicSim + 0.15*popScore
		reason := RecommendationReason{
			EmbeddingSimilarity: embSim,
			TopicSimilarity:     topicSim,
			LanguageMatch:       langMatch,
			PopularityScore:     popScore,
			CommonTopics:        commonTopics,
		}
		return score, reason, nil
	}

	score := 0.45*embSim + 0.20*graphSignal + 0.10*topicSim + 0.10*popScore + 0.15*trendScore

	reason := RecommendationReason{
		EmbeddingSimilarity: embSim,
		TopicSimilarity:     topicSim,
		LanguageMatch:       langMatch,
		PopularityScore:     popScore,
		CommonTopics:        commonTopics,
	}

	return score, reason, nil
}

// CalculateWithBreakdown V3 专用：返回评分分解（debug 模式）
func (s *WeightedSimilarityStrategyV3) CalculateWithBreakdown(ctx context.Context, target RepoFeatures, candidate RepoFeatures) (ScoreBreakdown, RecommendationReason, error) {
	embSim := cosineSimilarity(target.Embedding, candidate.Embedding)
	topicSim := jaccardSimilarity(target.Topics, candidate.Topics)
	langMatch := target.Language != "" && target.Language == candidate.Language
	popScore := normalizePopularity(candidate.Stars)
	commonTopics := intersection(target.Topics, candidate.Topics)

	var langScore float64
	if langMatch {
		langScore = 1.0
	}
	_ = langScore // reserved for future use

	var trendScore float64
	if s.enabled && s.trendProvider != nil && len(candidate.Topics) > 0 {
		totalTrend := float64(0)
		for _, t := range candidate.Topics {
			totalTrend += s.trendProvider.GetTopicTrendScore(ctx, t, "7d")
		}
		avgTrend := totalTrend / float64(len(candidate.Topics))
		trendScore = (avgTrend + 1) / 2
	}

	graphSignal := embSim * 0.8

	var embWeighted, graphWeighted, topicWeighted, popWeighted, trendWeighted, finalScore float64

	if s.enabled {
		embWeighted = 0.45 * embSim
		graphWeighted = 0.20 * graphSignal
		topicWeighted = 0.10 * topicSim
		popWeighted = 0.10 * popScore
		trendWeighted = 0.15 * trendScore
		finalScore = embWeighted + graphWeighted + topicWeighted + popWeighted + trendWeighted
	} else {
		embWeighted = 0.55 * embSim
		graphWeighted = 0.20 * graphSignal
		topicWeighted = 0.10 * topicSim
		popWeighted = 0.15 * popScore
		trendWeighted = 0
		finalScore = embWeighted + graphWeighted + topicWeighted + popWeighted
	}

	reason := RecommendationReason{
		EmbeddingSimilarity: embSim,
		TopicSimilarity:     topicSim,
		LanguageMatch:       langMatch,
		PopularityScore:     popScore,
		CommonTopics:        commonTopics,
	}

	breakdown := ScoreBreakdown{
		EmbeddingScore:  embWeighted,
		TopicScore:      topicWeighted,
		LanguageScore:   0, // V3 不单独加权 language
		PopularityScore: popWeighted,
		TrendScore:      trendWeighted,
		FinalScore:      finalScore,
	}

	return breakdown, reason, nil
}
