package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gitsense/gitsense/backend/internal/config"
	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// RepoStore 处理 PostgreSQL 数据访问
type RepoStore struct {
	pool *pgxpool.Pool
}

// NewRepoStore 创建新的 RepoStore
func NewRepoStore(ctx context.Context, cfg config.DBConfig) (*RepoStore, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return &RepoStore{pool: pool}, nil
}

// Close 关闭数据库连接池
func (s *RepoStore) Close() {
	s.pool.Close()
}

// Pool 返回底层连接池（供其他模块共享使用）
func (s *RepoStore) Pool() *pgxpool.Pool {
	return s.pool
}

// Ping 检查数据库连接
func (s *RepoStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// GetByFullName 根据 full_name 查询仓库
func (s *RepoStore) GetByFullName(ctx context.Context, fullName string) (*model.Repository, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, full_name, owner, name, description, language, stars, topics, readme,
		       embedding, embedding_status, retry_count, last_attempt_at, last_activity_at, pushed_at, created_at, updated_at
		FROM repositories WHERE full_name = $1
	`, fullName)
	return scanRepository(row)
}

// GetByID 根据 ID 查询仓库
func (s *RepoStore) GetByID(ctx context.Context, id int64) (*model.Repository, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, full_name, owner, name, description, language, stars, topics, readme,
		       embedding, embedding_status, retry_count, last_attempt_at, last_activity_at, pushed_at, created_at, updated_at
		FROM repositories WHERE id = $1
	`, id)
	return scanRepository(row)
}

// Upsert 插入或更新仓库（不含 embedding）
func (s *RepoStore) Upsert(ctx context.Context, repo *model.Repository) error {
	topicsJSON, err := json.Marshal(repo.Topics)
	if err != nil {
		return fmt.Errorf("marshal topics: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO repositories (full_name, owner, name, description, language, stars, topics, readme, embedding_status, last_activity_at, pushed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (full_name) DO UPDATE SET
			description = EXCLUDED.description,
			language = EXCLUDED.language,
			stars = EXCLUDED.stars,
			topics = EXCLUDED.topics,
			readme = EXCLUDED.readme,
			embedding_status = EXCLUDED.embedding_status,
			last_activity_at = EXCLUDED.last_activity_at,
			pushed_at = EXCLUDED.pushed_at,
			updated_at = NOW()
	`, repo.FullName, repo.Owner, repo.Name, repo.Description, repo.Language, repo.Stars, topicsJSON, repo.Readme, repo.EmbeddingStatus, repo.LastActivityAt, repo.PushedAt)

	return err
}

// UpdateEmbedding 更新仓库的 embedding 向量和状态（成功时重置 retry_count）
func (s *RepoStore) UpdateEmbedding(ctx context.Context, fullName string, embedding *pgvector.Vector, status model.EmbeddingStatus) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE repositories SET embedding = $1, embedding_status = $2, retry_count = 0, last_attempt_at = NOW(), updated_at = NOW()
		WHERE full_name = $3
	`, embedding, status, fullName)
	return err
}

// UpdateEmbeddingStatus 仅更新 embedding 状态
func (s *RepoStore) UpdateEmbeddingStatus(ctx context.Context, fullName string, status model.EmbeddingStatus) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE repositories SET embedding_status = $1, updated_at = NOW()
		WHERE full_name = $2
	`, status, fullName)
	return err
}

// MarkEmbeddingFailed 标记 embedding 失败并增加 retry_count
func (s *RepoStore) MarkEmbeddingFailed(ctx context.Context, fullName string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE repositories
		SET embedding_status = 'failed',
		    retry_count = retry_count + 1,
		    last_attempt_at = NOW(),
		    updated_at = NOW()
		WHERE full_name = $1
	`, fullName)
	return err
}

// SearchSimilar 使用 pgvector 检索相似仓库
func (s *RepoStore) SearchSimilar(ctx context.Context, fullName string, embedding pgvector.Vector, limit int) ([]model.Repository, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, full_name, owner, name, description, language, stars, topics, readme,
		       embedding, embedding_status, retry_count, last_attempt_at, last_activity_at, pushed_at, created_at, updated_at
		FROM repositories
		WHERE full_name != $1 AND embedding_status = 'done'
		ORDER BY embedding <=> $2
		LIMIT $3
	`, fullName, embedding, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRepositories(rows)
}

// GetPendingEmbeddings 获取待生成 embedding 的仓库列表
// 包含 pending + 可重试的 failed（retry_count < 3 且冷却时间已过）
func (s *RepoStore) GetPendingEmbeddings(ctx context.Context, limit int) ([]model.Repository, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, full_name, owner, name, description, language, stars, topics, readme,
		       embedding, embedding_status, retry_count, last_attempt_at, last_activity_at, pushed_at, created_at, updated_at
		FROM repositories
		WHERE embedding_status = 'pending'
		   OR (embedding_status = 'failed' AND retry_count < 3 AND (last_attempt_at IS NULL OR last_attempt_at < NOW() - INTERVAL '1 hour'))
		ORDER BY updated_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRepositories(rows)
}

// GetByTopics 根据 topics 查询仓库（用于生态发现）
func (s *RepoStore) GetByTopics(ctx context.Context, topics []string, limit int) ([]model.Repository, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, full_name, owner, name, description, language, stars, topics, readme,
		       embedding, embedding_status, retry_count, last_attempt_at, last_activity_at, pushed_at, created_at, updated_at
		FROM repositories
		WHERE topics ?| $1
		ORDER BY stars DESC
		LIMIT $2
	`, topics, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRepositories(rows)
}

// --- scan helpers ---

const repoColumns = `id, full_name, owner, name, description, language, stars, topics, readme, embedding, embedding_status, retry_count, last_attempt_at, last_activity_at, pushed_at, created_at, updated_at`

func scanRepository(row pgx.Row) (*model.Repository, error) {
	var r model.Repository
	var topicsJSON []byte

	err := row.Scan(
		&r.ID, &r.FullName, &r.Owner, &r.Name, &r.Description,
		&r.Language, &r.Stars, &topicsJSON, &r.Readme,
		&r.Embedding, &r.EmbeddingStatus, &r.RetryCount, &r.LastAttemptAt,
		&r.LastActivityAt, &r.PushedAt,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(topicsJSON, &r.Topics); err != nil {
		return nil, fmt.Errorf("unmarshal topics: %w", err)
	}

	return &r, nil
}

func scanRepositories(rows pgx.Rows) ([]model.Repository, error) {
	var repos []model.Repository
	for rows.Next() {
		var r model.Repository
		var topicsJSON []byte

		err := rows.Scan(
			&r.ID, &r.FullName, &r.Owner, &r.Name, &r.Description,
			&r.Language, &r.Stars, &topicsJSON, &r.Readme,
			&r.Embedding, &r.EmbeddingStatus, &r.RetryCount, &r.LastAttemptAt,
			&r.LastActivityAt, &r.PushedAt,
			&r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(topicsJSON, &r.Topics); err != nil {
			return nil, fmt.Errorf("unmarshal topics: %w", err)
		}

		repos = append(repos, r)
	}
	return repos, rows.Err()
}

// --- CacheStore ---

// CacheStore 处理 Redis 缓存
type CacheStore struct {
	client redisClient
	cfg    config.CacheConfig
}

type redisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

func NewCacheStore(ctx context.Context, cfg config.RedisConfig, cacheCfg config.CacheConfig) (*CacheStore, error) {
	_ = ctx
	return &CacheStore{cfg: cacheCfg}, nil
}

func (c *CacheStore) GetRepo(ctx context.Context, fullName string) (*model.Repository, error) {
	_ = ctx; _ = fullName
	return nil, nil
}

func (c *CacheStore) SetRepo(ctx context.Context, repo *model.Repository) error {
	_ = ctx; _ = repo
	return nil
}

func (c *CacheStore) GetRecommendation(ctx context.Context, fullName string, limit int) ([]model.SimilarRepository, error) {
	_ = ctx; _ = fullName; _ = limit
	return nil, nil
}

func (c *CacheStore) SetRecommendation(ctx context.Context, fullName string, limit int, recs []model.SimilarRepository) error {
	_ = ctx; _ = fullName; _ = limit; _ = recs
	return nil
}

func (c *CacheStore) GetEcosystem(ctx context.Context, fullName string) (*model.Ecosystem, error) {
	_ = ctx; _ = fullName
	return nil, nil
}

func (c *CacheStore) SetEcosystem(ctx context.Context, fullName string, eco *model.Ecosystem) error {
	_ = ctx; _ = fullName; _ = eco
	return nil
}
