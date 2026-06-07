package handler

import (
	"net/http"
	"strconv"

	"github.com/gitsense/gitsense/backend/internal/graph"
	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// GraphHandler 处理 Graph 相关 API
type GraphHandler struct {
	graphService *graph.Service
	graphStore   *graph.Store
}

// NewGraphHandler 创建新的 GraphHandler
func NewGraphHandler(graphService *graph.Service, graphStore *graph.Store) *GraphHandler {
	return &GraphHandler{graphService: graphService, graphStore: graphStore}
}

// GetRepoGraph 处理 GET /api/v1/graph/repo/:owner/:repo
// 获取 Repo 的图邻域（相似边、topics、language、ecosystem）
func (h *GraphHandler) GetRepoGraph(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	fullName := owner + "/" + repo

	ctx := c.Request.Context()

	result, err := h.graphService.GetRepoGraph(ctx, fullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "GRAPH_QUERY_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetEcosystemGraph 处理 GET /api/v1/graph/ecosystem/:name
// 获取生态图（repos、topic cluster、nodes）
func (h *GraphHandler) GetEcosystemGraph(c *gin.Context) {
	name := c.Param("name")

	ctx := c.Request.Context()

	result, err := h.graphService.GetEcosystemGraph(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "GRAPH_QUERY_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// FindPaths 处理 GET /api/v1/graph/path
// 查找两个 repo 之间的路径（1-hop / 2-hop）
func (h *GraphHandler) FindPaths(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "INVALID_REQUEST", Message: "both 'from' and 'to' query parameters are required"},
		})
		return
	}

	maxHops := 2
	if v := c.Query("max_hops"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxHops = n
		}
	}

	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	ctx := c.Request.Context()

	result, err := h.graphService.FindPaths(ctx, from, to, maxHops, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "PATH_QUERY_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGraphExplanation 处理 GET /api/v1/graph/explanation
// 获取两个 repo 之间的 Graph 解释（实时计算，debug only）
func (h *GraphHandler) GetGraphExplanation(c *gin.Context) {
	repoA := c.Query("repo_a")
	repoB := c.Query("repo_b")

	if repoA == "" || repoB == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "INVALID_REQUEST", Message: "both 'repo_a' and 'repo_b' query parameters are required"},
		})
		return
	}

	ctx := c.Request.Context()

	result, err := h.graphService.GetGraphExplanation(ctx, repoA, repoB, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "EXPLANATION_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGraphMetrics 处理 GET /admin/graph/metrics
func (h *GraphHandler) GetGraphMetrics(c *gin.Context) {
	metrics, err := h.graphStore.GetGraphMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}
