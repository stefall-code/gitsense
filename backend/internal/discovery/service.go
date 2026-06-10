package discovery

import (
	"context"
	"fmt"
	"sort"

	"github.com/gitsense/gitsense/backend/internal/graph"
	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/repository"
	"github.com/gitsense/gitsense/backend/internal/service"
	"github.com/gitsense/gitsense/backend/internal/trend"
)

// StackRepo 技术栈中的项目
type StackRepo struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
	Language    string `json:"language"`
	Trend       string `json:"trend"`
}

// Subcategory 子分类
type Subcategory struct {
	Name      string      `json:"name"`
	RepoCount int         `json:"repo_count"`
	TopRepos  []StackRepo `json:"top_repos"`
	Trending  []StackRepo `json:"trending"`
}

// TechStackTree 技术栈树
type TechStackTree struct {
	Ecosystem  string        `json:"ecosystem"`
	Categories []Subcategory `json:"categories"`
}

// EcosystemInfo 生态信息
type EcosystemInfo struct {
	Name        string `json:"name"`
	Subcategory string `json:"subcategory"`
	RepoCount   int    `json:"repo_count"`
	Trend       string `json:"trend"`
}

// RepoSummary 仓库摘要
type RepoSummary struct {
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	Stars       int      `json:"stars"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
}

// DiscoveryResponse 发现端点响应
type DiscoveryResponse struct {
	Repo            RepoSummary               `json:"repo"`
	Ecosystem       EcosystemInfo             `json:"ecosystem"`
	Stack           TechStackTree             `json:"stack"`
	Recommendations []model.SimilarRepository `json:"recommendations"`
}

// EcosystemSummary 生态列表项
type EcosystemSummary struct {
	Name          string  `json:"name"`
	RepoCount     int     `json:"repo_count"`
	CategoryCount int     `json:"category_count"`
	Trend         string  `json:"trend"`
	TrendScore    float64 `json:"trend_score"`
}

// EcosystemsResponse 生态列表响应
type EcosystemsResponse struct {
	Ecosystems []EcosystemSummary `json:"ecosystems"`
}

// EcosystemDetail 生态详情响应
type EcosystemDetail struct {
	Name        string        `json:"name"`
	RepoCount   int           `json:"repo_count"`
	Trend       string        `json:"trend"`
	TrendScore  float64       `json:"trend_score"`
	GrowthRate  float64       `json:"growth_rate"`
	Categories  []Subcategory `json:"categories"`
	TopRepos    []StackRepo   `json:"top_repos"`
}

// TrendingRepo 趋势项目
type TrendingRepo struct {
	FullName    string  `json:"full_name"`
	Stars       int     `json:"stars"`
	Language    string  `json:"language"`
	Trend       string  `json:"trend"`
	TrendScore  float64 `json:"trend_score"`
	Subcategory string  `json:"subcategory"`
}

// TrendingResponse 趋势项目响应
type TrendingResponse struct {
	Ecosystem string        `json:"ecosystem"`
	Window    string        `json:"window"`
	Trending  []TrendingRepo `json:"trending"`
}

// Service 发现服务
type Service struct {
	repoStore  *repository.RepoStore
	graphStore *graph.Store
	classifier *service.EcosystemClassifier
	recService *service.RecommendationService
	trendSvc   *trend.Service
}

// NewService 创建发现服务
func NewService(
	repoStore *repository.RepoStore,
	graphStore *graph.Store,
	classifier *service.EcosystemClassifier,
	recService *service.RecommendationService,
	trendSvc *trend.Service,
) *Service {
	return &Service{
		repoStore:  repoStore,
		graphStore: graphStore,
		classifier: classifier,
		recService: recService,
		trendSvc:   trendSvc,
	}
}

// Discover 一站式发现
func (s *Service) Discover(ctx context.Context, fullName string, limit int) (*DiscoveryResponse, error) {
	// Step 1: 查询 repo
	repo, err := s.repoStore.GetByFullName(ctx, fullName)
	if err != nil {
		return nil, fmt.Errorf("get repo: %w", err)
	}
	if repo == nil {
		return nil, fmt.Errorf("repository not found: %s", fullName)
	}

	// Step 2: 解析 ecosystem + subcategory
	ecoName := s.classifier.Classify(repo.Topics, repo.Description)
	if ecoName == "" {
		ecoName = s.classifier.ClassifyWithFallback(repo.Topics, repo.Description)
	}
	// Language-based fallback for repos with no matching topics
	if ecoName == "Unknown Ecosystem" && repo.Language != "" {
		ecoName = s.classifier.ClassifyByLanguage(repo.Language)
	}
	subcategory := s.classifier.ClassifySubcategory(repo.Topics, ecoName)

	// Step 3: 获取生态 repo 数量
	ecoRepos, _ := s.graphStore.GetEcosystemRepos(ctx, ecoName, 500)
	repoCount := len(ecoRepos)

	// Step 4: 获取趋势
	trendScore := s.trendSvc.GetEcosystemTrendScore(ctx, ecoName, trend.Window7d)
	trendStatus := classifyTrend(trendScore)

	// Step 5: 构建技术栈树
	stack, err := s.buildStack(ctx, ecoName)
	if err != nil {
		stack = &TechStackTree{Ecosystem: ecoName}
	}

	// Step 6: 获取推荐
	recs, err := s.recService.GetRecommendations(ctx, fullName, limit, false, "")
	if err != nil {
		recs = nil
	}

	return &DiscoveryResponse{
		Repo: RepoSummary{
			FullName:    repo.FullName,
			Description: repo.Description,
			Stars:       repo.Stars,
			Language:    repo.Language,
			Topics:      repo.Topics,
		},
		Ecosystem: EcosystemInfo{
			Name:        ecoName,
			Subcategory: subcategory,
			RepoCount:   repoCount,
			Trend:       trendStatus,
		},
		Stack:           *stack,
		Recommendations: recs,
	}, nil
}

// ListEcosystems 列出所有生态
func (s *Service) ListEcosystems(ctx context.Context) (*EcosystemsResponse, error) {
	rules := s.classifier.GetAllRules()
	var ecosystems []EcosystemSummary

	for _, rule := range rules {
		ecoRepos, _ := s.graphStore.GetEcosystemRepos(ctx, rule.Name, 500)
		repoCount := len(ecoRepos)

		trendScore := s.trendSvc.GetEcosystemTrendScore(ctx, rule.Name, trend.Window7d)

		ecosystems = append(ecosystems, EcosystemSummary{
			Name:          rule.Name,
			RepoCount:     repoCount,
			CategoryCount: len(rule.Subcategories),
			Trend:         classifyTrend(trendScore),
			TrendScore:    trendScore,
		})
	}

	return &EcosystemsResponse{Ecosystems: ecosystems}, nil
}

// GetEcosystem 获取生态详情
func (s *Service) GetEcosystem(ctx context.Context, name string) (*EcosystemDetail, error) {
	rule := s.classifier.GetRule(name)
	if rule == nil {
		return nil, fmt.Errorf("ecosystem not found: %s", name)
	}

	ecoRepos, _ := s.graphStore.GetEcosystemRepos(ctx, name, 500)
	repoCount := len(ecoRepos)

	trendScore := s.trendSvc.GetEcosystemTrendScore(ctx, name, trend.Window7d)

	stack, err := s.buildStack(ctx, name)
	if err != nil {
		stack = &TechStackTree{Ecosystem: name}
	}

	// Top repos (全生态)
	topRepos := s.getTopReposInEcosystem(ctx, name, 10)

	return &EcosystemDetail{
		Name:       name,
		RepoCount:  repoCount,
		Trend:      classifyTrend(trendScore),
		TrendScore: trendScore,
		Categories: stack.Categories,
		TopRepos:   topRepos,
	}, nil
}

// GetTrending 获取生态内趋势项目
func (s *Service) GetTrending(ctx context.Context, name string, window string, limit int) (*TrendingResponse, error) {
	rule := s.classifier.GetRule(name)
	if rule == nil {
		return nil, fmt.Errorf("ecosystem not found: %s", name)
	}

	ecoRepos, _ := s.graphStore.GetEcosystemRepos(ctx, name, 500)
	if len(ecoRepos) == 0 {
		return &TrendingResponse{Ecosystem: name, Window: window}, nil
	}

	// 查询每个 repo 的趋势
	var trending []TrendingRepo
	for _, repoName := range ecoRepos {
		if len(trending) >= limit*3 { // over-fetch then sort
			break
		}
		repo, err := s.repoStore.GetByFullName(ctx, repoName)
		if err != nil || repo == nil {
			continue
		}

		// 用 repo 的 topics 计算趋势分数
		var maxTopicScore float64
		for _, t := range repo.Topics {
			score := s.trendSvc.GetTopicTrendScore(ctx, t, trend.TimeWindow(window))
			if score > maxTopicScore {
				maxTopicScore = score
			}
		}

		// 归一化到 [0, 1]
		normalizedScore := (maxTopicScore + 1) / 2
		if normalizedScore < 0.6 {
			continue // 只返回 rising
		}

		subcat := s.classifier.ClassifySubcategory(repo.Topics, name)
		trending = append(trending, TrendingRepo{
			FullName:    repo.FullName,
			Stars:       repo.Stars,
			Language:    repo.Language,
			Trend:       "rising",
			TrendScore:  normalizedScore,
			Subcategory: subcat,
		})
	}

	// 按 trend_score DESC 排序
	sort.Slice(trending, func(i, j int) bool {
		return trending[i].TrendScore > trending[j].TrendScore
	})

	if len(trending) > limit {
		trending = trending[:limit]
	}

	return &TrendingResponse{
		Ecosystem: name,
		Window:    window,
		Trending:  trending,
	}, nil
}

// buildStack 构建技术栈树
func (s *Service) buildStack(ctx context.Context, ecosystem string) (*TechStackTree, error) {
	rule := s.classifier.GetRule(ecosystem)
	if rule == nil {
		return nil, fmt.Errorf("ecosystem rule not found: %s", ecosystem)
	}

	ecoRepos, _ := s.graphStore.GetEcosystemRepos(ctx, ecosystem, 500)

	// 查询所有 repo 详情
	var allRepos []model.Repository
	for _, name := range ecoRepos {
		repo, err := s.repoStore.GetByFullName(ctx, name)
		if err != nil || repo == nil {
			continue
		}
		allRepos = append(allRepos, *repo)
	}

	// Fallback: 如果 ecosystem_map 覆盖不足，用 classifier 实时分类补充
	// 只在 ecoRepos 数量较少时触发
	if len(allRepos) < 50 {
		// 从 DB 获取所有 repos 的 topics，用 classifier 分类
		rows, err := s.repoStore.Pool().Query(ctx, `
			SELECT full_name, description, language, stars, topics
			FROM repositories
			WHERE stars >= 100
			ORDER BY stars DESC
			LIMIT 2000
		`)
		if err == nil {
			defer rows.Close()
			seen := make(map[string]bool, len(allRepos))
			for _, r := range allRepos {
				seen[r.FullName] = true
			}
			for rows.Next() {
				var r model.Repository
				if err := rows.Scan(&r.FullName, &r.Description, &r.Language, &r.Stars, &r.Topics); err != nil {
					continue
				}
				if seen[r.FullName] {
					continue
				}
				classified := s.classifier.Classify(r.Topics, r.Description)
				if classified == ecosystem {
					allRepos = append(allRepos, r)
					seen[r.FullName] = true
				}
			}
		}
	}

	var categories []Subcategory
	for _, sub := range rule.Subcategories {
		subTopicSet := make(map[string]bool, len(sub.Topics))
		for _, t := range sub.Topics {
			subTopicSet[t] = true
		}

		var matched []model.Repository
		for _, r := range allRepos {
			for _, t := range r.Topics {
				if subTopicSet[t] {
					matched = append(matched, r)
					break
				}
			}
		}

		// 按 stars DESC 排序
		sort.Slice(matched, func(i, j int) bool {
			return matched[i].Stars > matched[j].Stars
		})

		topRepos := toStackRepos(matched, 5, ctx, s, ecosystem)

		// Trending: trend=rising 的项目
		var risingRepos []model.Repository
		for _, r := range matched {
			ts := s.trendSvc.GetTopicTrendScore(ctx, firstTopic(r.Topics), trend.Window7d)
			normalized := (ts + 1) / 2
			if normalized > 0.6 {
				risingRepos = append(risingRepos, r)
			}
		}
		sort.Slice(risingRepos, func(i, j int) bool {
			return risingRepos[i].Stars > risingRepos[j].Stars
		})
		trending := toStackRepos(risingRepos, 3, ctx, s, ecosystem)

		categories = append(categories, Subcategory{
			Name:      sub.Name,
			RepoCount: len(matched),
			TopRepos:  topRepos,
			Trending:  trending,
		})
	}

	return &TechStackTree{
		Ecosystem:  ecosystem,
		Categories: categories,
	}, nil
}

// getTopReposInEcosystem 获取生态内 top repos
func (s *Service) getTopReposInEcosystem(ctx context.Context, ecosystem string, limit int) []StackRepo {
	ecoRepos, _ := s.graphStore.GetEcosystemRepos(ctx, ecosystem, 500)

	var repos []model.Repository
	for _, name := range ecoRepos {
		repo, err := s.repoStore.GetByFullName(ctx, name)
		if err != nil || repo == nil {
			continue
		}
		repos = append(repos, *repo)
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Stars > repos[j].Stars
	})

	if len(repos) > limit {
		repos = repos[:limit]
	}

	return toStackRepos(repos, limit, ctx, s, ecosystem)
}

// toStackRepos 将 model.Repository 转换为 StackRepo
func toStackRepos(repos []model.Repository, limit int, ctx context.Context, s *Service, ecosystem string) []StackRepo {
	if len(repos) > limit {
		repos = repos[:limit]
	}
	var result []StackRepo
	for _, r := range repos {
		ts := s.trendSvc.GetTopicTrendScore(ctx, firstTopic(r.Topics), trend.Window7d)
		normalized := (ts + 1) / 2
		result = append(result, StackRepo{
			FullName:    r.FullName,
			Description: r.Description,
			Stars:       r.Stars,
			Language:    r.Language,
			Trend:       classifyTrend(normalized),
		})
	}
	return result
}

// classifyTrend 将归一化 [0,1] 分数转为趋势状态
func classifyTrend(score float64) string {
	if score > 0.6 {
		return "rising"
	}
	if score < 0.4 {
		return "declining"
	}
	return "stable"
}

// firstTopic 返回第一个 topic 或空字符串
func firstTopic(topics []string) string {
	if len(topics) > 0 {
		return topics[0]
	}
	return ""
}
