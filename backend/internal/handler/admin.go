package handler

import (
	"fmt"
	"net/http"

	"github.com/gitsense/gitsense/backend/internal/graph"
	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AdminHandler 处理管理接口
type AdminHandler struct {
	collector      *service.CollectorService
	embedding      *service.EmbeddingService
	worker         *service.EmbeddingWorker
	graphBuilder   *graph.BuilderWorker
	hasGitHubToken bool
}

// NewAdminHandler 创建新的 AdminHandler
func NewAdminHandler(collector *service.CollectorService, embedding *service.EmbeddingService, worker *service.EmbeddingWorker, graphBuilder *graph.BuilderWorker, hasGitHubToken bool) *AdminHandler {
	return &AdminHandler{
		collector:      collector,
		embedding:      embedding,
		worker:         worker,
		graphBuilder:   graphBuilder,
		hasGitHubToken: hasGitHubToken,
	}
}

// SeedRequest 批量导入请求
type SeedRequest struct {
	Repos []string `json:"repos" binding:"required"`
}

// SeedResponse 批量导入响应
type SeedResponse struct {
	Imported int      `json:"imported"`
	Failed   []string `json:"failed,omitempty"`
	Warning  string   `json:"warning,omitempty"`
}

// Seed 处理 POST /admin/seed
// 批量导入 GitHub 仓库
func (h *AdminHandler) Seed(c *gin.Context) {
	var req SeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "repos array is required",
			},
		})
		return
	}

	if len(req.Repos) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "repos array must not be empty",
			},
		})
		return
	}

	// GitHub Token 限流策略
	maxRepos := 100
	var warning string
	if !h.hasGitHubToken {
		maxRepos = 10
		warning = "GITHUB_TOKEN not configured. Limited to 10 repos per seed. Configure GITHUB_TOKEN for higher limits."
	}

	if len(req.Repos) > maxRepos {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: fmt.Sprintf("repos array must not exceed %d items without GITHUB_TOKEN", maxRepos),
			},
		})
		return
	}

	ctx := c.Request.Context()
	succeeded, failed := h.collector.BatchFetchAndStore(ctx, req.Repos)

	// 通过 Worker channel 立即触发 embedding（混合模型）
	for _, repo := range succeeded {
		if repo.EmbeddingStatus == model.EmbeddingPending {
			h.worker.Submit(repo.FullName)
		}
	}

	resp := SeedResponse{
		Imported: len(succeeded),
		Failed:   failed,
	}
	if warning != "" {
		resp.Warning = warning
	}

	c.JSON(http.StatusOK, resp)
}

// BuildGraph 处理 POST /admin/build-graph
// 手动触发 Graph 构建（全量或增量）
func (h *AdminHandler) BuildGraph(c *gin.Context) {
	var req graph.BuildGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认增量构建
		req = graph.BuildGraphRequest{FullRebuild: false}
	}

	ctx := c.Request.Context()
	result, err := h.graphBuilder.BuildGraph(ctx, req.FullRebuild)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "BUILD_GRAPH_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
