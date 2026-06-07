package handler

import (
	"context"
	"net/http"

	"github.com/gitsense/gitsense/backend/internal/bootstrap"
	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// BootstrapHandler 处理 Bootstrap Admin API
type BootstrapHandler struct {
	service *bootstrap.Service
}

// NewBootstrapHandler 创建 BootstrapHandler
func NewBootstrapHandler(service *bootstrap.Service) *BootstrapHandler {
	return &BootstrapHandler{service: service}
}

// Start 处理 POST /admin/bootstrap/start
func (h *BootstrapHandler) Start(c *gin.Context) {
	if h.service.IsRunning() {
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "ALREADY_RUNNING", Message: "Bootstrap is already running"},
		})
		return
	}

	if err := h.service.Start(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "START_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bootstrap started"})
}

// Stop 处理 POST /admin/bootstrap/stop
func (h *BootstrapHandler) Stop(c *gin.Context) {
	if !h.service.IsRunning() {
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "NOT_RUNNING", Message: "Bootstrap is not running"},
		})
		return
	}

	h.service.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "Bootstrap stopped"})
}

// Status 处理 GET /admin/bootstrap/status
func (h *BootstrapHandler) Status(c *gin.Context) {
	status := h.service.GetStatus(c.Request.Context())
	c.JSON(http.StatusOK, status)
}

// Resume 处理 POST /admin/bootstrap/resume
func (h *BootstrapHandler) Resume(c *gin.Context) {
	if h.service.IsRunning() {
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "ALREADY_RUNNING", Message: "Bootstrap is already running"},
		})
		return
	}

	if err := h.service.Resume(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Code: "RESUME_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bootstrap resumed"})
}
