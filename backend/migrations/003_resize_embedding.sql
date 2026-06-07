-- 003_resize_embedding.sql
-- 切换 Embedding Provider 维度时执行此迁移
-- 警告：此操作会清除所有 embedding 数据（需要重新生成）

-- 使用方法：
--   切换到 OpenAI (1536): 修改下方维度为 1536 后执行
--   切换到 MiniLM (384):  修改下方维度为 384 后执行

-- 步骤 1: 清除 embedding 数据
UPDATE repositories SET embedding = NULL, embedding_status = 'pending' WHERE embedding IS NOT NULL;

-- 步骤 2: 删除旧索引（如果存在）
DROP INDEX IF EXISTS idx_repositories_embedding;

-- 步骤 3: 修改列维度
-- ⚠️ 修改下方 384 为目标维度
ALTER TABLE repositories ALTER COLUMN embedding TYPE vector(384);

-- 步骤 4: 重建索引（数据量足够时）
-- CREATE INDEX idx_repositories_embedding ON repositories
--     USING ivfflat (embedding vector_cosine_ops)
--     WITH (lists = 100);
