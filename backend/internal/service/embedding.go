package service

import (
	"context"
	"fmt"
	"log"

	"github.com/gitsense/gitsense/backend/internal/github"
	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/repository"
	"github.com/pgvector/pgvector-go"
)

// EmbeddingService 处理 Embedding 异步生成
type EmbeddingService struct {
	store    *repository.RepoStore
	provider EmbeddingProvider
}

// NewEmbeddingService 创建新的 EmbeddingService
func NewEmbeddingService(store *repository.RepoStore, provider EmbeddingProvider) *EmbeddingService {
	return &EmbeddingService{store: store, provider: provider}
}

// SetProvider 运行时替换 provider（用于启动时延迟注入）
func (s *EmbeddingService) SetProvider(provider EmbeddingProvider) {
	s.provider = provider
}

// GenerateForRepo 为指定仓库生成 embedding（单条，用于 channel 触发）
func (s *EmbeddingService) GenerateForRepo(ctx context.Context, fullName string) error {
	if s.provider == nil {
		return fmt.Errorf("embedding provider not configured")
	}

	// 更新状态为 generating
	if err := s.store.UpdateEmbeddingStatus(ctx, fullName, model.EmbeddingGenerating); err != nil {
		return fmt.Errorf("update embedding status to generating: %w", err)
	}

	repo, err := s.store.GetByFullName(ctx, fullName)
	if err != nil {
		return fmt.Errorf("get repository: %w", err)
	}
	if repo == nil {
		return fmt.Errorf("repository not found: %s", fullName)
	}

	text := github.BuildEmbeddingText(repo.Readme, repo.Description, repo.Topics)
	if text == "" {
		_ = s.store.MarkEmbeddingFailed(ctx, fullName)
		return fmt.Errorf("no text content for embedding: %s", fullName)
	}

	vec, err := s.provider.Generate(ctx, text)
	if err != nil {
		_ = s.store.MarkEmbeddingFailed(ctx, fullName)
		return fmt.Errorf("generate embedding: %w", err)
	}

	embedding := pgvector.NewVector(vec)
	if err := s.store.UpdateEmbedding(ctx, fullName, &embedding, model.EmbeddingDone); err != nil {
		return fmt.Errorf("update embedding: %w", err)
	}

	log.Printf("[embedding] generated for %s (dims=%d)", fullName, len(vec))
	return nil
}

// ProcessPending 批量处理待生成 embedding 的仓库（使用 batch API）
func (s *EmbeddingService) ProcessPending(ctx context.Context, batchSize int) error {
	if s.provider == nil {
		return fmt.Errorf("embedding provider not configured")
	}

	repos, err := s.store.GetPendingEmbeddings(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("get pending embeddings: %w", err)
	}

	if len(repos) == 0 {
		return nil
	}

	// 构建 batch 输入
	texts := make([]string, len(repos))
	for i, repo := range repos {
		texts[i] = github.BuildEmbeddingText(repo.Readme, repo.Description, repo.Topics)
	}

	// 标记所有为 generating
	for _, repo := range repos {
		_ = s.store.UpdateEmbeddingStatus(ctx, repo.FullName, model.EmbeddingGenerating)
	}

	// 批量生成 embedding
	embeddings, err := s.provider.GenerateBatch(ctx, texts)
	if err != nil {
		// batch 失败，全部标记 failed
		for _, repo := range repos {
			_ = s.store.MarkEmbeddingFailed(ctx, repo.FullName)
		}
		return fmt.Errorf("batch generate embedding: %w", err)
	}

	// 逐个写入
	for i, repo := range repos {
		if i >= len(embeddings) {
			_ = s.store.MarkEmbeddingFailed(ctx, repo.FullName)
			continue
		}

		if texts[i] == "" {
			_ = s.store.MarkEmbeddingFailed(ctx, repo.FullName)
			continue
		}

		embedding := pgvector.NewVector(embeddings[i])
		if err := s.store.UpdateEmbedding(ctx, repo.FullName, &embedding, model.EmbeddingDone); err != nil {
			log.Printf("[embedding] failed to store for %s: %v", repo.FullName, err)
			_ = s.store.MarkEmbeddingFailed(ctx, repo.FullName)
			continue
		}

		log.Printf("[embedding] generated for %s (dims=%d)", repo.FullName, len(embeddings[i]))
	}

	return nil
}

// GenerateEmbedding 为任意文本生成 embedding（用于 /search/similar API）
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("embedding provider not configured")
	}
	return s.provider.Generate(ctx, text)
}
