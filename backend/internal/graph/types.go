package graph

import "time"

// --- Node Types ---

type NodeType string

const (
	NodeRepo      NodeType = "repo"
	NodeTopic     NodeType = "topic"
	NodeLanguage  NodeType = "language"
	NodeEcosystem NodeType = "ecosystem"
)

// GraphNode 图节点
type GraphNode struct {
	ID    string    `json:"id"`
	Type  NodeType  `json:"type"`
	Label string    `json:"label"`
}

// --- Edge Types ---

type EdgeType string

const (
	EdgeSimilarTo   EdgeType = "SIMILAR_TO"
	EdgeHasTopic    EdgeType = "HAS_TOPIC"
	EdgeUsesLang    EdgeType = "USES_LANGUAGE"
	EdgeBelongsTo   EdgeType = "BELONGS_TO"
	EdgeCoOccur     EdgeType = "CO_OCCUR"
)

// RepoEdge Repo ↔ Repo 相似关系边
type RepoEdge struct {
	ID             int64     `json:"id"`
	SrcRepo        string    `json:"src_repo"`
	DstRepo        string    `json:"dst_repo"`
	WeightEmbedding float64  `json:"weight_embedding"`
	WeightTopic    float64   `json:"weight_topic"`
	Score          float64   `json:"score"`
	CreatedAt      time.Time `json:"created_at"`
}

// TopicEdge Topic ↔ Topic 共现关系边
type TopicEdge struct {
	ID        int64     `json:"id"`
	TopicA    string    `json:"topic_a"`
	TopicB    string    `json:"topic_b"`
	Weight    int       `json:"weight"`
	Strength  string    `json:"strength"` // candidate / strong
	CreatedAt time.Time `json:"created_at"`
}

// EcosystemMapEntry Repo → Ecosystem 映射
type EcosystemMapEntry struct {
	ID         int64   `json:"id"`
	RepoID     int64   `json:"repo_id"`
	Ecosystem  string  `json:"ecosystem"`
	Confidence float64 `json:"confidence"`
}

// --- API Response Types ---

// RepoGraphResponse GET /graph/repo/:owner/:repo 响应
type RepoGraphResponse struct {
	Repo        string      `json:"repo"`
	SimilarTo   []RepoEdge  `json:"similar_to"`
	Topics      []string    `json:"topics"`
	Language    string      `json:"language"`
	Ecosystem   string      `json:"ecosystem,omitempty"`
	Nodes       []GraphNode `json:"nodes"`
}

// EcosystemGraphResponse GET /graph/ecosystem/:name 响应
type EcosystemGraphResponse struct {
	Ecosystem   string      `json:"ecosystem"`
	Repos       []string    `json:"repos"`
	TopicCluster []TopicEdge `json:"topic_cluster"`
	Nodes       []GraphNode `json:"nodes"`
}

// PathResponse GET /graph/path 响应
type PathResponse struct {
	From  string       `json:"from"`
	To    string       `json:"to"`
	Paths [][]RepoEdge `json:"paths"`
	Hops  int          `json:"max_hops"`
}

// GraphExplanation 推荐解释中的图路径信息
type GraphExplanation struct {
	SimilarityPath  []string `json:"similarity_path,omitempty"`
	SharedEcosystem string   `json:"shared_ecosystem,omitempty"`
	TopicBridge     []string `json:"topic_bridge,omitempty"`
	GraphScore      float64  `json:"graph_score,omitempty"`
}

// BuildGraphRequest POST /admin/build-graph 请求
type BuildGraphRequest struct {
	FullRebuild bool `json:"full_rebuild"`
}

// BuildGraphResponse POST /admin/build-graph 响应
type BuildGraphResponse struct {
	RepoEdges    int `json:"repo_edges"`
	TopicEdges   int `json:"topic_edges"`
	EcoMappings  int `json:"ecosystem_mappings"`
}
