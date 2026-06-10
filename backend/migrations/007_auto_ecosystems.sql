-- 007_auto_ecosystems.sql
-- Auto Ecosystem Discovery Tables (Phase 11)

CREATE TABLE auto_ecosystems (
    id SERIAL PRIMARY KEY,
    build_version INT NOT NULL,
    name TEXT NOT NULL,
    rule_mapping TEXT,
    is_emerging BOOLEAN NOT NULL DEFAULT FALSE,
    repo_count INT NOT NULL DEFAULT 0,
    topic_count INT NOT NULL DEFAULT 0,
    top_topics JSONB NOT NULL DEFAULT '[]',
    trend_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    modularity_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_auto_eco_version ON auto_ecosystems(build_version);

CREATE TABLE auto_ecosystem_topics (
    id SERIAL PRIMARY KEY,
    ecosystem_id INT NOT NULL REFERENCES auto_ecosystems(id) ON DELETE CASCADE,
    topic TEXT NOT NULL,
    normalized_topic TEXT NOT NULL,
    weight INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_auto_eco_topics_eco ON auto_ecosystem_topics(ecosystem_id);
CREATE INDEX idx_auto_eco_topics_norm ON auto_ecosystem_topics(normalized_topic);

CREATE TABLE auto_ecosystem_repos (
    id SERIAL PRIMARY KEY,
    ecosystem_id INT NOT NULL REFERENCES auto_ecosystems(id) ON DELETE CASCADE,
    repo_id INT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    build_version INT NOT NULL,
    assignment_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    assigned_topics JSONB NOT NULL DEFAULT '[]',
    assignment_method TEXT NOT NULL DEFAULT 'topic',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(repo_id, ecosystem_id, build_version)
);

CREATE INDEX idx_auto_eco_repos_eco ON auto_ecosystem_repos(ecosystem_id);
CREATE INDEX idx_auto_eco_repos_repo ON auto_ecosystem_repos(repo_id);
CREATE INDEX idx_auto_eco_repos_version ON auto_ecosystem_repos(build_version);

CREATE TABLE auto_ecosystem_builds (
    id SERIAL PRIMARY KEY,
    build_version INT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'building',
    clusters_found INT NOT NULL DEFAULT 0,
    repos_assigned INT NOT NULL DEFAULT 0,
    coverage_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    purity_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    emerging_count INT NOT NULL DEFAULT 0,
    duration_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);
