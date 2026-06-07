package handler

import (
	"net/http"
	"strconv"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/trend"
	"github.com/gin-gonic/gin"
)

// TrendHandler 处理 Trend API
type TrendHandler struct {
	trendService *trend.Service
}

// NewTrendHandler 创建新的 TrendHandler
func NewTrendHandler(trendService *trend.Service) *TrendHandler {
	return &TrendHandler{trendService: trendService}
}

// GetTopicTrends 处理 GET /api/v1/trends/topics
func (h *TrendHandler) GetTopicTrends(c *gin.Context) {
	window := trend.TimeWindow(c.DefaultQuery("window", "7d"))
	if window != trend.Window7d && window != trend.Window30d {
		window = trend.Window7d
	}

	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	ctx := c.Request.Context()

	trends, err := h.trendService.GetTopicTrends(ctx, window, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "TREND_QUERY_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"window": window,
		"count":  len(trends),
	})
}

// GetEcosystemTrends 处理 GET /api/v1/trends/ecosystems
func (h *TrendHandler) GetEcosystemTrends(c *gin.Context) {
	window := trend.TimeWindow(c.DefaultQuery("window", "7d"))
	if window != trend.Window7d && window != trend.Window30d {
		window = trend.Window7d
	}

	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	ctx := c.Request.Context()

	trends, err := h.trendService.GetEcosystemTrends(ctx, window, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "TREND_QUERY_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"window": window,
		"count":  len(trends),
	})
}

// GetOverview 处理 GET /api/v1/trends/overview
func (h *TrendHandler) GetOverview(c *gin.Context) {
	window := trend.TimeWindow(c.DefaultQuery("window", "7d"))
	if window != trend.Window7d && window != trend.Window30d {
		window = trend.Window7d
	}

	ctx := c.Request.Context()

	overview, err := h.trendService.GetOverview(ctx, window)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "TREND_QUERY_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}
