package audit

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler audit HTTP handler
type Handler struct {
	store *Store
}

// NewHandler 创建 audit handler
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// GraphAudit GET /admin/audit/graph
func (h *Handler) GraphAudit(c *gin.Context) {
	result, err := h.store.RunGraphAudit(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// RankingAblation POST /admin/audit/ablation
func (h *Handler) RankingAblation(c *gin.Context) {
	var req struct {
		Queries []string `json:"queries"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认查询列表
		req.Queries = []string{
			"langchain-ai/langgraph",
			"crewAIInc/crewAI",
			"langchain-ai/langchain",
			"gin-gonic/gin",
			"apache/airflow",
			"apache/kafka",
		}
	}

	result, err := h.store.RunRankingAblation(c.Request.Context(), req.Queries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ensure json import is used
var _ = json.Marshal
