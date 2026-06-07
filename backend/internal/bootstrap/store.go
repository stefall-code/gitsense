package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store Bootstrap 数据访问层
type Store struct {
	pool *pgxpool.Pool
}

// NewStore 创建 Bootstrap Store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// --- Job 操作 ---

// CreateJob 创建新的 Bootstrap 任务
func (s *Store) CreateJob(ctx context.Context) (*BootstrapJob, error) {
	now := time.Now()
	job := &BootstrapJob{
		Status:    JobRunning,
		StartedAt: &now,
		UpdatedAt: &now,
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO bootstrap_jobs (status, started_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`, job.Status, job.StartedAt, job.UpdatedAt).Scan(&job.ID)

	return job, err
}

// GetActiveJob 获取当前活跃任务
func (s *Store) GetActiveJob(ctx context.Context) (*BootstrapJob, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, status, processed_count, success_count, failed_count, queue_size,
		       started_at, updated_at, finished_at
		FROM bootstrap_jobs
		WHERE status IN ('running', 'paused')
		ORDER BY id DESC LIMIT 1
	`)

	var job BootstrapJob
	err := row.Scan(
		&job.ID, &job.Status, &job.ProcessedCount, &job.SuccessCount,
		&job.FailedCount, &job.QueueSize,
		&job.StartedAt, &job.UpdatedAt, &job.FinishedAt,
	)
	if err != nil {
		return nil, nil // no active job
	}
	return &job, nil
}

// GetLatestJob 获取最新任务（含 completed/failed）
func (s *Store) GetLatestJob(ctx context.Context) (*BootstrapJob, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, status, processed_count, success_count, failed_count, queue_size,
		       started_at, updated_at, finished_at
		FROM bootstrap_jobs
		ORDER BY id DESC LIMIT 1
	`)

	var job BootstrapJob
	err := row.Scan(
		&job.ID, &job.Status, &job.ProcessedCount, &job.SuccessCount,
		&job.FailedCount, &job.QueueSize,
		&job.StartedAt, &job.UpdatedAt, &job.FinishedAt,
	)
	if err != nil {
		return nil, nil
	}
	return &job, nil
}

// UpdateJob 更新任务状态
func (s *Store) UpdateJob(ctx context.Context, job *BootstrapJob) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE bootstrap_jobs SET
			status = $2, processed_count = $3, success_count = $4, failed_count = $5,
			queue_size = $6, updated_at = NOW(), finished_at = $7
		WHERE id = $1
	`, job.ID, job.Status, job.ProcessedCount, job.SuccessCount,
		job.FailedCount, job.QueueSize, job.FinishedAt)
	return err
}

// --- Queue 操作 ---

// Enqueue 入队（去重：repo_full_name UNIQUE）
func (s *Store) Enqueue(ctx context.Context, jobID int, repoFullName string, sourceType SourceType, discoveredFrom string, depth int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO bootstrap_queue (job_id, repo_full_name, source_type, discovered_from, depth, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		ON CONFLICT (repo_full_name) DO NOTHING
	`, jobID, repoFullName, sourceType, discoveredFrom, depth)
	return err
}

// EnqueueBatch 批量入队
func (s *Store) EnqueueBatch(ctx context.Context, jobID int, items []QueueItem) (int, error) {
	inserted := 0
	for _, item := range items {
		err := s.Enqueue(ctx, jobID, item.RepoFullName, item.SourceType, item.DiscoveredFrom, item.Depth)
		if err == nil {
			inserted++
		}
	}
	return inserted, nil
}

// Dequeue 取出下一个待处理项
func (s *Store) Dequeue(ctx context.Context, jobID int) (*QueueItem, error) {
	// 原子操作：pending → processing
	row := s.pool.QueryRow(ctx, `
		UPDATE bootstrap_queue SET status = 'processing', updated_at = NOW()
		WHERE id = (
			SELECT id FROM bootstrap_queue
			WHERE job_id = $1 AND status = 'pending'
			ORDER BY id LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, job_id, repo_full_name, source_type, discovered_from, depth, status, retry_count, created_at, updated_at
	`, jobID)

	var item QueueItem
	err := row.Scan(
		&item.ID, &item.JobID, &item.RepoFullName, &item.SourceType,
		&item.DiscoveredFrom, &item.Depth, &item.Status, &item.RetryCount,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, nil // queue empty
	}
	return &item, nil
}

// MarkDone 标记完成
func (s *Store) MarkDone(ctx context.Context, itemID int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE bootstrap_queue SET status = 'done', updated_at = NOW() WHERE id = $1
	`, itemID)
	return err
}

// MarkFailed 标记失败
func (s *Store) MarkFailed(ctx context.Context, itemID int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE bootstrap_queue SET status = 'failed', retry_count = retry_count + 1, updated_at = NOW()
		WHERE id = $1
	`, itemID)
	return err
}

// ResetProcessing 重置 processing 状态为 pending（断点恢复用）
func (s *Store) ResetProcessing(ctx context.Context, jobID int) (int, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE bootstrap_queue SET status = 'pending', updated_at = NOW()
		WHERE job_id = $1 AND status = 'processing'
	`, jobID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// GetQueueSize 获取待处理队列长度
func (s *Store) GetQueueSize(ctx context.Context, jobID int) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bootstrap_queue WHERE job_id = $1 AND status = 'pending'
	`, jobID).Scan(&count)
	return count, err
}

// --- 统计 ---

// GetDatasetStats 获取数据集统计
func (s *Store) GetDatasetStats(ctx context.Context) (*DatasetStats, error) {
	stats := &DatasetStats{}

	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories`).Scan(&stats.Repos)
	s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT topic) FROM (
			SELECT jsonb_array_elements_text(topics) AS topic FROM repositories WHERE topics != '[]'::jsonb
		) sub
	`).Scan(&stats.Topics)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repo_edges`).Scan(&stats.Edges)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories WHERE embedding_status = 'done'`).Scan(&stats.EmbeddingDone)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories WHERE embedding_status = 'pending'`).Scan(&stats.EmbeddingPending)

	return stats, nil
}

// IsRepoExists 检查 repo 是否已存在
func (s *Store) IsRepoExists(ctx context.Context, fullName string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM repositories WHERE full_name = $1)
	`, fullName).Scan(&exists)
	return exists, err
}

// GetProcessedReposBySource 统计各来源的 repo 数量
func (s *Store) GetProcessedReposBySource(ctx context.Context, jobID int) (map[SourceType]int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT source_type, COUNT(*) FROM bootstrap_queue
		WHERE job_id = $1 AND status = 'done'
		GROUP BY source_type
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[SourceType]int)
	for rows.Next() {
		var st SourceType
		var cnt int
		if err := rows.Scan(&st, &cnt); err != nil {
			continue
		}
		result[st] = cnt
	}
	return result, nil
}

// IncrementJobCounts 原子递增 job 计数
func (s *Store) IncrementJobCounts(ctx context.Context, jobID int, success bool) error {
	if success {
		_, err := s.pool.Exec(ctx, `
			UPDATE bootstrap_jobs SET
				processed_count = processed_count + 1,
				success_count = success_count + 1,
				updated_at = NOW()
			WHERE id = $1
		`, jobID)
		return err
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE bootstrap_jobs SET
			processed_count = processed_count + 1,
			failed_count = failed_count + 1,
			updated_at = NOW()
		WHERE id = $1
	`, jobID)
	return err
}

// UpdateQueueSize 更新 job 的 queue_size 字段
func (s *Store) UpdateQueueSize(ctx context.Context, jobID int) error {
	var count int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bootstrap_queue WHERE job_id = $1 AND status = 'pending'
	`, jobID).Scan(&count)

	_, err := s.pool.Exec(ctx, `
		UPDATE bootstrap_jobs SET queue_size = $2, updated_at = NOW() WHERE id = $1
	`, jobID, count)
	return err
}

// GetDiscoveredFrom 获取某个 repo 的 discovered_from
func (s *Store) GetDiscoveredFrom(ctx context.Context, repoFullName string) (string, error) {
	var discoveredFrom string
	err := s.pool.QueryRow(ctx, `
		SELECT discovered_from FROM bootstrap_queue WHERE repo_full_name = $1 LIMIT 1
	`, repoFullName).Scan(&discoveredFrom)
	if err != nil {
		return "", fmt.Errorf("discovered_from not found for %s: %w", repoFullName, err)
	}
	return discoveredFrom, nil
}
