-- 001_init.sql
-- GitSense 初始化数据库迁移
--
-- 重要：embedding 维度必须与 EMBEDDING_PROVIDER 匹配
--   local/http (MiniLM) → 384
--   openai (text-embedding-3-small) → 1536
--   openai (text-embedding-3-large) → 3072
-- 切换 provider 时需要重建表或执行 003_resize_embedding.sql

-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 仓库表
CREATE TABLE repositories (
    id            BIGSERIAL PRIMARY KEY,
    full_name     VARCHAR(511) NOT NULL,       -- owner/name
    owner         VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    description   TEXT DEFAULT '',
    language      VARCHAR(100) DEFAULT '',
    stars         INTEGER DEFAULT 0,
    topics        JSONB DEFAULT '[]',
    readme        TEXT DEFAULT '',
    embedding     vector(384),                 -- 默认 MiniLM 维度，OpenAI 需改为 1536
    embedding_status VARCHAR(20) DEFAULT 'pending',  -- pending / generating / done / failed
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(full_name)
);

-- pgvector 近邻检索索引（IVFFlat，适合万级数据量）
-- 注意：IVFFlat 索引需要数据量 > lists 才能创建，初始数据少时跳过
-- 生产环境数据量达到 lists * 10 后执行：
-- CREATE INDEX idx_repositories_embedding ON repositories
--     USING ivfflat (embedding vector_cosine_ops)
--     WITH (lists = 100);

-- 语言过滤索引
CREATE INDEX idx_repositories_language ON repositories (language);

-- GIN 索引加速 topics JSONB 查询
CREATE INDEX idx_repositories_topics ON repositories USING GIN (topics);

-- embedding_status 索引，用于异步任务查询
CREATE INDEX idx_repositories_embedding_status ON repositories (embedding_status);

-- updated_at 索引
CREATE INDEX idx_repositories_updated_at ON repositories (updated_at);
