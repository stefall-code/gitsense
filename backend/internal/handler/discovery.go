package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gitsense/gitsense/backend/internal/cache"
	"github.com/gitsense/gitsense/backend/internal/discovery"
)

// DiscoveryHandler 发现 API handler（缓存逻辑在 Handler 层）
type DiscoveryHandler struct {
	service *discovery.Service
	cache   *cache.Client
}

// NewDiscoveryHandler 创建 DiscoveryHandler
func NewDiscoveryHandler(service *discovery.Service, cacheClient *cache.Client) *DiscoveryHandler {
	return &DiscoveryHandler{service: service, cache: cacheClient}
}

// cacheStatus 根据 cache.Client 可用性确定默认状态
func (h *DiscoveryHandler) defaultCacheStatus() cache.CacheStatus {
	if h.cache == nil || !h.cache.IsAvailable() {
		return cache.CacheBypass
	}
	return cache.CacheMiss
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

	// Handler 层查缓存
	cacheKey := cache.DiscoveryKey(owner, repo)
	if h.cache != nil {
		var cached discovery.DiscoveryResponse
		if status, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil {
			c.Header("X-Cache", string(status))
			c.JSON(http.StatusOK, &cached)
			return
		}
	}

	// 缓存未命中，调 Service
	result, err := h.service.Discover(c.Request.Context(), fullName, limit)
	if err != nil {
		c.Header("X-Cache", string(h.defaultCacheStatus()))
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	// 写入缓存
	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, result, cache.DiscoveryTTL)
	}

	c.Header("X-Cache", string(h.defaultCacheStatus()))
	c.JSON(http.StatusOK, result)
}

// ListEcosystems 列出所有生态 GET /api/v1/ecosystems
func (h *DiscoveryHandler) ListEcosystems(c *gin.Context) {
	// Handler 层查缓存
	cacheKey := cache.EcosystemsListKey()
	if h.cache != nil {
		var cached discovery.EcosystemsResponse
		if status, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil {
			c.Header("X-Cache", string(status))
			c.JSON(http.StatusOK, &cached)
			return
		}
	}

	// 缓存未命中，调 Service
	result, err := h.service.ListEcosystems(c.Request.Context())
	if err != nil {
		c.Header("X-Cache", string(h.defaultCacheStatus()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// 写入缓存
	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, result, cache.EcosystemTTL)
	}

	c.Header("X-Cache", string(h.defaultCacheStatus()))
	c.JSON(http.StatusOK, result)
}

// GetEcosystem 获取生态详情 GET /api/v1/ecosystem/:name
func (h *DiscoveryHandler) GetEcosystem(c *gin.Context) {
	name := c.Param("name")

	// Handler 层查缓存
	cacheKey := cache.EcosystemKey(name)
	if h.cache != nil {
		var cached discovery.EcosystemDetail
		if status, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil {
			c.Header("X-Cache", string(status))
			c.JSON(http.StatusOK, &cached)
			return
		}
	}

	// 缓存未命中，调 Service
	result, err := h.service.GetEcosystem(c.Request.Context(), name)
	if err != nil {
		c.Header("X-Cache", string(h.defaultCacheStatus()))
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "ECOSYSTEM_NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	// 写入缓存
	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, result, cache.EcosystemTTL)
	}

	c.Header("X-Cache", string(h.defaultCacheStatus()))
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

	// Handler 层查缓存
	cacheKey := cache.TrendingKey(name)
	if h.cache != nil {
		var cached discovery.TrendingResponse
		if status, err := h.cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil {
			c.Header("X-Cache", string(status))
			c.JSON(http.StatusOK, &cached)
			return
		}
	}

	// 缓存未命中，调 Service
	result, err := h.service.GetTrending(c.Request.Context(), name, window, limit)
	if err != nil {
		c.Header("X-Cache", string(h.defaultCacheStatus()))
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "ECOSYSTEM_NOT_FOUND",
				"message": err.Error(),
			},
		})
		return
	}

	// 写入缓存
	if h.cache != nil {
		_ = h.cache.SetJSON(c.Request.Context(), cacheKey, result, cache.TrendingTTL)
	}

	c.Header("X-Cache", string(h.defaultCacheStatus()))
	c.JSON(http.StatusOK, result)
}
