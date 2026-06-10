package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gitsense/gitsense/backend/internal/github"
	"github.com/gitsense/gitsense/backend/internal/service"
)

// Service Bootstrap 服务
type Service struct {
	store     *Store
	client    *github.Client
	collector *service.CollectorService
	worker    *service.EmbeddingWorker

	// 配置
	config BootstrapConfig

	// 状态
	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	job     *BootstrapJob
}

// BootstrapConfig Bootstrap 配置
type BootstrapConfig struct {
	MinStars       int
	MinForks       int
	ActiveYears    int
	MaxDepth       int
	Workers        int
	SeedConfig     SeedConfig
}

// DefaultBootstrapConfig 默认配置
func DefaultBootstrapConfig() BootstrapConfig {
	return BootstrapConfig{
		MinStars:    100,
		MinForks:    10,
		ActiveYears: 2,
		MaxDepth:    2,
		Workers:     3,
		SeedConfig:  DefaultSeedConfig(),
	}
}

// NewService 创建 Bootstrap Service
func NewService(
	store *Store,
	client *github.Client,
	collector *service.CollectorService,
	worker *service.EmbeddingWorker,
	config BootstrapConfig,
) *Service {
	return &Service{
		store:     store,
		client:    client,
		collector: collector,
		worker:    worker,
		config:    config,
	}
}

// Start 启动 Bootstrap
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// 创建 job
	job, err := s.store.CreateJob(ctx)
	if err != nil {
		return err
	}
	s.job = job

	// Phase 1: 种子入队
	if err := s.seedPhase1(ctx, job.ID); err != nil {
		log.Printf("[bootstrap] seed phase 1 error: %v", err)
	}

	// 启动 worker pool
	bsCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true

	for i := 0; i < s.config.Workers; i++ {
		go s.workerLoop(bsCtx, i, job.ID)
	}

	log.Printf("[bootstrap] started with %d workers, job_id=%d", s.config.Workers, job.ID)
	return nil
}

// Stop 停止 Bootstrap
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cancel()
	s.running = false

	if s.job != nil {
		s.job.Status = JobPaused
		now := time.Now()
		s.job.UpdatedAt = &now
		ctx := context.Background()
		s.store.UpdateJob(ctx, s.job)
	}

	log.Printf("[bootstrap] stopped")
}

// Resume 恢复 Bootstrap
func (s *Service) Resume(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// 获取最新 job
	job, err := s.store.GetLatestJob(ctx)
	if job == nil {
		return nil
	}
	// 容器重启后 running 状态的 job 实际没有 worker 在跑，需要恢复
	if job.Status == JobRunning {
		log.Printf("[bootstrap] found job %d in running state without active worker, resuming", job.ID)
	} else if job.Status != JobPaused && job.Status != JobFailed {
		return nil
	}

	// 重置 processing → pending
	resetCount, _ := s.store.ResetProcessing(ctx, job.ID)
	log.Printf("[bootstrap] resumed job %d, reset %d processing items", job.ID, resetCount)

	job.Status = JobRunning
	now := time.Now()
	job.UpdatedAt = &now
	s.store.UpdateJob(ctx, job)
	s.job = job

	bsCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true

	for i := 0; i < s.config.Workers; i++ {
		go s.workerLoop(bsCtx, i, job.ID)
	}

	log.Printf("[bootstrap] resumed job %d with %d workers", job.ID, s.config.Workers)
	return err
}

// GetStatus 获取状态
func (s *Service) GetStatus(ctx context.Context) *BootstrapStatus {
	status := &BootstrapStatus{}

	// Job 状态
	job, _ := s.store.GetLatestJob(ctx)
	status.Job = job

	// GitHub 额度
	coreRem, searchRem, coreReset, searchReset := s.client.GetRateLimitInfo()
	status.GitHub = GitHubRateLimit{
		CoreRemaining:   coreRem,
		SearchRemaining: searchRem,
	}
	if !coreReset.IsZero() {
		status.GitHub.CoreResetAt = &coreReset
	}
	if !searchReset.IsZero() {
		status.GitHub.SearchResetAt = &searchReset
	}

	// 数据集统计
	stats, _ := s.store.GetDatasetStats(ctx)
	if stats != nil {
		status.Dataset = *stats
	}

	return status
}

// IsRunning 是否运行中
func (s *Service) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// --- 核心逻辑 ---

// seedPhase1 第一层种子：awesome-list README 提取
func (s *Service) seedPhase1(ctx context.Context, jobID int) error {
	for _, repo := range s.config.SeedConfig.AwesomeRepos {
		// 种子仓库入队（depth=0）
		if err := s.store.Enqueue(ctx, jobID, repo, SourceAwesome, "", 0); err != nil {
			log.Printf("[bootstrap] enqueue seed %s: %v", repo, err)
		}
	}
	log.Printf("[bootstrap] seeded %d awesome repos", len(s.config.SeedConfig.AwesomeRepos))
	return nil
}

// workerLoop Worker 主循环
func (s *Service) workerLoop(ctx context.Context, workerID, jobID int) {
	log.Printf("[bootstrap-worker-%d] started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[bootstrap-worker-%d] stopped", workerID)
			return
		default:
		}

		// 取下一个任务
		item, err := s.store.Dequeue(ctx, jobID)
		if item == nil {
			// 队列空，等待
			time.Sleep(5 * time.Second)
			continue
		}
		if err != nil {
			log.Printf("[bootstrap-worker-%d] dequeue error: %v", workerID, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// 处理
		success := s.processItem(ctx, item, jobID)

		// 更新状态
		if success {
			s.store.MarkDone(ctx, item.ID)
		} else {
			s.store.MarkFailed(ctx, item.ID)
		}
		s.store.IncrementJobCounts(ctx, jobID, success)
		s.store.UpdateQueueSize(ctx, jobID)
	}
}

// processItem 处理单个队列项
func (s *Service) processItem(ctx context.Context, item *QueueItem, jobID int) bool {
	parts := strings.Split(item.RepoFullName, "/")
	if len(parts) != 2 {
		log.Printf("[bootstrap] invalid repo name: %s", item.RepoFullName)
		return false
	}
	owner, repo := parts[0], parts[1]

	// 1. 检查是否已存在
	exists, _ := s.store.IsRepoExists(ctx, item.RepoFullName)
	if exists {
		log.Printf("[bootstrap] skip existing: %s", item.RepoFullName)
		return true
	}

	// 2. 获取 repo 信息
	repoInfo, err := s.client.FetchRepository(ctx, owner, repo)
	if err != nil {
		log.Printf("[bootstrap] fetch %s error: %v", item.RepoFullName, err)
		return false
	}

	// 3. 质量过滤
	if !s.passQualityFilter(repoInfo) {
		log.Printf("[bootstrap] quality filter rejected: %s (stars=%d)", item.RepoFullName, repoInfo.Stars)
		return false
	}

	// 4. 采集入库
	_, err = s.collector.FetchAndStore(ctx, owner, repo)
	if err != nil {
		log.Printf("[bootstrap] collect %s error: %v", item.RepoFullName, err)
		return false
	}

	// 5. 触发 embedding
	if s.worker != nil {
		s.worker.Submit(item.RepoFullName)
	}

	// 6. 扩散（depth < maxDepth）
	if item.Depth < s.config.MaxDepth {
		s.expandFromRepo(ctx, item, jobID)
	}

	log.Printf("[bootstrap] processed: %s (depth=%d, source=%s)", item.RepoFullName, item.Depth, item.SourceType)
	return true
}

// passQualityFilter 数据质量过滤
func (s *Service) passQualityFilter(info *github.RepoInfo) bool {
	if info.Stars < s.config.MinStars {
		return false
	}
	if info.Language == "" {
		return false
	}

	// 活跃度检查
	if info.UpdatedAt != "" {
		t, err := time.Parse(time.RFC3339, info.UpdatedAt)
		if err == nil {
			cutoff := time.Now().AddDate(-s.config.ActiveYears, 0, 0)
			if t.Before(cutoff) {
				return false
			}
		}
	}

	return true
}

// expandFromRepo 从已采集 repo 扩散发现新 repo
func (s *Service) expandFromRepo(ctx context.Context, item *QueueItem, jobID int) {
	parts := strings.Split(item.RepoFullName, "/")
	if len(parts) != 2 {
		return
	}
	owner, repo := parts[0], parts[1]

	// Depth 0→1: 从 awesome-list README 提取
	if item.Depth == 0 && isAwesomeRepo(item.RepoFullName) {
		s.expandFromReadme(ctx, owner, repo, item.RepoFullName, jobID)
	}

	// Depth 1→2: Topic 扩散
	if item.Depth >= 1 {
		s.expandFromTopics(ctx, owner, repo, item.RepoFullName, jobID, item.Depth+1)
	}
}

// expandFromReadme 从 README 提取 GitHub repo URL
func (s *Service) expandFromReadme(ctx context.Context, owner, repo, discoveredFrom string, jobID int) {
	readme, err := s.client.FetchReadme(ctx, owner, repo)
	if err != nil || readme == "" {
		return
	}

	repos := extractGitHubRepos(readme)
	inserted := 0
	for _, r := range repos {
		if isAwesomeRepo(r) {
			continue // 跳过 awesome-list 自身
		}
		if err := s.store.Enqueue(ctx, jobID, r, SourceAwesome, discoveredFrom, 1); err == nil {
			inserted++
		}
	}
	log.Printf("[bootstrap] readme expansion: %s → %d new repos", discoveredFrom, inserted)
}

// expandFromTopics 通过 GitHub Search API 扩散
func (s *Service) expandFromTopics(ctx context.Context, owner, repo, discoveredFrom string, jobID, depth int) {
	// 获取 repo 的 topics
	topics, err := s.client.FetchTopics(ctx, owner, repo)
	if err != nil || len(topics) == 0 {
		return
	}

	inserted := 0
	for _, topic := range topics {
		query := url.QueryEscape(fmt.Sprintf("topic:%s stars:>%d", topic, s.config.MinStars))
		results, err := s.client.SearchRepositories(ctx, query, 30)
		if err != nil {
			log.Printf("[bootstrap] search topic:%s error: %v", topic, err)
			continue
		}

		for _, r := range results {
			if isAwesomeRepo(r.FullName) {
				continue
			}
			if err := s.store.Enqueue(ctx, jobID, r.FullName, SourceTopic, discoveredFrom, depth); err == nil {
				inserted++
			}
		}

		// Search API 限流：每请求间隔 2s
		time.Sleep(2 * time.Second)
	}
	log.Printf("[bootstrap] topic expansion: %s → %d new repos (depth=%d)", discoveredFrom, inserted, depth)
}

// --- 辅助函数 ---

// githubRepoRegex 匹配 GitHub repo URL
var githubRepoRegex = regexp.MustCompile(`https?://github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)`)

// extractGitHubRepos 从文本中提取 GitHub repo URL 并归一化
func extractGitHubRepos(text string) []string {
	matches := githubRepoRegex.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var repos []string

	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		owner := m[1]
		repo := m[2]

		// 过滤非 repo 路径
		if isNonRepoPath(repo) {
			continue
		}

		fullName := owner + "/" + repo
		if !seen[fullName] {
			seen[fullName] = true
			repos = append(repos, fullName)
		}
	}

	return repos
}

// isNonRepoPath 判断是否为非 repo 路径（issues, pulls, wiki 等）
func isNonRepoPath(path string) bool {
	nonRepoPaths := []string{"issues", "pulls", "wiki", "actions", "security", "releases",
		"blob", "tree", "commit", "compare", "archive", "tags", "milestone",
		"settings", "notifications", "marketplace", "apps", "orgs", "users",
		"explore", "topics", "trending", "collections", "events", "sponsors",
		"features", "pricing", "login", "signup", "join", "new", "create",
		"organizations", "dashboard", "account", "repositories", "stars",
		"watching", "gist"}

	lower := strings.ToLower(path)
	for _, p := range nonRepoPaths {
		if lower == p {
			return true
		}
	}
	return false
}

// isAwesomeRepo 判断是否为 awesome-list 仓库
func isAwesomeRepo(fullName string) bool {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return false
	}
	repo := strings.ToLower(parts[1])
	return strings.HasPrefix(repo, "awesome-") || strings.HasPrefix(repo, "awesome_") ||
		strings.Contains(repo, "awesome")
}
