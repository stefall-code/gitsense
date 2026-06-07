-- 006_bootstrap_tables.sql
-- Dataset Bootstrap System: 任务管理 + 持久化队列

-- Bootstrap 任务表
CREATE TABLE IF NOT EXISTS bootstrap_jobs (
    id SERIAL PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'pending',  -- running/paused/completed/failed
    processed_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    queue_size INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

-- Bootstrap 持久化队列
CREATE TABLE IF NOT EXISTS bootstrap_queue (
    id SERIAL PRIMARY KEY,
    job_id INT NOT NULL REFERENCES bootstrap_jobs(id),
    repo_full_name TEXT NOT NULL,
    source_type TEXT NOT NULL DEFAULT 'awesome',  -- awesome/topic/related
    discovered_from TEXT,                          -- 发现来源 repo
    depth INT NOT NULL DEFAULT 0,                  -- BFS 深度 (0=种子, 1=README发现, 2=Topic扩散)
    status TEXT NOT NULL DEFAULT 'pending',        -- pending/processing/done/failed
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 唯一索引：同一 repo 不重复入队
    UNIQUE(repo_full_name)
);

-- 索引：按状态查询待处理项
CREATE INDEX IF NOT EXISTS idx_bootstrap_queue_status ON bootstrap_queue(status);
CREATE INDEX IF NOT EXISTS idx_bootstrap_queue_job ON bootstrap_queue(job_id);
