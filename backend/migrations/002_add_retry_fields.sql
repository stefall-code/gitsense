-- 002_add_retry_fields.sql
-- 添加 embedding 重试字段

ALTER TABLE repositories
    ADD COLUMN IF NOT EXISTS retry_count INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_attempt_at TIMESTAMPTZ;

-- 更新 embedding_status 索引，覆盖重试场景
DROP INDEX IF EXISTS idx_repositories_embedding_status;
CREATE INDEX idx_repositories_embedding_status ON repositories (embedding_status, retry_count, last_attempt_at);
