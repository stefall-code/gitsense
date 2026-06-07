package handler

import (
	"net/http"
	"strconv"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/pgvector/pgvector-go"
)

// SearchSimilarHandler 处理文本相似搜索
type SearchSimilarHandler struct {
	embeddingSvc   *service.EmbeddingService
	recommendation *service.RecommendationService
	classifier     *service.EcosystemClassifier
}

// NewSearchSimilarHandler 创建新的 SearchSimilarHandler
func NewSearchSimilarHandler(
	embeddingSvc *service.EmbeddingService,
	recommendation *service.RecommendationService,
	classifier *service.EcosystemClassifier,
) *SearchSimilarHandler {
	return &SearchSimilarHandler{
		embeddingSvc:   embeddingSvc,
		recommendation: recommendation,
		classifier:     classifier,
	}
}

// SearchSimilar 处理 GET /api/v1/search/similar?text=...&limit=10
// 文本 → embedding → pgvector search → ranked repos
func (h *SearchSimilarHandler) SearchSimilar(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "INVALID_REQUEST", Message: "text parameter is required"},
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	ctx := c.Request.Context()

	// 1. 生成 query embedding
	vec, err := h.embeddingSvc.GenerateEmbedding(ctx, text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "EMBEDDING_FAILED", Message: err.Error()},
		})
		return
	}

	// 2. pgvector 相似搜索
	embedding := pgvector.NewVector(vec)
	// 使用空 full_name 表示不是排除自身
	candidates, err := h.recommendation.SearchByEmbedding(ctx, embedding, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "SEARCH_FAILED", Message: err.Error()},
		})
		return
	}

	// 3. 为每个候选生成生态标签
	type SimilarResult struct {
		Repository   model.Repository          `json:"repository"`
		Score        float64                   `json:"score"`
		EcosystemTag string                    `json:"ecosystem,omitempty"`
	}

	results := make([]SimilarResult, len(candidates))
	for i, repo := range candidates {
		ecoTag := h.classifier.Classify(repo.Topics, repo.Description)
		results[i] = SimilarResult{
			Repository:   repo,
			EcosystemTag: ecoTag,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"query_text": text,
		"results":    results,
	})
}
