-- 005_add_activity_fields.sql
-- 为 Trend Detection 增加时间维度字段
-- 仅扩展，不修改现有列逻辑

-- GitHub updated_at：综合活跃度信号（README 更新、issue/PR 活动、release 更新）
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS last_activity_at TIMESTAMPTZ;

-- GitHub pushed_at：真实开发活跃度（可选辅助信号）
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS pushed_at TIMESTAMPTZ;

-- last_activity_at 索引（Trend 计算核心查询字段）
CREATE INDEX IF NOT EXISTS idx_repositories_last_activity_at ON repositories (last_activity_at);

-- 初始化：用 updated_at 回填 last_activity_at（存量数据兼容）
UPDATE repositories SET last_activity_at = updated_at WHERE last_activity_at IS NULL;
