package service

// SubcategoryRule 定义生态内的子分类规则
type SubcategoryRule struct {
	Name   string   // 子分类名称
	Topics []string // 该子分类的 topic 特征
}

// EcosystemRule 定义技术生态分类规则
type EcosystemRule struct {
	Name          string            // 生态名称
	Keywords      []string          // description 中匹配的关键词
	Topics        []string          // topics 中匹配的关键词
	Subcategories []SubcategoryRule // 生态内子分类
}

// DefaultEcosystemRules 默认生态分类规则（MVP 内置）
var DefaultEcosystemRules = []EcosystemRule{
	{
		Name:     "AI Agent Infrastructure",
		Keywords: []string{"ai agent", "agent framework", "llm agent"},
		Topics:   []string{"llm", "ai-agent", "agent", "agents", "ai-agents", "ai-gateway", "ai-framework", "langchain", "langgraph", "autogen", "crewai", "rag", "generative-ai", "multiagent", "ai", "chatbot", "openai", "gpt", "prompt", "llamaindex"},
		Subcategories: []SubcategoryRule{
			{Name: "Agent Framework", Topics: []string{"agent", "ai-agent", "langchain", "crewai", "autogen", "agentic", "ai-framework", "agent-framework"}},
			{Name: "Vector Database", Topics: []string{"vector", "vector-database", "embedding", "qdrant", "milvus", "chromadb", "pinecone", "weaviate"}},
			{Name: "Serving & Inference", Topics: []string{"inference", "vllm", "ollama", "serving", "llm-inference", "model-serving"}},
			{Name: "MCP & Tooling", Topics: []string{"mcp", "tool", "model-context-protocol", "ai-tool", "function-calling"}},
			{Name: "Agent UI", Topics: []string{"agent-ui", "chatbot", "open-webui", "chatbot-ui", "ai-chat"}},
		},
	},
	{
		Name:     "LLM & NLP",
		Keywords: []string{"llm", "nlp", "language model", "transformer", "text generation"},
		Topics:   []string{"nlp", "natural-language-processing", "transformer", "llm", "language-model", "text-generation", "speech-recognition", "text-to-speech", "translation", "tokenization", "bert", "gpt", "huggingface", "diffusers", "stable-diffusion", "text-to-image", "image-generation"},
		Subcategories: []SubcategoryRule{
			{Name: "NLP", Topics: []string{"nlp", "natural-language-processing", "transformer", "bert", "tokenization", "spacy"}},
			{Name: "Generative AI", Topics: []string{"diffusers", "stable-diffusion", "text-to-image", "image-generation", "gan", "generative"}},
			{Name: "Speech", Topics: []string{"speech-recognition", "text-to-speech", "whisper", "tts", "asr"}},
		},
	},
	{
		Name:     "Machine Learning",
		Keywords: []string{"machine learning", "deep learning", "ml framework"},
		Topics:   []string{"machine-learning", "deep-learning", "ml", "neural-network", "tensorflow", "pytorch", "keras", "scikit-learn", "xgboost", "lightgbm", "catboost", "model-training", "training", "onnx", "model-optimization", "reinforcement-learning"},
		Subcategories: []SubcategoryRule{
			{Name: "DL Framework", Topics: []string{"pytorch", "tensorflow", "keras", "deep-learning", "neural-network"}},
			{Name: "Classical ML", Topics: []string{"scikit-learn", "xgboost", "lightgbm", "catboost", "random-forest", "gradient-boosting"}},
			{Name: "MLOps", Topics: []string{"mlflow", "wandb", "experiment-tracking", "model-registry", "model-serving"}},
		},
	},
	{
		Name:     "Data Science & Analytics",
		Keywords: []string{"data science", "analytics", "visualization", "jupyter"},
		Topics:   []string{"data-science", "data-analysis", "visualization", "jupyter", "notebook", "pandas", "numpy", "matplotlib", "plotly", "seaborn", "streamlit", "gradio", "dashboard", "bi", "analytics", "statistical-analysis", "r", "d3", "chart", "echarts"},
		Subcategories: []SubcategoryRule{
			{Name: "Data Analysis", Topics: []string{"pandas", "numpy", "data-analysis", "data-science", "statistics"}},
			{Name: "Visualization", Topics: []string{"matplotlib", "plotly", "seaborn", "d3", "chart", "visualization", "echarts", "dashboard"}},
			{Name: "Notebook & App", Topics: []string{"jupyter", "notebook", "streamlit", "gradio", "dash", "panel"}},
		},
	},
	{
		Name:     "LLM Gateway & Proxy",
		Keywords: []string{"llm gateway", "llm proxy", "ai gateway", "model router"},
		Topics:   []string{"llm", "gateway", "proxy", "ai-gateway", "model-routing"},
		Subcategories: []SubcategoryRule{
			{Name: "API Gateway", Topics: []string{"gateway", "api-gateway", "proxy", "reverse-proxy"}},
			{Name: "Model Router", Topics: []string{"model-routing", "model-router", "load-balancer", "llm-router"}},
			{Name: "Auth & Rate Limit", Topics: []string{"rate-limiting", "authentication", "api-key", "quota"}},
		},
	},
	{
		Name:     "Python Web Framework",
		Keywords: []string{"python web", "async framework", "python api"},
		Topics:   []string{"python", "web", "framework", "fastapi", "flask", "django", "asgi", "starlette", "sanic", "tornado", "bottle", "celery", "uvicorn", "gunicorn", "wsgi"},
		Subcategories: []SubcategoryRule{
			{Name: "Full-Stack Framework", Topics: []string{"django", "flask", "fastapi", "pyramid", "sanic"}},
			{Name: "ASGI & Async", Topics: []string{"asgi", "uvicorn", "asyncio", "starlette", "gunicorn"}},
			{Name: "API Framework", Topics: []string{"fastapi", "api", "rest", "openapi", "swagger"}},
			{Name: "Task Queue", Topics: []string{"celery", "rq", "task-queue", "worker", "job-queue"}},
		},
	},
	{
		Name:     "Go Web Framework",
		Keywords: []string{"go web", "http framework", "golang"},
		Topics:   []string{"golang", "web", "framework", "gin", "echo", "fiber", "chi", "mux", "gorilla", "grpc", "protobuf", "go-micro", "kit"},
		Subcategories: []SubcategoryRule{
			{Name: "HTTP Framework", Topics: []string{"gin", "echo", "fiber", "chi", "mux", "gorilla"}},
			{Name: "Microservice", Topics: []string{"microservice", "grpc", "protobuf", "go-micro", "kit"}},
			{Name: "Middleware", Topics: []string{"middleware", "ratelimit", "cors", "auth"}},
		},
	},
	{
		Name:     "Java & JVM Framework",
		Keywords: []string{"java", "spring", "kotlin", "jvm"},
		Topics:   []string{"java", "spring", "spring-boot", "spring-framework", "kotlin", "quarkus", "micronaut", "gradle", "maven", "jvm", "jakarta", "servlet", "hibernate", "jpa"},
		Subcategories: []SubcategoryRule{
			{Name: "Spring", Topics: []string{"spring", "spring-boot", "spring-framework", "spring-security", "spring-cloud"}},
			{Name: "Alternative JVM", Topics: []string{"quarkus", "micronaut", "kotlin", "graalvm"}},
			{Name: "Build & ORM", Topics: []string{"gradle", "maven", "hibernate", "jpa", "orm"}},
		},
	},
	{
		Name:     "Ruby Framework",
		Keywords: []string{"ruby", "rails", "sinatra"},
		Topics:   []string{"ruby", "rails", "sinatra", "ruby-on-rails", "gem", "bundler", "rack"},
		Subcategories: []SubcategoryRule{
			{Name: "Rails", Topics: []string{"rails", "ruby-on-rails", "activerecord"}},
			{Name: "Lightweight", Topics: []string{"sinatra", "rack", "roda"}},
		},
	},
	{
		Name:     "PHP Framework",
		Keywords: []string{"php", "laravel", "symfony"},
		Topics:   []string{"php", "laravel", "symfony", "wordpress", "drupal", "composer", "phpunit"},
		Subcategories: []SubcategoryRule{
			{Name: "Laravel", Topics: []string{"laravel", "blade", "eloquent"}},
			{Name: "Symfony", Topics: []string{"symfony", "doctrine", "twig"}},
			{Name: "CMS", Topics: []string{"wordpress", "drupal", "cms"}},
		},
	},
	{
		Name:     "Frontend Framework",
		Keywords: []string{"frontend framework", "ui library", "web framework"},
		Topics:   []string{"react", "vue", "frontend", "ui", "angular", "svelte", "nextjs", "nuxt", "gatsby", "remix", "solid", "astro", "web", "component", "typescript", "javascript", "ssr", "spa", "webpack", "vite", "esbuild", "babel", "rollup", "parcel"},
		Subcategories: []SubcategoryRule{
			{Name: "UI Framework", Topics: []string{"react", "vue", "angular", "svelte", "solid", "astro"}},
			{Name: "Meta Framework", Topics: []string{"nextjs", "nuxt", "gatsby", "remix", "ssr"}},
			{Name: "Component Library", Topics: []string{"ui-library", "component", "design-system", "tailwind", "material-ui", "ant-design", "chakra", "radix"}},
			{Name: "State Management", Topics: []string{"state", "redux", "zustand", "pinia", "mobx", "recoil"}},
			{Name: "Build Tool", Topics: []string{"webpack", "vite", "esbuild", "babel", "rollup", "parcel", "swc"}},
		},
	},
	{
		Name:     "Mobile Development",
		Keywords: []string{"mobile", "ios", "android", "react native", "flutter"},
		Topics:   []string{"mobile", "ios", "android", "react-native", "flutter", "swift", "kotlin", "swiftui", "compose", "uikit", "xamarin", "capacitor", "ionic", "expo", "native", "app", "kotlin-multiplatform"},
		Subcategories: []SubcategoryRule{
			{Name: "Cross-Platform", Topics: []string{"react-native", "flutter", "xamarin", "capacitor", "ionic", "expo"}},
			{Name: "iOS", Topics: []string{"ios", "swift", "swiftui", "uikit", "alamofire"}},
			{Name: "Android", Topics: []string{"android", "kotlin", "compose", "jetpack", "gradle"}},
		},
	},
	{
		Name:     "Database & ORM",
		Keywords: []string{"database", "orm", "sql", "nosql"},
		Topics:   []string{"database", "orm", "sql", "postgres", "mysql", "sqlite", "mongodb", "redis", "cassandra", "dynamodb", "couchdb", "elasticsearch", "clickhouse", "cockroachdb", "mariadb", "timescaledb", "influxdb", "neo4j", "graphql", "prisma", "sqlalchemy", "gorm", "sequelize", "typeorm", "migration", "query-builder"},
		Subcategories: []SubcategoryRule{
			{Name: "Relational DB", Topics: []string{"postgres", "mysql", "sqlite", "mariadb", "cockroachdb", "sql"}},
			{Name: "ORM & Query", Topics: []string{"orm", "sql", "query-builder", "prisma", "sqlalchemy", "gorm", "sequelize", "typeorm", "graphql"}},
			{Name: "NoSQL", Topics: []string{"mongodb", "redis", "cassandra", "dynamodb", "couchdb", "neo4j"}},
			{Name: "Search & Analytics DB", Topics: []string{"elasticsearch", "clickhouse", "timescaledb", "influxdb", "opensearch"}},
		},
	},
	{
		Name:     "DevOps & Infrastructure",
		Keywords: []string{"devops", "infrastructure", "ci/cd", "deployment"},
		Topics:   []string{"kubernetes", "docker", "devops", "cloud", "terraform", "ci-cd", "ansible", "helm", "container", "containers", "deployment", "infrastructure", "automation", "configuration", "packer", "nomad", "vagrant", "argo", "gitops", "kustomize", "pulumi", "cncf", "cloud-native"},
		Subcategories: []SubcategoryRule{
			{Name: "Container & Orchestration", Topics: []string{"kubernetes", "docker", "container", "helm", "podman", "containerd"}},
			{Name: "IaC & Config", Topics: []string{"terraform", "ansible", "cloudformation", "infrastructure-as-code", "pulumi", "puppet", "chef"}},
			{Name: "CI/CD", Topics: []string{"ci-cd", "github-actions", "jenkins", "gitlab-ci", "pipeline", "argo", "tekton", "spinnaker"}},
			{Name: "GitOps", Topics: []string{"gitops", "flux", "argocd", "kustomize"}},
		},
	},
	{
		Name:     "Observability & Monitoring",
		Keywords: []string{"monitoring", "observability", "tracing", "logging"},
		Topics:   []string{"monitoring", "observability", "tracing", "metrics", "logging", "grafana", "prometheus", "opentelemetry", "jaeger", "zipkin", "loki", "fluentd", "fluent-bit", "alerting", "alertmanager", "datadog", "kibana", "logstash", "telegraf", "statsd", "apm"},
		Subcategories: []SubcategoryRule{
			{Name: "Tracing", Topics: []string{"tracing", "opentelemetry", "jaeger", "zipkin", "distributed-tracing", "apm"}},
			{Name: "Metrics", Topics: []string{"metrics", "prometheus", "grafana", "monitoring", "telegraf", "statsd"}},
			{Name: "Logging", Topics: []string{"logging", "log", "elk", "fluentd", "loki", "kibana", "logstash", "fluent-bit"}},
		},
	},
	{
		Name:     "Data Engineering",
		Keywords: []string{"data pipeline", "etl", "data processing", "data warehouse"},
		Topics:   []string{"data", "etl", "pipeline", "streaming", "kafka", "spark", "airflow", "dbt", "flink", "pulsar", "hadoop", "data-lake", "data-warehouse", "data-pipeline", "batch-processing", "data-ingestion", "sqoop", "hive", "presto", "trino", "snowflake", "bigquery", "duckdb", "polars"},
		Subcategories: []SubcategoryRule{
			{Name: "Stream Processing", Topics: []string{"kafka", "streaming", "flink", "spark-streaming", "pulsar", "rabbitmq", "mq"}},
			{Name: "Batch Processing", Topics: []string{"spark", "hadoop", "etl", "data-processing", "batch", "hive", "presto", "trino"}},
			{Name: "Data Lake & Warehouse", Topics: []string{"data-lake", "delta-lake", "iceberg", "parquet", "warehouse", "snowflake", "bigquery", "duckdb", "dbt"}},
			{Name: "Orchestration", Topics: []string{"airflow", "dagster", "prefect", "workflow", "orchestration", "scheduler"}},
		},
	},
	{
		Name:     "Security & Auth",
		Keywords: []string{"security", "authentication", "authorization", "encryption"},
		Topics:   []string{"security", "auth", "oauth", "jwt", "encryption", "authentication", "authorization", "keycloak", "vault", "sso", "openid", "rbac", "tls", "certificate", "crypto", "pentest", "vulnerability", "owasp", "trivy", "falco", "cve", "secret", "credential"},
		Subcategories: []SubcategoryRule{
			{Name: "Authentication", Topics: []string{"oauth", "jwt", "authentication", "sso", "openid", "keycloak", "auth"}},
			{Name: "Authorization", Topics: []string{"rbac", "authorization", "abac", "policy", "permission", "casbin"}},
			{Name: "Encryption & Secrets", Topics: []string{"encryption", "tls", "certificate", "crypto", "vault", "secret"}},
			{Name: "Security Scanning", Topics: []string{"trivy", "falco", "vulnerability", "owasp", "pentest", "cve", "sast", "dast"}},
		},
	},
	{
		Name:     "Networking & Proxy",
		Keywords: []string{"proxy", "load balancer", "networking", "cdn"},
		Topics:   []string{"nginx", "proxy", "load-balancer", "cdn", "dns", "network", "http", "tcp", "udp", "caddy", "traefik", "envoy", "haproxy", "vpn", "tunnel", "wireguard", "cloudflare"},
		Subcategories: []SubcategoryRule{
			{Name: "Web Server", Topics: []string{"nginx", "apache", "caddy", "http-server"}},
			{Name: "Reverse Proxy", Topics: []string{"proxy", "reverse-proxy", "traefik", "envoy", "haproxy", "load-balancer"}},
			{Name: "Network & VPN", Topics: []string{"vpn", "tunnel", "wireguard", "dns", "network", "cloudflare"}},
		},
	},
	{
		Name:     "Messaging & Queue",
		Keywords: []string{"message queue", "mq", "pub/sub", "event streaming"},
		Topics:   []string{"kafka", "rabbitmq", "mqtt", "amqp", "pubsub", "message-queue", "event-driven", "nats", "zeromq", "pulsar", "redis", "sqs", "event-streaming", "celery", "bull"},
		Subcategories: []SubcategoryRule{
			{Name: "Message Broker", Topics: []string{"kafka", "rabbitmq", "pulsar", "nats", "mqtt", "amqp"}},
			{Name: "Task Queue", Topics: []string{"celery", "bull", "rq", "sidekiq", "task-queue"}},
		},
	},
	{
		Name:     "Testing Framework",
		Keywords: []string{"testing", "test framework", "unit test"},
		Topics:   []string{"testing", "test", "unit-test", "integration-test", "e2e", "mock", "tdd", "jest", "pytest", "junit", "cypress", "playwright", "selenium", "testify", "goconvey", "mocha", "chai", "cucumber", "bdd"},
		Subcategories: []SubcategoryRule{
			{Name: "Unit Testing", Topics: []string{"jest", "pytest", "junit", "testify", "mocha", "unit-test"}},
			{Name: "E2E Testing", Topics: []string{"cypress", "playwright", "selenium", "e2e", "puppeteer"}},
			{Name: "BDD & Mock", Topics: []string{"cucumber", "bdd", "mock", "stub", "faker"}},
		},
	},
	{
		Name:     "Cloud & Serverless",
		Keywords: []string{"cloud", "serverless", "aws", "azure", "gcp"},
		Topics:   []string{"aws", "azure", "gcp", "cloud", "serverless", "lambda", "cloudformation", "sam", "cdk", "terraform", "iam", "s3", "ec2", "kubernetes", "helm", "cloud-native"},
		Subcategories: []SubcategoryRule{
			{Name: "AWS", Topics: []string{"aws", "lambda", "s3", "ec2", "dynamodb", "cloudformation", "sam", "cdk"}},
			{Name: "Azure", Topics: []string{"azure", "azure-functions", "azure-sdk"}},
			{Name: "GCP", Topics: []string{"gcp", "google-cloud", "bigquery", "gke"}},
			{Name: "Serverless", Topics: []string{"serverless", "lambda", "faas", "edge"}},
		},
	},
	{
		Name:     "Blockchain & Web3",
		Keywords: []string{"blockchain", "web3", "smart contract", "ethereum", "defi"},
		Topics:   []string{"blockchain", "web3", "ethereum", "solidity", "smart-contract", "defi", "nft", "crypto", "bitcoin", "solana", "polygon", "ipfs", "dao", "token", "wallet", "dapp", "evm"},
		Subcategories: []SubcategoryRule{
			{Name: "Ethereum", Topics: []string{"ethereum", "solidity", "evm", "smart-contract", "web3"}},
			{Name: "DeFi", Topics: []string{"defi", "token", "swap", "lending", "yield"}},
			{Name: "Other Chain", Topics: []string{"solana", "bitcoin", "polygon", "cosmos", "substrate"}},
		},
	},
	{
		Name:     "Game Development",
		Keywords: []string{"game engine", "game development", "gamedev"},
		Topics:   []string{"game", "game-engine", "gamedev", "unity", "unreal", "godot", "bevy", "3d", "2d", "opengl", "vulkan", "directx", "physics", "rendering", "voxel"},
		Subcategories: []SubcategoryRule{
			{Name: "Game Engine", Topics: []string{"unity", "unreal", "godot", "bevy", "game-engine"}},
			{Name: "Graphics", Topics: []string{"opengl", "vulkan", "directx", "rendering", "shader", "3d"}},
		},
	},
	{
		Name:     "Embedded & IoT",
		Keywords: []string{"embedded", "iot", "hardware", "firmware"},
		Topics:   []string{"embedded", "iot", "hardware", "firmware", "arduino", "raspberry-pi", "rtos", "sensors", "mqtt", "zigbee", "bluetooth", "esp32", "stm32", "microcontroller"},
		Subcategories: []SubcategoryRule{
			{Name: "Embedded OS", Topics: []string{"rtos", "embedded", "freertos", "zephyr"}},
			{Name: "IoT Platform", Topics: []string{"iot", "mqtt", "sensors", "home-assistant", "zigbee"}},
			{Name: "Hardware", Topics: []string{"arduino", "raspberry-pi", "esp32", "stm32", "fpga"}},
		},
	},
	{
		Name:     "Systems Programming",
		Keywords: []string{"systems programming", "operating system", "compiler", "low-level"},
		Topics:   []string{"rust", "c", "cpp", "systems", "os", "kernel", "compiler", "linker", "assembler", "llvm", "wasm", "webassembly", "assembly", "x86", "arm", "riscv"},
		Subcategories: []SubcategoryRule{
			{Name: "OS & Kernel", Topics: []string{"os", "kernel", "linux", "x86", "arm", "riscv"}},
			{Name: "Compiler", Topics: []string{"compiler", "llvm", "parser", "interpreter", "linker"}},
			{Name: "WASM", Topics: []string{"wasm", "webassembly", "wasi"}},
		},
	},
	{
		Name:     "Content Management",
		Keywords: []string{"cms", "content management", "blog", "static site"},
		Topics:   []string{"cms", "content-management", "blog", "static-site", "hugo", "jekyll", "gatsby", "nextjs", "wordpress", "strapi", "ghost", "netlify-cms", "headless-cms"},
		Subcategories: []SubcategoryRule{
			{Name: "Static Site", Topics: []string{"hugo", "jekyll", "static-site", "11ty", "astro"}},
			{Name: "Headless CMS", Topics: []string{"strapi", "ghost", "netlify-cms", "headless-cms", "sanity", "contentful"}},
		},
	},
	{
		Name:     "API & Protocol",
		Keywords: []string{"api", "graphql", "grpc", "rest", "protocol"},
		Topics:   []string{"api", "graphql", "grpc", "rest", "openapi", "swagger", "protobuf", "rpc", "websocket", "http", "protocol", "soap", "api-gateway", "api-management"},
		Subcategories: []SubcategoryRule{
			{Name: "GraphQL", Topics: []string{"graphql", "apollo", "relay", "hasura", "prisma"}},
			{Name: "gRPC & RPC", Topics: []string{"grpc", "protobuf", "rpc", "thrift"}},
			{Name: "REST & OpenAPI", Topics: []string{"rest", "openapi", "swagger", "api-design"}},
		},
	},
	{
		Name:     "CLI & Developer Tools",
		Keywords: []string{"cli", "command line", "developer tools", "terminal"},
		Topics:   []string{"cli", "command-line", "terminal", "shell", "bash", "zsh", "prompt", "tui", "ncurses", "developer-tools", "devtools", "debugger", "profiler", "linter", "formatter"},
		Subcategories: []SubcategoryRule{
			{Name: "Shell & Prompt", Topics: []string{"shell", "bash", "zsh", "prompt", "fish"}},
			{Name: "TUI", Topics: []string{"tui", "ncurses", "terminal", "terminal-emulator"}},
			{Name: "Dev Tools", Topics: []string{"linter", "formatter", "debugger", "profiler", "cli"}},
		},
	},
}

// EcosystemClassifier 技术生态分类器
type EcosystemClassifier struct {
	rules []EcosystemRule
}

// NewEcosystemClassifier 创建分类器
func NewEcosystemClassifier(rules []EcosystemRule) *EcosystemClassifier {
	if rules == nil {
		rules = DefaultEcosystemRules
	}
	return &EcosystemClassifier{rules: rules}
}

// noiseTopics 不应参与生态命名的噪声 topic
var noiseTopics = map[string]bool{
	"hacktoberfest": true, "good-first-issue": true, "help-wanted": true,
	"documentation": true, "beginner-friendly": true, "up-for-grabs": true,
	"question": true, "wontfix": true, "invalid": true, "duplicate": true,
	"enhancement": true, "bug": true, "priority-high": true, "priority-low": true,
	"stale": true, "area/help-wanted": true, "hacktoberfest-accepted": true,
	"good-first-issue-2": true, "ready": true, "approved": true,
}

// normalizeTopic 归一化 topic：小写 + 去连字符/下划线
func normalizeTopic(s string) string {
	s = toLower(s)
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' || c == '_' || c == ' ' {
			continue
		}
		result = append(result, c)
	}
	return string(result)
}

// Classify 根据仓库的 topics 和 description 分类生态
// 返回匹配度最高的生态名称，无匹配返回空字符串
func (c *EcosystemClassifier) Classify(topics []string, description string) string {
	// 归一化 repo topics 用于匹配
	topicSet := make(map[string]bool, len(topics))
	for _, t := range topics {
		topicSet[normalizeTopic(t)] = true
	}

	bestMatch := ""
	bestScore := 0

	for _, rule := range c.rules {
		score := 0

		// Topic 匹配（权重更高），归一化后匹配
		for _, t := range rule.Topics {
			if topicSet[normalizeTopic(t)] {
				score += 2
			}
		}

		// Keyword 匹配（description 中出现）
		if description != "" {
			desc := toLower(description)
			for _, kw := range rule.Keywords {
				if containsSubstring(desc, toLower(kw)) {
					score += 1
				}
			}
		}

		if score > bestScore {
			bestScore = score
			bestMatch = rule.Name
		}
	}

	// 至少需要 2 分才认为匹配（1 个 topic 或 2 个 keyword）
	if bestScore < 2 {
		return ""
	}

	return bestMatch
}

// ClassifyWithFallback 分类，带 fallback
func (c *EcosystemClassifier) ClassifyWithFallback(topics []string, description string) string {
	result := c.Classify(topics, description)
	if result != "" {
		return result
	}
	// Fallback: 使用第一个非 noise topic
	for _, t := range topics {
		if !noiseTopics[toLower(t)] {
			return t + " Ecosystem"
		}
	}
	return "Unknown Ecosystem"
}

// ClassifyByLanguage 基于 programming language 的 fallback 分类
func (c *EcosystemClassifier) ClassifyByLanguage(language string) string {
	langMap := map[string]string{
		"javascript": "Frontend Framework",
		"typescript": "Frontend Framework",
		"python":     "Python Web Framework",
		"go":         "Go Web Framework",
		"java":       "Java & JVM Framework",
		"ruby":       "Ruby Framework",
		"php":        "PHP Framework",
		"rust":       "Systems Programming",
		"c":          "Systems Programming",
		"c++":        "Systems Programming",
		"swift":      "Mobile Development",
		"kotlin":     "Mobile Development",
		"dart":       "Mobile Development",
		"shell":      "CLI & Developer Tools",
	}
	if eco, ok := langMap[toLower(language)]; ok {
		return eco
	}
	return "Unknown Ecosystem"
}

// ClassifySubcategory 分类子分类
// 返回 repo 在指定生态内的子分类名称，无匹配返回 "Other"
func (c *EcosystemClassifier) ClassifySubcategory(topics []string, ecosystem string) string {
	topicSet := make(map[string]bool, len(topics))
	for _, t := range topics {
		topicSet[t] = true
	}

	for _, rule := range c.rules {
		if rule.Name != ecosystem {
			continue
		}
		bestMatch := ""
		bestScore := 0
		for _, sub := range rule.Subcategories {
			score := 0
			for _, t := range sub.Topics {
				if topicSet[t] {
					score++
				}
			}
			if score > bestScore {
				bestScore = score
				bestMatch = sub.Name
			}
		}
		if bestScore > 0 {
			return bestMatch
		}
		break
	}
	return "Other"
}

// GetRule 获取指定生态的规则
func (c *EcosystemClassifier) GetRule(ecosystem string) *EcosystemRule {
	for i := range c.rules {
		if c.rules[i].Name == ecosystem {
			return &c.rules[i]
		}
	}
	return nil
}

// GetAllRules 获取所有规则
func (c *EcosystemClassifier) GetAllRules() []EcosystemRule {
	return c.rules
}

func toLower(s string) string {
	// 简单 ASCII toLower
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func containsSubstring(s, substr string) bool {
	return len(substr) <= len(s) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
