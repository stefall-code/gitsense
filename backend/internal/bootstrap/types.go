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
			// AI / ML / LLM
			"sindresorhus/awesome-ai",
			"vinta/awesome-python",
			"josephmisiti/awesome-machine-learning",
			"Hannibal046/Awesome-LLM",
			"e2b-dev/awesome-ai-agents",
			"kyrolabs/awesome-agents",
			"deepseek-ai/awesome-deepseek",
			"hijkzzz/Awesome-LLM-Strawberry",
			"AGI-Edgerunners/LLM-Agents-Papers",
			"kaushikb11/awesome-llm-apps",
			"microsoft/ML-For-Beginners",
			"microsoft/generative-ai-for-beginners",
			"langchain-ai/langchain",
			"openai/openai-cookbook",
			"brexhq/prompt-engineering",
			"dair-ai/Prompt-Engineering-Guide",
			"ollama/ollama",
			"lm-sys/FastChat",
			"oobabooga/text-generation-webui",
			"AUTOMATIC1111/stable-diffusion-webui",
			"comfyanonymous/ComfyUI",
			// Go
			"avelino/awesome-go",
			// DevOps / Infra
			"awesome-selfhosted/awesome-selfhosted",
			"tiimgreen/github-awesome-devops",
			"veggiemonk/awesome-docker",
			"alexkorr/awesome-kubernetes",
			"ramitsurana/awesome-kubernetes",
			"jenkinsci/awesome-jenkins",
			"cicdops/awesome-ciandcd",
			"operate-first/SRE",
			"upgundecha/howtheysre",
			// Security
			"sbilly/awesome-security",
			"paralax/awesome-cybersecurity",
			"onlurking/awesome-infosec",
			"geraldb/talks",
			"joe-shenouda/awesome-cyber-skills",
			// Data / Visualization
			"fabiorosado/awesome-data-science",
			"josephmisiti/awesome-data-engineering",
			"christophergandrud/awesome-R",
			"toplap/awesome-livecoding",
			"markusschanta/awesome-jupyter",
			"krzjoa/awesome-python-data-science",
			// Frontend / Web
			"dypsilon/frontend-dev-bookmarks",
			"nicolesaidy/awesome-web-design",
			"brillout/awesome-react-components",
			"vuejs/awesome-vue",
			"sindresorhus/awesome-nodejs",
			"TheComputerM/awesome-svelte",
			"angularbrazil/awesome-angular",
			"ember-community-russia/awesome-ember",
			"solidjs/solid",
			"qwikdev/qwik",
			// Mobile
			"vsouza/awesome-ios",
			"JStumpp/awesome-android",
			"vuejs/awesome-flutter",
			"reactnativecommunity/awesome-react-native",
			// Systems / Low-level
			"aleksandar-todorovic/awesome-c",
			"rust-unofficial/awesome-rust",
			"jondot/awesome-devenv",
			"ipfs/awesome-ipfs",
			// Open Source
			"rosarior/awesome-open-source",
			// Database
			"numetriclabz/awesome-db",
			"dhamaniasad/awesome-postgres",
			"erictleung/awesome-nosql",
			"mehdihadeli/awesome-software-engineering",
			// Blockchain / Web3
			"0xpranay/awesome-web3",
			"chaozh/awesome-blockchain",
			// Gaming
			"ellisonleao/magictools",
			"leereilly/games",
			// Education
			"josephmcasey/awesome-education",
			"prakhar1989/awesome-courses",
			// Design
			"brabadu/awesome-fonts",
			// CLI / Terminal
			"alebcay/awesome-elixir",
			"alebcay/awesome-shell",
			"k4m4/terminals-are-sexy",
			// Science
			"nschloe/awesome-scientific-computing",
			"rossant/awesome-math",
			// Java / JVM
			"akullpp/awesome-java",
			"eleventy/awesome-eleventy",
			"json4s/awesome-scala",
			// C# / .NET
			"quozd/awesome-dotnet",
			// PHP
			"ziadoz/awesome-php",
			// Ruby
			"markets/awesome-ruby",
			// Swift
			"matteocrippa/awesome-swift",
			// Kotlin
			"KotlinBy/awesome-kotlin",
			// TypeScript
			"dzharii/awesome-typescript",
			// C++
			"fffaraz/awesome-cpp",
			// Haskell
			"krispo/awesome-haskell",
			// Lua
			"LewisJEllis/awesome-lua",
			// Zig
			"catdevnull/awesome-zig",
			// Embedded / IoT
			"nholut/awesome-embedded",
			"hq450/fancyss",
			// Audio / Music
			"notthetup/awesome-music",
			"worldmaking/awesome-gamedev",
			// Testing
			" TheJambo/awesome-testing",
			// API
			"Kikobeats/awesome-api",
			// Microservices
			"mfornos/awesome-microservices",
			// Sysadmin
			"nahamsec/Resources-for-Beginner-Bug-Bounty-Hunters",
			"k4m4/movies-for-hackers",
			// === 新增领域 ===
			// DevTools / IDE
			"viatsko/awesome-vscode",
			"rothgar/awesome-tmux",
			"alebcay/awesome-shell",
			"iperezdw/awesome-ides",
			// Math / Science
			"rossant/awesome-math",
			"nschloe/awesome-scientific-computing",
			"openjournals/awesome-journal",
			// Biology / Bioinformatics
			"danieldooley/awesome-bioinformatics",
			"seandavi/awesome-bioinformatics",
			// Chemistry
			"hsiaoyi0504/awesome-chemistry",
			// Physics
			"wbnns/awesome-quantum-computing",
			// Geography / GIS
			"sacridini/Awesome-Geospatial",
			// Finance / Trading
			"shlomikushchi/awesome-finance",
			"wilfreddesert/awesome-quant",
			"edwardleardi/awesome-quant",
			// Design / UI/UX
			"alexpate/awesome-design-systems",
			"diessica/awesome-sketch",
			"ktzn/awesome-figma",
			// Productivity
			"jyguyomarch/awesome-productivity",
			// Writing / Docs
			"matteocrippa/awesome-swift",
			"BubuAnabelas/awesome-markdown",
			"ekstay/awesome-technical-writing",
			// Privacy
			"pluja/awesome-privacy",
			"Lissy93/awesome-privacy",
			// Self-hosted (more)
			"awesome-foss/awesome-selfhosted",
			// Networking
			"clowwindy/awesome-networking",
			// Distributed Systems
			"theanalyst/awesome-distributed-systems",
			// Algorithms
			"tayllan/awesome-algorithms",
			"enjalot/awesome-algorithms",
			// Data Structures
			"ramitsurana/awesome-kubernetes",
			// Concurrency
			"lubobill1990/awesome-concurrency",
			// Functional Programming
			"lucasviola/awesome-functional-programming",
			// Compilers
			"aaronhertzmann/awesome-compilers",
			// Formal Methods
			"johnyf/awesome-formal-methods",
			// Robotics
			"kiloreux/awesome-robotics",
			// Computer Graphics
			"ericjang/awesome-graphics",
			// Computer Vision (more)
			"jamiebuilds/awesome-computer-vision",
			// NLP (more)
			"keon/awesome-nlp",
			// Big Data
			"onurakpolat/awesome-bigdata",
			// Streaming
			"anthonynsimon/awesome-streaming",
			// Datasets
			"awesomedata/awesome-public-datasets",
			"academic/awesome-datascience",
			// Open Source (more)
			"rosarior/awesome-open-source",
			// Linux
			"inputsh/awesome-linux",
			"LewisJEllis/awesome-linux-software",
			// macOS
			"iCHAIT/awesome-macOS",
			"serhii-londar/open-source-mac-os-apps",
			// Windows
			"0pandadev/awesome-windows",
			// Homelab
			"awesome-selfhosted/awesome-selfhosted",
			// SRE
			"dastergon/awesome-sre",
			// Chaos Engineering
			"dastergon/awesome-chaos-engineering",
			// Site Reliability
			"upgundecha/howtheysre",
			// Serverless
			"pmuens/awesome-serverless",
			// Terraform
			"shuaibiyy/awesome-terraform",
			// Ansible
			"ansible-community/awesome-ansible",
			// Prometheus
			"roaldnefs/awesome-prometheus",
			// Grafana
			"grafana/awesome-grafana",
			// Nginx
			"fcambus/nginx-resources",
			// Rust (more)
			"rust-unofficial/awesome-rust",
			"burntsushi/awesome-rust",
			// Elixir (more)
			"h4cc/awesome-elixir",
			// Scala (more)
			"lauris/awesome-scala",
			// Dart / Flutter (more)
			"yissachar/awesome-dart",
			// R
			"qinwf/awesome-R",
			// Julia
			"svaksha/Julia.jl",
			// Perl
			"hachiojipm/awesome-perl",
			// MATLAB
			"mikecroucher/awesome-MATLAB",
			// DevOps (more)
			"joefitzgerald/awesome-devops",
			// Cloud Native
			"burib/awesome-cloud-native",
			// Git
			"dtinth/awesome-git",
			// Regex
			"aloisdg/awesome-regex",
			// Unicode
			"jawb/awesome-unicode",
			// Accessibility
			"brunopulis/awesome-a11y",
			// Internationalization
			"localize-js/awesome-i18n",
			// Performance
			"nbubna/awesome-performance",
			// Progressive Web Apps
			"nicokoenig/awesome-pwa",
			// Web Components
			"obetomuniz/awesome-webcomponents",
			// CSS
			"awesome-css-group/awesome-css",
			// SVG
			"willianjusten/awesome-svg",
			// Icons
			"notlmn/awesome-icons",
			// Colors
			"crazyhitty/awesome-colors",
			// Fonts (more)
			"brabadu/awesome-fonts",
			// Typography
			"deanhume/typography",
			// Email
			"jonathandion/awesome-emails",
			// Chat
			"nicolaiarocci/awesome-chatgpt",
			// No-Code / Low-Code
			"valentin-vignal/awesome-no-code",
			// CMS (more)
			"parro-it/awesome-cms",
			// Static Site (more)
			"agarrharr/awesome-static-website-services",
			// Headless CMS
			"geekplux/awesome-headless-cms",
			// API Design
			"grapis/awesome-api-design",
			// REST
			"marmelab/awesome-rest",
			// GraphQL (more)
			"chentsulin/awesome-graphql",
			// WebAssembly (more)
			"mbasso/awesome-wasm",
			// Web Scraping (more)
			"lorien/awesome-web-scraping",
			// Browser Extensions
			"bfred-it/Awesome-WebExtensions",
			// Electron
			"sindresorhus/awesome-electron",
			// Tauri
			"tauri-apps/awesome-tauri",
			// Deno
			"denolib/awesome-deno",
			// Bun
			"omrilotan/awesome-bun",
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
			{Topic: "rust", MinStars: 100},
			{Topic: "kubernetes", MinStars: 100},
			{Topic: "docker", MinStars: 100},
			{Topic: "database", MinStars: 100},
			{Topic: "react", MinStars: 100},
			{Topic: "vue", MinStars: 100},
			{Topic: "flutter", MinStars: 100},
			{Topic: "android", MinStars: 100},
			{Topic: "ios", MinStars: 100},
			{Topic: "web3", MinStars: 100},
			{Topic: "game-engine", MinStars: 100},
			{Topic: "cli", MinStars: 100},
			{Topic: "visualization", MinStars: 100},
			{Topic: "deep-learning", MinStars: 100},
			{Topic: "nlp", MinStars: 100},
			{Topic: "computer-vision", MinStars: 100},
			{Topic: "mcp", MinStars: 50},
			{Topic: "function-calling", MinStars: 50},
			{Topic: "diffusion-model", MinStars: 50},
			{Topic: "text-to-speech", MinStars: 100},
			{Topic: "speech-recognition", MinStars: 100},
			// New topics for broader coverage
			{Topic: "typescript", MinStars: 100},
			{Topic: "java", MinStars: 500},
			{Topic: "spring-boot", MinStars: 100},
			{Topic: "kotlin", MinStars: 100},
			{Topic: "swift", MinStars: 100},
			{Topic: "csharp", MinStars: 100},
			{Topic: "php", MinStars: 100},
			{Topic: "ruby", MinStars: 100},
			{Topic: "scala", MinStars: 100},
			{Topic: "elixir", MinStars: 100},
			{Topic: "haskell", MinStars: 100},
			{Topic: "zig", MinStars: 100},
			{Topic: "lua", MinStars: 100},
			{Topic: "cpp", MinStars: 100},
			{Topic: "embedded", MinStars: 100},
			{Topic: "iot", MinStars: 100},
			{Topic: "monitoring", MinStars: 100},
			{Topic: "observability", MinStars: 100},
			{Topic: "testing", MinStars: 100},
			{Topic: "api", MinStars: 100},
			{Topic: "microservice", MinStars: 100},
			{Topic: "graphql", MinStars: 100},
			{Topic: "grpc", MinStars: 100},
			{Topic: "webassembly", MinStars: 100},
			{Topic: "compiler", MinStars: 100},
			{Topic: "parser", MinStars: 100},
			{Topic: "editor", MinStars: 100},
			{Topic: "neovim", MinStars: 100},
			{Topic: "terminal", MinStars: 100},
			{Topic: "ssh", MinStars: 100},
			{Topic: "proxy", MinStars: 100},
			{Topic: "vpn", MinStars: 100},
			{Topic: "dns", MinStars: 100},
			{Topic: "load-balancer", MinStars: 100},
			{Topic: "message-queue", MinStars: 100},
			{Topic: "cache", MinStars: 100},
			{Topic: "search-engine", MinStars: 100},
			{Topic: "object-storage", MinStars: 100},
			{Topic: "blockchain", MinStars: 100},
			{Topic: "cryptocurrency", MinStars: 100},
			{Topic: "defi", MinStars: 100},
			{Topic: "nft", MinStars: 100},
			{Topic: "game-development", MinStars: 100},
			{Topic: "audio", MinStars: 100},
			{Topic: "music", MinStars: 100},
			{Topic: "image-processing", MinStars: 100},
			{Topic: "video", MinStars: 100},
			{Topic: "pdf", MinStars: 100},
			{Topic: "ocr", MinStars: 100},
			{Topic: "automation", MinStars: 100},
			{Topic: "scraping", MinStars: 100},
			{Topic: "cms", MinStars: 100},
			{Topic: "ecommerce", MinStars: 100},
			{Topic: "authentication", MinStars: 100},
			{Topic: "authorization", MinStars: 100},
			{Topic: "encryption", MinStars: 100},
			{Topic: "static-site-generator", MinStars: 100},
			{Topic: "framework", MinStars: 500},
			{Topic: "library", MinStars: 500},
			{Topic: "tool", MinStars: 500},
			{Topic: "reinforcement-learning", MinStars: 100},
			{Topic: "generative-ai", MinStars: 100},
			{Topic: "stable-diffusion", MinStars: 100},
			{Topic: "text-to-image", MinStars: 100},
			{Topic: "code-generation", MinStars: 100},
			{Topic: "chatbot", MinStars: 100},
			// === 新增 Topics ===
			// Science & Research
			{Topic: "bioinformatics", MinStars: 100},
			{Topic: "computational-biology", MinStars: 100},
			{Topic: "quantum-computing", MinStars: 100},
			{Topic: "scientific-computing", MinStars: 100},
			{Topic: "simulation", MinStars: 100},
			{Topic: "physics", MinStars: 100},
			{Topic: "chemistry", MinStars: 100},
			// Finance
			{Topic: "finance", MinStars: 100},
			{Topic: "trading", MinStars: 100},
			{Topic: "quant", MinStars: 100},
			{Topic: "algorithmic-trading", MinStars: 100},
			{Topic: "risk-management", MinStars: 100},
			// GIS / Maps
			{Topic: "gis", MinStars: 100},
			{Topic: "geospatial", MinStars: 100},
			{Topic: "mapping", MinStars: 100},
			// Robotics
			{Topic: "robotics", MinStars: 100},
			{Topic: "ros", MinStars: 100},
			{Topic: "autonomous-driving", MinStars: 100},
			// Hardware
			{Topic: "fpga", MinStars: 100},
			{Topic: "verilog", MinStars: 100},
			{Topic: "risc-v", MinStars: 100},
			{Topic: "arm", MinStars: 100},
			// Operating Systems
			{Topic: "operating-system", MinStars: 100},
			{Topic: "linux-kernel", MinStars: 100},
			{Topic: "filesystem", MinStars: 100},
			// Networking (more)
			{Topic: "tcp", MinStars: 100},
			{Topic: "http", MinStars: 100},
			{Topic: "websocket", MinStars: 100},
			{Topic: "p2p", MinStars: 100},
			{Topic: "webrtc", MinStars: 100},
			{Topic: "sdn", MinStars: 100},
			// Distributed Systems
			{Topic: "consensus", MinStars: 100},
			{Topic: "raft", MinStars: 100},
			{Topic: "distributed-database", MinStars: 100},
			{Topic: "distributed-system", MinStars: 100},
			// Data Engineering (more)
			{Topic: "etl", MinStars: 100},
			{Topic: "data-pipeline", MinStars: 100},
			{Topic: "data-lake", MinStars: 100},
			{Topic: "data-warehouse", MinStars: 100},
			{Topic: "olap", MinStars: 100},
			// DevTools (more)
			{Topic: "linter", MinStars: 100},
			{Topic: "formatter", MinStars: 100},
			{Topic: "debugger", MinStars: 100},
			{Topic: "profiler", MinStars: 100},
			{Topic: "build-tool", MinStars: 100},
			{Topic: "package-manager", MinStars: 100},
			{Topic: "dependency-management", MinStars: 100},
			// Education
			{Topic: "education", MinStars: 100},
			{Topic: "tutorial", MinStars: 100},
			{Topic: "course", MinStars: 100},
			// Design
			{Topic: "design-system", MinStars: 100},
			{Topic: "ui-components", MinStars: 100},
			{Topic: "icon", MinStars: 100},
			{Topic: "color", MinStars: 100},
			// Media
			{Topic: "image-processing", MinStars: 100},
			{Topic: "video-processing", MinStars: 100},
			{Topic: "audio-processing", MinStars: 100},
			{Topic: "streaming", MinStars: 100},
			{Topic: "transcoding", MinStars: 100},
			// Privacy & Security (more)
			{Topic: "privacy", MinStars: 100},
			{Topic: "penetration-testing", MinStars: 100},
			{Topic: "vulnerability", MinStars: 100},
			{Topic: "malware", MinStars: 100},
			{Topic: "forensics", MinStars: 100},
			{Topic: "zero-trust", MinStars: 100},
			// Cloud (more)
			{Topic: "aws", MinStars: 100},
			{Topic: "azure", MinStars: 100},
			{Topic: "gcp", MinStars: 100},
			{Topic: "cloud-computing", MinStars: 100},
			{Topic: "multi-cloud", MinStars: 100},
			// Mobile (more)
			{Topic: "mobile", MinStars: 100},
			{Topic: "cross-platform", MinStars: 100},
			{Topic: "react-native", MinStars: 100},
			{Topic: "maui", MinStars: 100},
			// AI Applications
			{Topic: "text-to-speech", MinStars: 100},
			{Topic: "speech-to-text", MinStars: 100},
			{Topic: "text-to-image", MinStars: 100},
			{Topic: "image-generation", MinStars: 100},
			{Topic: "video-generation", MinStars: 100},
			{Topic: "music-generation", MinStars: 100},
			{Topic: "code-ai", MinStars: 100},
			{Topic: "ai-agent", MinStars: 100},
			{Topic: "vector-database", MinStars: 100},
			{Topic: "embedding-model", MinStars: 100},
			{Topic: "fine-tuning", MinStars: 100},
			{Topic: "model-serving", MinStars: 100},
			{Topic: "inference-engine", MinStars: 100},
			{Topic: "prompt-engineering", MinStars: 100},
			// Web (more)
			{Topic: "web-framework", MinStars: 100},
			{Topic: "http-server", MinStars: 100},
			{Topic: "middleware", MinStars: 100},
			{Topic: "session", MinStars: 100},
			{Topic: "cors", MinStars: 100},
			// Math
			{Topic: "linear-algebra", MinStars: 100},
			{Topic: "optimization", MinStars: 100},
			{Topic: "statistics", MinStars: 100},
			{Topic: "probability", MinStars: 100},
			// Game Dev (more)
			{Topic: "game-engine", MinStars: 100},
			{Topic: "unity3d", MinStars: 100},
			{Topic: "unreal-engine", MinStars: 100},
			{Topic: "godot", MinStars: 100},
			{Topic: "3d", MinStars: 100},
			{Topic: "opengl", MinStars: 100},
			{Topic: "vulkan", MinStars: 100},
			{Topic: "directx", MinStars: 100},
			// Misc
			{Topic: "homelab", MinStars: 100},
			{Topic: "self-hosted", MinStars: 100},
			{Topic: "backup", MinStars: 100},
			{Topic: "sync", MinStars: 100},
			{Topic: "file-manager", MinStars: 100},
			{Topic: "note-taking", MinStars: 100},
			{Topic: "knowledge-base", MinStars: 100},
			{Topic: "wiki", MinStars: 100},
			{Topic: "calendar", MinStars: 100},
			{Topic: "email", MinStars: 100},
			{Topic: "rss", MinStars: 100},
			{Topic: "podcast", MinStars: 100},
		},
	}
}
