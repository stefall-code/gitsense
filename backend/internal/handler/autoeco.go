package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/gitsense/gitsense/backend/internal/autoeco"
	"github.com/gin-gonic/gin"
)

// AutoEcosystemHandler 自动生态发现 Handler (G68 Shadow Mode)
type AutoEcosystemHandler struct {
	builder *autoeco.Builder
}

// NewAutoEcosystemHandler 创建 Handler
func NewAutoEcosystemHandler(builder *autoeco.Builder) *AutoEcosystemHandler {
	return &AutoEcosystemHandler{builder: builder}
}

// BuildAutoEcosystems POST /admin/build-auto-ecosystems
// Query params: ?hub_penalty=true (optional, Task B)
func (h *AutoEcosystemHandler) BuildAutoEcosystems(c *gin.Context) {
	hubPenalty := c.Query("hub_penalty") == "true"

	// Run in background context so build continues even if HTTP connection drops
	bgCtx := context.Background()
	go func() {
		var result *autoeco.BuildResult
		var err error
		if hubPenalty {
			result, err = h.builder.BuildWithPenalty(bgCtx)
		} else {
			result, err = h.builder.Build(bgCtx)
		}
		if err != nil {
			log.Printf("[admin] background build-auto-ecosystems error (hub_penalty=%v): %v", hubPenalty, err)
		} else {
			log.Printf("[admin] background build-auto-ecosystems completed (hub_penalty=%v): coverage=%.2f%% clusters=%d emerging=%d",
				hubPenalty, result.CoveragePct, result.ClustersFound, result.EmergingCount)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":      "Auto ecosystem build started in background",
		"hub_penalty":  hubPenalty,
	})
}

// GetAutoEcosystemReport GET /admin/auto-ecosystems/report
func (h *AutoEcosystemHandler) GetAutoEcosystemReport(c *gin.Context) {
	report, err := h.builder.GetReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NO_REPORT",
				"message": err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, report)
}

// BenchmarkHubPenalty GET /admin/auto-ecosystems/benchmark
// Task B: Hub Topic Penalty 基准测试
func (h *AutoEcosystemHandler) BenchmarkHubPenalty(c *gin.Context) {
	benchmark, err := h.builder.BenchmarkHubPenalty(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "BENCHMARK_FAILED",
				"message": err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, benchmark)
}
