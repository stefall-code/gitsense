package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsHandler handles analytics event collection
type AnalyticsHandler struct {
	pool *pgxpool.Pool
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(pool *pgxpool.Pool) *AnalyticsHandler {
	return &AnalyticsHandler{pool: pool}
}

type analyticsEvent struct {
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  string                 `json:"timestamp"`
}

type analyticsPayload struct {
	SessionID string            `json:"session_id"`
	Events    []analyticsEvent  `json:"events"`
}

// CollectEvents receives analytics events and logs them
// MVP: Just log events. No database storage yet.
func (h *AnalyticsHandler) CollectEvents(c *gin.Context) {
	var payload analyticsPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	for _, event := range payload.Events {
		log.Printf("[analytics] session=%s event=%s props=%v ts=%s",
			payload.SessionID, event.Event, event.Properties, event.Timestamp)
	}

	c.JSON(http.StatusOK, gin.H{"received": len(payload.Events)})
}

// GetStats returns basic analytics stats
func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	// MVP: Return placeholder stats
	c.JSON(http.StatusOK, gin.H{
		"total_repos":      0,
		"total_searches":   0,
		"total_clicks":     0,
		"last_updated":     time.Now().Format(time.RFC3339),
		"note":             "Analytics collection is in MVP mode (logging only)",
	})
}
