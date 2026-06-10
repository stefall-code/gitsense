package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gitsense/gitsense/backend/internal/discovery"
)

// DiscoveryHandler 发现 API handler
type DiscoveryHandler struct {
	service *discovery.Service
}

// NewDiscoveryHandler 创建 DiscoveryHandler
func NewDiscoveryHandler(service *discovery.Service) *DiscoveryHandler {
	return &DiscoveryHandler{service: service}
}

// Discover 一站式发现 GET /api/v1/discovery/:owner/:repo
func (h *DiscoveryHandler) Discover(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	fullName := owner + "/" + repo

	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	result, err := h.service.Discover(c.Request.Context(), fullName, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListEcosystems 列出所有生态 GET /api/v1/ecosystems
func (h *DiscoveryHandler) ListEcosystems(c *gin.Context) {
	result, err := h.service.ListEcosystems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetEcosystem 获取生态详情 GET /api/v1/ecosystem/:name
func (h *DiscoveryHandler) GetEcosystem(c *gin.Context) {
	name := c.Param("name")

	result, err := h.service.GetEcosystem(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "ECOSYSTEM_NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTrending 获取生态内趋势项目 GET /api/v1/ecosystem/:name/trending
func (h *DiscoveryHandler) GetTrending(c *gin.Context) {
	name := c.Param("name")
	window := c.DefaultQuery("window", "7d")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	result, err := h.service.GetTrending(c.Request.Context(), name, window, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "ECOSYSTEM_NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
