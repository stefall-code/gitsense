package model

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// EmbeddingStatus 表示 Embedding 生成状态
type EmbeddingStatus string

const (
	EmbeddingPending    EmbeddingStatus = "pending"
	EmbeddingGenerating EmbeddingStatus = "generating"
	EmbeddingDone       EmbeddingStatus = "done"
	EmbeddingFailed     EmbeddingStatus = "failed"
)

// Repository 表示一个 GitHub 仓库
type Repository struct {
	ID              int64           `json:"id"`
	FullName        string          `json:"full_name"`
	Owner           string          `json:"owner"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Language        string          `json:"language"`
	Stars           int             `json:"stars"`
	Topics          []string        `json:"topics"`
	Readme          string          `json:"readme,omitempty"`
	Embedding       *pgvector.Vector `json:"-"`
	EmbeddingStatus EmbeddingStatus `json:"embedding_status"`
	RetryCount      int             `json:"retry_count"`
	LastAttemptAt   *time.Time      `json:"last_attempt_at,omitempty"`
	LastActivityAt  *time.Time      `json:"last_activity_at,omitempty"`
	PushedAt        *time.Time      `json:"pushed_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// RecommendationReason 表示单条推荐的推荐原因
type RecommendationReason struct {
	EmbeddingSimilarity float64  `json:"embedding_similarity"`
	TopicSimilarity     float64  `json:"topic_similarity"`
	LanguageMatch       bool     `json:"language_match"`
	PopularityScore     float64  `json:"popularity_score"`
	CommonTopics        []string `json:"common_topics"`
}

// ScoreBreakdown debug 模式下的评分分解
type ScoreBreakdown struct {
	EmbeddingScore  float64 `json:"embedding_score"`
	TopicScore      float64 `json:"topic_score"`
	LanguageScore   float64 `json:"language_score"`
	PopularityScore float64 `json:"popularity_score"`
	TrendScore      float64 `json:"trend_score,omitempty"`
	FinalScore      float64 `json:"final_score"`
}

// RecommendationFeatures 推荐特征分数（Explanation Contract 的单一数据源）
// reasons 必须完全基于此结构派生，不允许引入额外隐含逻辑
type RecommendationFeatures struct {
	EmbeddingScore  float64 `json:"embedding_score"`
	GraphScore      float64 `json:"graph_score"`
	TrendScore      float64 `json:"trend_score"`
	PopularityScore float64 `json:"popularity_score"`
	TopicScore      float64 `json:"topic_score"`
}

// SimilarRepository 表示一个相似仓库及其推荐原因
type SimilarRepository struct {
	Repository   Repository             `json:"repository"`
	Score        float64                `json:"score"`
	Features     RecommendationFeatures `json:"features"`
	Reasons      []string               `json:"reasons"`
	EcosystemTag string                 `json:"ecosystem,omitempty"`
	Breakdown    *ScoreBreakdown        `json:"breakdown,omitempty"` // debug mode only
}

// Ecosystem 表示技术生态
type Ecosystem struct {
	Name         string       `json:"name"`
	Repositories []Repository `json:"repositories"`
}

// SearchRequest 表示搜索请求
type SearchRequest struct {
	RepoURL string `json:"repo_url" binding:"required"`
	Limit   int    `json:"limit,omitempty"`
}

// SearchResponse 表示搜索响应
type SearchResponse struct {
	Repository          *Repository          `json:"repository"`
	SimilarRepositories []SimilarRepository `json:"similar_repositories"`
	Ecosystem          *Ecosystem           `json:"ecosystem,omitempty"`
	Strategy           string               `json:"strategy,omitempty"`  // debug: 使用的策略名
	Debug              bool                 `json:"debug,omitempty"`     // debug: 是否开启
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 表示错误详情
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
