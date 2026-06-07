package bootstrap

import "time"

// JobStatus Bootstrap 任务状态
type JobStatus string

const (
	JobRunning   JobStatus = "running"
	JobPaused    JobStatus = "paused"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
)

// QueueStatus 队列项状态
type QueueStatus string

const (
	QueuePending   QueueStatus = "pending"
	QueueProcessing QueueStatus = "processing"
	QueueDone      QueueStatus = "done"
	QueueFailed    QueueStatus = "failed"
)

// SourceType 发现来源类型
type SourceType string

const (
	SourceAwesome SourceType = "awesome"
	SourceTopic   SourceType = "topic"
	SourceRelated SourceType = "related"
)

// BootstrapJob Bootstrap 任务
type BootstrapJob struct {
	ID             int        `json:"id"`
	Status         JobStatus  `json:"status"`
	ProcessedCount int        `json:"processed_count"`
	SuccessCount   int        `json:"success_count"`
	FailedCount    int        `json:"failed_count"`
	QueueSize      int        `json:"queue_size"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

// QueueItem 队列项
type QueueItem struct {
	ID             int        `json:"id"`
	JobID          int        `json:"job_id"`
	RepoFullName   string     `json:"repo_full_name"`
	SourceType     SourceType `json:"source_type"`
	DiscoveredFrom string     `json:"discovered_from,omitempty"`
	Depth          int        `json:"depth"`
	Status         QueueStatus `json:"status"`
	RetryCount     int        `json:"retry_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// BootstrapStatus Bootstrap 状态响应
type BootstrapStatus struct {
	Job    *BootstrapJob    `json:"job,omitempty"`
	GitHub GitHubRateLimit  `json:"github"`
	Dataset DatasetStats    `json:"dataset"`
}

// GitHubRateLimit GitHub API 额度信息
type GitHubRateLimit struct {
	CoreRemaining   int        `json:"core_remaining"`
	SearchRemaining int        `json:"search_remaining"`
	CoreResetAt     *time.Time `json:"core_reset_at,omitempty"`
	SearchResetAt   *time.Time `json:"search_reset_at,omitempty"`
}

// DatasetStats 数据集统计
type DatasetStats struct {
	Repos             int `json:"repos"`
	Topics            int `json:"topics"`
	Edges             int `json:"edges"`
	EmbeddingPending  int `json:"embedding_pending"`
	EmbeddingDone     int `json:"embedding_done"`
}

// SeedConfig 种子配置
type SeedConfig struct {
	AwesomeRepos []string
	TopicQueries []TopicQuery
}

// TopicQuery Topic 搜索查询
type TopicQuery struct {
	Topic    string
	MinStars int
}

// DefaultSeedConfig 默认种子配置
func DefaultSeedConfig() SeedConfig {
	return SeedConfig{
		AwesomeRepos: []string{
			"sindresorhus/awesome-ai",
			"vinta/awesome-python",
			"avelino/awesome-go",
			"josephmisiti/awesome-machine-learning",
			"awesome-selfhosted/awesome-selfhosted",
			"tiimgreen/github-awesome-devops",
			"sbilly/awesome-security",
			"rosarior/awesome-open-source",
			"Hannibal046/Awesome-LLM",
			"e2b-dev/awesome-ai-agents",
		},
		TopicQueries: []TopicQuery{
			{Topic: "agent", MinStars: 100},
			{Topic: "llm", MinStars: 100},
			{Topic: "security", MinStars: 100},
			{Topic: "rag", MinStars: 100},
			{Topic: "gateway", MinStars: 100},
			{Topic: "golang", MinStars: 100},
			{Topic: "python", MinStars: 100},
			{Topic: "devops", MinStars: 100},
			{Topic: "selfhosted", MinStars: 100},
			{Topic: "machine-learning", MinStars: 100},
		},
	}
}
