package service

import (
	"context"
	"fmt"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/repository"
)

// EcosystemService 处理技术生态发现
type EcosystemService struct {
	store      *repository.RepoStore
	cache      *repository.CacheStore
	classifier *EcosystemClassifier
}

// NewEcosystemService 创建新的 EcosystemService
func NewEcosystemService(store *repository.RepoStore, cache *repository.CacheStore, classifier *EcosystemClassifier) *EcosystemService {
	return &EcosystemService{store: store, cache: cache, classifier: classifier}
}

// GetEcosystem 发现仓库所属技术生态
func (s *EcosystemService) GetEcosystem(ctx context.Context, fullName string) (*model.Ecosystem, error) {
	// 查缓存
	cached, err := s.cache.GetEcosystem(ctx, fullName)
	if err == nil && cached != nil {
		return cached, nil
	}

	// 查目标仓库
	target, err := s.store.GetByFullName(ctx, fullName)
	if err != nil {
		return nil, fmt.Errorf("get target repository: %w", err)
	}
	if target == nil {
		return nil, fmt.Errorf("repository not found: %s", fullName)
	}

	if len(target.Topics) == 0 {
		return nil, nil
	}

	// 基于 topics 查询关联仓库
	related, err := s.store.GetByTopics(ctx, target.Topics, 50)
	if err != nil {
		return nil, fmt.Errorf("get by topics: %w", err)
	}

	// 过滤掉自身
	var ecosystemRepos []model.Repository
	for _, r := range related {
		if r.FullName == fullName {
			continue
		}
		ecosystemRepos = append(ecosystemRepos, r)
	}

	// 使用 EcosystemClassifier 分类
	ecoName := s.classifier.ClassifyWithFallback(target.Topics, target.Description)

	eco := &model.Ecosystem{
		Name:         ecoName,
		Repositories: ecosystemRepos,
	}

	// 写缓存
	_ = s.cache.SetEcosystem(ctx, fullName, eco)

	return eco, nil
}
