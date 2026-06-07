package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gitsense/gitsense/backend/internal/github"
	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/repository"
)

const (
	maxReadmeBytes = 8000 // README 清洗后最大字节数（按 bytes 截断，防 emoji/中文爆炸）
)

// CollectorService 处理 GitHub 数据采集
type CollectorService struct {
	store  *repository.RepoStore
	client *github.Client
}

// NewCollectorService 创建新的 CollectorService
func NewCollectorService(store *repository.RepoStore, client *github.Client) *CollectorService {
	return &CollectorService{store: store, client: client}
}

// FetchAndStore 采集 GitHub 仓库信息并入库（不含 embedding）
// embedding 由 EmbeddingWorker 异步生成
func (s *CollectorService) FetchAndStore(ctx context.Context, owner, name string) (*model.Repository, error) {
	fullName := fmt.Sprintf("%s/%s", owner, name)

	// 先查 DB
	existing, err := s.store.GetByFullName(ctx, fullName)
	if err != nil {
		return nil, fmt.Errorf("query repository: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// 采集
	repo, err := s.fetchFromGitHub(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("fetch from github: %w", err)
	}

	// 数据完整性校验：必须有 full_name, description/readme, topics, language
	if !s.validateRepo(repo) {
		return nil, fmt.Errorf("incomplete repository data for %s: missing required fields", fullName)
	}

	// 入库，embedding_status = pending
	repo.EmbeddingStatus = model.EmbeddingPending
	if err := s.store.Upsert(ctx, repo); err != nil {
		return nil, fmt.Errorf("upsert repository: %w", err)
	}

	log.Printf("[collector] fetched and stored %s (stars=%d, lang=%s, topics=%v)",
		fullName, repo.Stars, repo.Language, repo.Topics)

	return repo, nil
}

// fetchFromGitHub 完整采集：元信息 + README + Languages
func (s *CollectorService) fetchFromGitHub(ctx context.Context, owner, name string) (*model.Repository, error) {
	// 1. 获取仓库元信息
	repoInfo, err := s.client.FetchRepository(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("fetch repository info: %w", err)
	}

	// 2. 获取 README
	readme, err := s.client.FetchReadme(ctx, owner, name)
	if err != nil {
		log.Printf("[collector] failed to fetch readme for %s/%s: %v", owner, name, err)
		readme = ""
	}

	// 3. 清洗 README
	readmeClean := github.CleanReadme(readme, maxReadmeBytes)

	// 4. 获取语言占比，确定主要语言
	langMap, err := s.client.FetchLanguages(ctx, owner, name)
	if err != nil {
		log.Printf("[collector] failed to fetch languages for %s/%s: %v", owner, name, err)
	}
	primaryLang := repoInfo.Language
	if primaryLang == "" && len(langMap) > 0 {
		primaryLang = github.GetPrimaryLanguage(langMap)
	}

	// 5. 组装 Repository
	topics := repoInfo.Topics
	if topics == nil {
		topics = []string{}
	}

	// 解析 GitHub 时间字段
	var lastActivityAt, pushedAt *time.Time
	if t, err := time.Parse(time.RFC3339, repoInfo.UpdatedAt); err == nil {
		lastActivityAt = &t
	}
	if t, err := time.Parse(time.RFC3339, repoInfo.PushedAt); err == nil {
		pushedAt = &t
	}

	return &model.Repository{
		FullName:       repoInfo.FullName,
		Owner:          repoInfo.Owner,
		Name:           repoInfo.Name,
		Description:    repoInfo.Description,
		Language:       primaryLang,
		Stars:          repoInfo.Stars,
		Topics:         topics,
		Readme:         readmeClean,
		LastActivityAt: lastActivityAt,
		PushedAt:       pushedAt,
	}, nil
}

// validateRepo 校验仓库数据完整性
// 必须有 full_name + (description 或 readme) + topics + language
func (s *CollectorService) validateRepo(repo *model.Repository) bool {
	if repo.FullName == "" {
		return false
	}
	// 至少要有 description 或 readme
	if repo.Description == "" && repo.Readme == "" {
		log.Printf("[collector] validation failed: %s has no description and no readme", repo.FullName)
		return false
	}
	// language 可以为空（某些仓库确实没有），但记录警告
	if repo.Language == "" {
		log.Printf("[collector] warning: %s has no language", repo.FullName)
	}
	return true
}

// BatchFetchAndStore 批量采集并入库
func (s *CollectorService) BatchFetchAndStore(ctx context.Context, repoNames []string) ([]*model.Repository, []string) {
	var succeeded []*model.Repository
	var failed []string

	for _, input := range repoNames {
		owner, name, err := ParseRepoInput(strings.TrimSpace(input))
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", input, err))
			continue
		}

		repo, err := s.FetchAndStore(ctx, owner, name)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", input, err))
			continue
		}

		succeeded = append(succeeded, repo)
	}

	return succeeded, failed
}

// ParseRepoInput 解析用户输入，支持两种格式：
// 1. URL: https://github.com/owner/name
// 2. owner/name
func ParseRepoInput(input string) (owner, name string, err error) {
	input = strings.TrimSpace(input)

	// 尝试解析为 URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		path := input
		for _, prefix := range []string{"https://", "http://"} {
			path = strings.TrimPrefix(path, prefix)
		}
		path = strings.TrimPrefix(path, "github.com/")
		path = strings.TrimPrefix(path, "www.github.com/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
		return "", "", fmt.Errorf("invalid github url: %s", input)
	}

	// 尝试解析为 owner/name
	parts := strings.Split(input, "/")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("invalid repository input: %s (expected 'owner/name' or github url)", input)
}
