-- 004_graph_tables.sql
-- GitSense Graph 系统：边表 + 生态映射

-- repo_edges: Repo ↔ Repo 相似关系边
CREATE TABLE repo_edges (
    id              BIGSERIAL PRIMARY KEY,
    src_repo        VARCHAR(511) NOT NULL,
    dst_repo        VARCHAR(511) NOT NULL,
    weight_embedding FLOAT DEFAULT 0,      -- 原始 embedding 信号
    weight_topic    FLOAT DEFAULT 0,       -- 原始 topic 信号
    score           FLOAT NOT NULL,        -- 最终融合评分
    created_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(src_repo, dst_repo)
);

CREATE INDEX idx_repo_edges_src ON repo_edges (src_repo);
CREATE INDEX idx_repo_edges_dst ON repo_edges (dst_repo);
CREATE INDEX idx_repo_edges_score ON repo_edges (score DESC);

-- topic_edges: Topic ↔ Topic 共现关系
CREATE TABLE topic_edges (
    id          BIGSERIAL PRIMARY KEY,
    topic_a     VARCHAR(255) NOT NULL,
    topic_b     VARCHAR(255) NOT NULL,
    weight      INT NOT NULL DEFAULT 1,    -- 共现次数
    strength    VARCHAR(10) DEFAULT 'candidate', -- candidate (>=2) / strong (>=3)
    created_at  TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(topic_a, topic_b)
);

CREATE INDEX idx_topic_edges_a ON topic_edges (topic_a);
CREATE INDEX idx_topic_edges_b ON topic_edges (topic_b);
CREATE INDEX idx_topic_edges_weight ON topic_edges (weight DESC);

-- ecosystem_map: Repo → Ecosystem 映射
CREATE TABLE ecosystem_map (
    id          BIGSERIAL PRIMARY KEY,
    repo_id     BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    ecosystem   VARCHAR(255) NOT NULL,
    confidence  FLOAT DEFAULT 1.0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(repo_id, ecosystem)
);

CREATE INDEX idx_ecosystem_map_ecosystem ON ecosystem_map (ecosystem);
CREATE INDEX idx_ecosystem_map_repo ON ecosystem_map (repo_id);
