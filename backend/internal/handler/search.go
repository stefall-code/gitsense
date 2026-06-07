package handler

import (
	"net/http"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// SearchHandler 处理搜索请求
type SearchHandler struct {
	collector      *service.CollectorService
	embedding      *service.EmbeddingService
	recommendation *service.RecommendationService
	ecosystem      *service.EcosystemService
}

// NewSearchHandler 创建新的 SearchHandler
func NewSearchHandler(
	collector *service.CollectorService,
	embedding *service.EmbeddingService,
	recommendation *service.RecommendationService,
	ecosystem *service.EcosystemService,
) *SearchHandler {
	return &SearchHandler{
		collector:      collector,
		embedding:      embedding,
		recommendation: recommendation,
		ecosystem:      ecosystem,
	}
}

// Search 处理 POST /api/v1/search
func (h *SearchHandler) Search(c *gin.Context) {
	var req model.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "INVALID_REQUEST", Message: "repo_url is required"},
		})
		return
	}

	owner, name, err := service.ParseRepoInput(req.RepoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "INVALID_URL", Message: err.Error()},
		})
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	debug := c.Query("debug") == "true"
	strategy := c.Query("strategy") // "v1" or "v2" (default)

	ctx := c.Request.Context()

	// 1. 采集仓库信息并入库
	repo, err := h.collector.FetchAndStore(ctx, owner, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "COLLECT_FAILED", Message: err.Error()},
		})
		return
	}

	// 2. 如果 embedding 未就绪，异步触发
	if repo.EmbeddingStatus != model.EmbeddingDone {
		go func() {
			_ = h.embedding.GenerateForRepo(ctx, repo.FullName)
		}()
		c.JSON(http.StatusAccepted, model.SearchResponse{
			Repository: repo,
		})
		return
	}

	// 3. 获取推荐
	similarRepos, err := h.recommendation.GetRecommendations(ctx, repo.FullName, limit, debug, strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "RECOMMENDATION_FAILED", Message: err.Error()},
		})
		return
	}

	// 4. 获取生态
	eco, _ := h.ecosystem.GetEcosystem(ctx, repo.FullName)

	resp := model.SearchResponse{
		Repository:          repo,
		SimilarRepositories: similarRepos,
		Ecosystem:          eco,
	}

	if debug {
		resp.Debug = true
		if strategy == "v1" {
			resp.Strategy = "weighted_v1"
		} else {
			resp.Strategy = "weighted_v2"
		}
	}

	c.JSON(http.StatusOK, resp)
}
