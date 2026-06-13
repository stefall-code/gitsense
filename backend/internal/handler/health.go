package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gitsense/gitsense/backend/internal/cache"
)

// HealthHandler 处理健康检查
type HealthHandler struct {
	cacheClient *cache.Client
	dbHealthy   bool
}

// NewHealthHandler 创建新的 HealthHandler
func NewHealthHandler(cacheClient *cache.Client) *HealthHandler {
	return &HealthHandler{cacheClient: cacheClient}
}

// SetDBHealthy 设置数据库健康状态
func (h *HealthHandler) SetDBHealthy(healthy bool) {
	h.dbHealthy = healthy
}

// Health 处理 GET /api/v1/health
func (h *HealthHandler) Health(c *gin.Context) {
	dbStatus := "connected"
	if !h.dbHealthy {
		dbStatus = "disconnected"
	}

	redisStatus := "disconnected"
	if h.cacheClient != nil && h.cacheClient.IsAvailable() {
		redisStatus = "connected"
	}

	status := "ok"
	if !h.dbHealthy {
		status = "degraded"
	} else if redisStatus == "disconnected" {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"database": dbStatus,
		"redis":    redisStatus,
	})
}
