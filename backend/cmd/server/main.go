package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gitsense/gitsense/backend/internal/audit"
	"github.com/gitsense/gitsense/backend/internal/autoeco"
	"github.com/gitsense/gitsense/backend/internal/bootstrap"
	"github.com/gitsense/gitsense/backend/internal/cache"
	"github.com/gitsense/gitsense/backend/internal/config"
	"github.com/gitsense/gitsense/backend/internal/discovery"
	"github.com/gitsense/gitsense/backend/internal/embedding"
	"github.com/gitsense/gitsense/backend/internal/github"
	"github.com/gitsense/gitsense/backend/internal/graph"
	"github.com/gitsense/gitsense/backend/internal/handler"
	"github.com/gitsense/gitsense/backend/internal/repository"
	"github.com/gitsense/gitsense/backend/internal/router"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/gitsense/gitsense/backend/internal/trend"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.Server.Mode)

	ctx := context.Background()
	store, err := repository.NewRepoStore(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer store.Close()

	cacheClient := cache.NewClient(cfg.Redis)
	defer cacheClient.Close()

	// 保留旧 CacheStore 以兼容 recommendation/ecosystem service
	cacheStore, err := repository.NewCacheStore(ctx, cfg.Redis, cfg.Cache)
	if err != nil {
		log.Printf("warning: redis not available: %v", err)
	}

	ghClient := github.NewClient(cfg.GitHub.Token)

	// --- 初始化 Embedding Provider ---
	var provider service.EmbeddingProvider
	switch cfg.Embedding.Provider {
	case "openai":
		if cfg.Embedding.APIKey == "" {
			log.Fatalf("OPENAI_API_KEY required for openai provider")
		}
		provider = embedding.NewOpenAIProvider(cfg.Embedding.APIKey, cfg.Embedding.Model, cfg.Embedding.Dimensions)
		log.Printf("using OpenAI embedding provider (model=%s, dims=%d)", cfg.Embedding.Model, cfg.Embedding.Dimensions)
	case "local", "http":
		baseURL := fmt.Sprintf("http://%s:%s", cfg.Embedding.Host, cfg.Embedding.Port)
		provider = embedding.NewHTTPProvider(baseURL, cfg.Embedding.Dimensions)
		log.Printf("using HTTP embedding provider (url=%s, dims=%d)", baseURL, cfg.Embedding.Dimensions)
	case "mock":
		provider = embedding.NewMockProvider(cfg.Embedding.Dimensions)
		log.Printf("using Mock embedding provider (dims=%d)", cfg.Embedding.Dimensions)
	default:
		log.Printf("warning: unknown embedding provider '%s', using mock", cfg.Embedding.Provider)
		provider = embedding.NewMockProvider(cfg.Embedding.Dimensions)
	}

	// --- 初始化服务 ---
	collectorSvc := service.NewCollectorService(store, ghClient)
	embeddingSvc := service.NewEmbeddingService(store, provider)

	strategyV1 := service.NewWeightedSimilarityStrategy()
	strategyV2 := service.NewWeightedSimilarityStrategyV2()
	classifier := service.NewEcosystemClassifier(nil)

	recommendationSvc := service.NewRecommendationService(store, cacheStore, strategyV1, strategyV2, classifier)
	ecosystemSvc := service.NewEcosystemService(store, cacheStore, classifier)

	// Embedding Worker（混合模型）
	worker := service.NewEmbeddingWorker(embeddingSvc, 30*time.Second, cfg.Embedding.BatchSize)
	go worker.Start(ctx)
	defer worker.Stop()

	// --- 初始化 Graph 组件 ---
	graphStore := graph.NewStore(store.Pool())
	graphBuilder := graph.NewBuilderWorker(graphStore, classifier, 6*time.Hour)
	graphService := graph.NewService(graphStore, classifier)

	// Graph Worker 定时构建
	go graphBuilder.Start(ctx)
	defer graphBuilder.Stop()

	// --- 初始化 Trend 组件 ---
	trendProvider := trend.NewCachedTrendProvider(cacheClient, store.Pool(), time.Duration(cfg.Trend.CacheTTL)*time.Second)
	trendService := trend.NewService(trendProvider)
	trendWorker := trend.NewWorker(trendService, 6*time.Hour)

	// Trend Worker 定时刷新缓存
	go trendWorker.Start(ctx)
	defer trendWorker.Stop()

	// V3 Strategy（Trend-aware，feature flag 控制）
	trendAdapter := trend.NewTrendScoreProviderAdapter(trendProvider)
	strategyV3 := service.NewWeightedSimilarityStrategyV3(trendAdapter, cfg.Trend.EnableRanking)
	log.Printf("trend ranking: enabled=%v", cfg.Trend.EnableRanking)

	// 更新 RecommendationService 注入 V3
	recommendationSvc.SetStrategyV3(strategyV3)

	// --- 初始化 Bootstrap 组件 ---
	bootstrapStore := bootstrap.NewStore(store.Pool())
	bootstrapSvc := bootstrap.NewService(bootstrapStore, ghClient, collectorSvc, worker, cacheClient, bootstrap.BootstrapConfig{
		MinStars:    cfg.Bootstrap.MinStars,
		MinForks:    cfg.Bootstrap.MinForks,
		ActiveYears: cfg.Bootstrap.ActiveYears,
		MaxDepth:    cfg.Bootstrap.MaxDepth,
		Workers:     cfg.Bootstrap.Workers,
		SeedConfig:  bootstrap.DefaultSeedConfig(),
	})
	log.Printf("bootstrap config: min_stars=%d, max_depth=%d, workers=%d", cfg.Bootstrap.MinStars, cfg.Bootstrap.MaxDepth, cfg.Bootstrap.Workers)

	// 启动时自动 resume 未完成的 bootstrap job
	if err := bootstrapSvc.Resume(context.Background()); err != nil {
		log.Printf("[bootstrap] auto-resume: %v", err)
	}

	// --- 初始化 Handler ---
	searchHandler := handler.NewSearchHandler(collectorSvc, embeddingSvc, recommendationSvc, ecosystemSvc)
	repoHandler := handler.NewRepoHandler(store)
	recHandler := handler.NewRecommendationHandler(recommendationSvc)
	ecoHandler := handler.NewEcosystemHandler(ecosystemSvc)
	healthHandler := handler.NewHealthHandler(cacheClient)
	adminHandler := handler.NewAdminHandler(collectorSvc, embeddingSvc, worker, graphBuilder, cfg.GitHub.Token != "")
	searchSimilarHandler := handler.NewSearchSimilarHandler(embeddingSvc, recommendationSvc, classifier)
	graphHandler := handler.NewGraphHandler(graphService, graphStore)
	trendHandler := handler.NewTrendHandler(trendService)
	bootstrapHandler := handler.NewBootstrapHandler(bootstrapSvc)

	// --- 初始化 Audit 组件 ---
	auditStore := audit.NewStore(store.Pool())
	auditHandler := audit.NewHandler(auditStore)

	// --- 初始化 Discovery 组件 ---
	discoverySvc := discovery.NewService(store, graphStore, classifier, recommendationSvc, trendService)
	discoveryHandler := handler.NewDiscoveryHandler(discoverySvc, cacheClient)

	// --- 初始化 Auto Ecosystem 组件 (Phase 11, G68 Shadow Mode) ---
	autoEcoBuilder := autoeco.NewBuilder(store.Pool(), classifier)
	autoEcoHandler := handler.NewAutoEcosystemHandler(autoEcoBuilder)

	// --- 初始化 Analytics 组件 ---
	analyticsHandler := handler.NewAnalyticsHandler(store.Pool())

	// --- 初始化 Cache Stats Handler ---
	cacheStatsHandler := handler.NewCacheStatsHandler(cacheClient)

	healthHandler.SetDBHealthy(store.Ping(ctx) == nil)

	engine := gin.Default()
	router.Setup(engine, searchHandler, repoHandler, recHandler, ecoHandler, healthHandler, adminHandler, searchSimilarHandler, graphHandler, trendHandler, bootstrapHandler, auditHandler, discoveryHandler, autoEcoHandler, analyticsHandler, cacheStatsHandler, cfg.Server.AdminToken)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("GitSense server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}
	log.Println("server exited")
}
