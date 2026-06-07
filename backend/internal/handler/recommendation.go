package handler

import (
	"net/http"
	"strconv"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// RecommendationHandler 处理推荐相关请求
type RecommendationHandler struct {
	recommendation *service.RecommendationService
}

// NewRecommendationHandler 创建新的 RecommendationHandler
func NewRecommendationHandler(recommendation *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendation: recommendation}
}

// GetRecommendations 处理 GET /api/v1/repos/:owner/:name/recommendations
func (h *RecommendationHandler) GetRecommendations(c *gin.Context) {
	owner := c.Param("owner")
	name := c.Param("name")
	fullName := owner + "/" + name

	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	debug := c.Query("debug") == "true"
	strategy := c.Query("strategy") // "v1" or "v2"

	recs, err := h.recommendation.GetRecommendations(c.Request.Context(), fullName, limit, debug, strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "RECOMMENDATION_FAILED", Message: err.Error()},
		})
		return
	}

	resp := gin.H{"similar_repositories": recs}
	if debug {
		if strategy == "v1" {
			resp["strategy"] = "weighted_v1"
		} else {
			resp["strategy"] = "weighted_v2"
		}
	}

	c.JSON(http.StatusOK, resp)
}
