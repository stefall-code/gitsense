package service

import (
	"fmt"

	"github.com/gitsense/gitsense/backend/internal/model"
)

// ExplanationInput 生成 explanation 所需的辅助信息（features 之外的上下文）
type ExplanationInput struct {
	Stars          int
	EcosystemMatch bool
	CommonTopics   []string
	DirectEdge     bool
	TwoHop         bool
}

// ExplanationGenerator 生成推荐理由（纯加层，不修改 ranking 逻辑）
// 核心约束：reasons 必须完全基于 features 派生，不允许引入额外隐含逻辑
type ExplanationGenerator struct{}

// NewExplanationGenerator 创建 ExplanationGenerator
func NewExplanationGenerator() *ExplanationGenerator {
	return &ExplanationGenerator{}
}

// Generate 根据特征分数和辅助信息生成人类可读的推荐理由列表
// features 是单一数据源，reasons 中的数值必须来自 features
func (g *ExplanationGenerator) Generate(features model.RecommendationFeatures, input ExplanationInput) []string {
	var reasons []string

	// 1. Embedding reason — 基于 features.EmbeddingScore
	if features.EmbeddingScore > 0.8 {
		reasons = append(reasons, fmt.Sprintf("High semantic similarity (embedding_score=%.2f)", features.EmbeddingScore))
	} else if features.EmbeddingScore > 0.6 {
		reasons = append(reasons, fmt.Sprintf("Moderate semantic similarity (embedding_score=%.2f)", features.EmbeddingScore))
	}

	// 2. Graph reason — 基于 features.GraphScore
	if features.GraphScore > 0.5 {
		reasons = append(reasons, fmt.Sprintf("Strong graph connectivity (graph_score=%.2f)", features.GraphScore))
	} else if input.DirectEdge {
		reasons = append(reasons, "Direct graph similarity connection")
	} else if input.TwoHop {
		reasons = append(reasons, "Connected via 2-hop graph path")
	}

	// 3. Trend reason — 基于 features.TrendScore（归一化到 [0,1]）
	if features.TrendScore > 0.6 {
		reasons = append(reasons, fmt.Sprintf("Rising trend in ecosystem (trend_score=%.2f)", features.TrendScore))
	} else if features.TrendScore < 0.4 && features.TrendScore > 0 {
		reasons = append(reasons, fmt.Sprintf("Declining trend signal (trend_score=%.2f)", features.TrendScore))
	}

	// 4. Popularity reason — 基于 features.PopularityScore
	if features.PopularityScore > 0.5 {
		reasons = append(reasons, fmt.Sprintf("High popularity (popularity_score=%.2f, stars=%d)", features.PopularityScore, input.Stars))
	} else if input.Stars > 1000 {
		reasons = append(reasons, fmt.Sprintf("Popular project (stars=%d)", input.Stars))
	}

	// 5. Topic reason — 基于 features.TopicScore
	if features.TopicScore > 0.7 {
		reasons = append(reasons, fmt.Sprintf("Strong topic overlap (topic_score=%.2f)", features.TopicScore))
	} else if features.TopicScore > 0.4 && len(input.CommonTopics) > 0 {
		topics := input.CommonTopics
		if len(topics) > 3 {
			topics = topics[:3]
		}
		reasons = append(reasons, fmt.Sprintf("Shared topics: %v (topic_score=%.2f)", topics, features.TopicScore))
	}

	// 6. Ecosystem reason
	if input.EcosystemMatch {
		reasons = append(reasons, "Same technology ecosystem")
	}

	// 保底：至少一个理由
	if len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf("Similarity score: %.2f", features.EmbeddingScore))
	}

	return reasons
}
