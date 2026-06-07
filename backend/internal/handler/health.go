package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 处理健康检查
type HealthHandler struct {
	dbHealthy    bool
	redisHealthy bool
}

// NewHealthHandler 创建新的 HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// SetDBHealthy 设置数据库健康状态
func (h *HealthHandler) SetDBHealthy(healthy bool) {
	h.dbHealthy = healthy
}

// SetRedisHealthy 设置 Redis 健康状态
func (h *HealthHandler) SetRedisHealthy(healthy bool) {
	h.redisHealthy = healthy
}

// Health 处理 GET /api/v1/health
func (h *HealthHandler) Health(c *gin.Context) {
	dbStatus := "connected"
	if !h.dbHealthy {
		dbStatus = "disconnected"
	}

	redisStatus := "connected"
	if !h.redisHealthy {
		redisStatus = "disconnected"
	}

	status := "ok"
	if !h.dbHealthy {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"database": dbStatus,
		"redis":    redisStatus,
	})
}
