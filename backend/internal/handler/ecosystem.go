package handler

import (
	"net/http"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// EcosystemHandler 处理技术生态相关请求
type EcosystemHandler struct {
	ecosystem *service.EcosystemService
}

// NewEcosystemHandler 创建新的 EcosystemHandler
func NewEcosystemHandler(ecosystem *service.EcosystemService) *EcosystemHandler {
	return &EcosystemHandler{ecosystem: ecosystem}
}

// GetEcosystem 处理 GET /api/v1/repos/:owner/:name/ecosystem
func (h *EcosystemHandler) GetEcosystem(c *gin.Context) {
	owner := c.Param("owner")
	name := c.Param("name")
	fullName := owner + "/" + name

	eco, err := h.ecosystem.GetEcosystem(c.Request.Context(), fullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "ECOSYSTEM_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	if eco == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "ECOSYSTEM_NOT_FOUND",
				Message: "ecosystem not found for: " + fullName,
			},
		})
		return
	}

	c.JSON(http.StatusOK, eco)
}
