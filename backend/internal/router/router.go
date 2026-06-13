package router

import (
	"github.com/gitsense/gitsense/backend/internal/audit"
	"github.com/gitsense/gitsense/backend/internal/handler"
	"github.com/gitsense/gitsense/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Setup 注册所有路由
func Setup(
	engine *gin.Engine,
	searchHandler *handler.SearchHandler,
	repoHandler *handler.RepoHandler,
	recHandler *handler.RecommendationHandler,
	ecoHandler *handler.EcosystemHandler,
	healthHandler *handler.HealthHandler,
	adminHandler *handler.AdminHandler,
	searchSimilarHandler *handler.SearchSimilarHandler,
	graphHandler *handler.GraphHandler,
	trendHandler *handler.TrendHandler,
	bootstrapHandler *handler.BootstrapHandler,
	auditHandler *audit.Handler,
	discoveryHandler *handler.DiscoveryHandler,
	autoEcoHandler *handler.AutoEcosystemHandler,
	analyticsHandler *handler.AnalyticsHandler,
	cacheStatsHandler *handler.CacheStatsHandler,
	adminToken string,
) {
	api := engine.Group("/api/v1")
	{
		api.POST("/search", searchHandler.Search)
		api.GET("/health", healthHandler.Health)
		api.GET("/search/similar", searchSimilarHandler.SearchSimilar)

		repos := api.Group("/repos")
		{
			repos.GET("/:owner/:name", repoHandler.GetRepo)
			repos.GET("/:owner/:name/recommendations", recHandler.GetRecommendations)
			repos.GET("/:owner/:name/ecosystem", ecoHandler.GetEcosystem)
		}

		// Discovery API
		api.GET("/discovery/:owner/:repo", discoveryHandler.Discover)
		api.GET("/ecosystems", discoveryHandler.ListEcosystems)
		api.GET("/ecosystem/:name", discoveryHandler.GetEcosystem)
		api.GET("/ecosystem/:name/trending", discoveryHandler.GetTrending)

		// Graph API
		graphGroup := api.Group("/graph")
		{
			graphGroup.GET("/repo/:owner/:repo", graphHandler.GetRepoGraph)
			graphGroup.GET("/ecosystem/:name", graphHandler.GetEcosystemGraph)
			graphGroup.GET("/path", graphHandler.FindPaths)
			graphGroup.GET("/explanation", graphHandler.GetGraphExplanation)
		}

		// Trend API
		trendGroup := api.Group("/trends")
		{
			trendGroup.GET("/topics", trendHandler.GetTopicTrends)
			trendGroup.GET("/ecosystems", trendHandler.GetEcosystemTrends)
			trendGroup.GET("/overview", trendHandler.GetOverview)
		}

		// Analytics API
		api.POST("/analytics/events", analyticsHandler.CollectEvents)
		api.GET("/analytics/stats", analyticsHandler.GetStats)
	}

	// 管理接口 — 需要 ADMIN_TOKEN 鉴权
	admin := engine.Group("/admin", middleware.AdminAuth(adminToken))
	{
		admin.POST("/seed", adminHandler.Seed)
		admin.POST("/build-graph", adminHandler.BuildGraph)

		// Graph Metrics
		admin.GET("/graph/metrics", graphHandler.GetGraphMetrics)

		// Bootstrap API
		admin.POST("/bootstrap/start", bootstrapHandler.Start)
		admin.POST("/bootstrap/stop", bootstrapHandler.Stop)
		admin.GET("/bootstrap/status", bootstrapHandler.Status)
		admin.POST("/bootstrap/resume", bootstrapHandler.Resume)

		// Audit API
		admin.GET("/audit/graph", auditHandler.GraphAudit)
		admin.POST("/audit/ablation", auditHandler.RankingAblation)

		// Auto Ecosystem API (G68 Shadow Mode — admin only)
		admin.POST("/build-auto-ecosystems", autoEcoHandler.BuildAutoEcosystems)
		admin.GET("/auto-ecosystems/report", autoEcoHandler.GetAutoEcosystemReport)
		admin.GET("/auto-ecosystems/benchmark", autoEcoHandler.BenchmarkHubPenalty)

		// Cache Stats API
		admin.GET("/cache/stats", cacheStatsHandler.GetStats)
	}
}
