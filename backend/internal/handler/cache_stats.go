package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gitsense/gitsense/backend/internal/cache"
)

// CacheStatsHandler 缓存统计 API handler
type CacheStatsHandler struct {
	cacheClient *cache.Client
}

// NewCacheStatsHandler 创建 CacheStatsHandler
func NewCacheStatsHandler(cacheClient *cache.Client) *CacheStatsHandler {
	return &CacheStatsHandler{cacheClient: cacheClient}
}

// GetStats 处理 GET /api/v1/admin/cache/stats
func (h *CacheStatsHandler) GetStats(c *gin.Context) {
	if h.cacheClient == nil {
		c.JSON(http.StatusOK, cache.StatsResponse{RedisConnected: false})
		return
	}
	c.JSON(http.StatusOK, h.cacheClient.GetStats())
}
