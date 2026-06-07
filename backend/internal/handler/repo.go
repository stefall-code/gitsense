package handler

import (
	"net/http"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

// RepoHandler 处理仓库相关请求
type RepoHandler struct {
	store *repository.RepoStore
}

// NewRepoHandler 创建新的 RepoHandler
func NewRepoHandler(store *repository.RepoStore) *RepoHandler {
	return &RepoHandler{store: store}
}

// GetRepo 处理 GET /api/v1/repos/:owner/:name
func (h *RepoHandler) GetRepo(c *gin.Context) {
	owner := c.Param("owner")
	name := c.Param("name")
	fullName := owner + "/" + name

	repo, err := h.store.GetByFullName(c.Request.Context(), fullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	if repo == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "REPO_NOT_FOUND",
				Message: "repository not found: " + fullName,
			},
		})
		return
	}

	c.JSON(http.StatusOK, repo)
}
