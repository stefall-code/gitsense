package service

// EcosystemRule 定义技术生态分类规则
type EcosystemRule struct {
	Name     string   // 生态名称
	Keywords []string // description 中匹配的关键词
	Topics   []string // topics 中匹配的关键词
}

// DefaultEcosystemRules 默认生态分类规则（MVP 内置）
var DefaultEcosystemRules = []EcosystemRule{
	{
		Name:     "AI Agent Infrastructure",
		Keywords: []string{"ai agent", "agent framework", "llm agent"},
		Topics:   []string{"llm", "ai-agent", "agent", "ai-gateway", "ai-framework"},
	},
	{
		Name:     "LLM Gateway & Proxy",
		Keywords: []string{"llm gateway", "llm proxy", "ai gateway", "model router"},
		Topics:   []string{"llm", "gateway", "proxy", "ai-gateway", "model-routing"},
	},
	{
		Name:     "Python Web Framework",
		Keywords: []string{"python web", "async framework"},
		Topics:   []string{"python", "web", "framework", "fastapi", "flask", "django", "asgi"},
	},
	{
		Name:     "Go Web Framework",
		Keywords: []string{"go web", "http framework"},
		Topics:   []string{"go", "web", "framework", "gin", "echo", "fiber", "chi"},
	},
	{
		Name:     "Frontend Framework",
		Keywords: []string{"frontend framework", "ui library"},
		Topics:   []string{"react", "vue", "frontend", "ui", "angular", "svelte"},
	},
	{
		Name:     "Database & ORM",
		Keywords: []string{"database", "orm", "sql"},
		Topics:   []string{"database", "orm", "sql", "postgres", "mysql", "sqlite"},
	},
	{
		Name:     "DevOps & Infrastructure",
		Keywords: []string{"devops", "infrastructure", "ci/cd"},
		Topics:   []string{"kubernetes", "docker", "devops", "cloud", "terraform", "ci-cd"},
	},
	{
		Name:     "Observability & Monitoring",
		Keywords: []string{"monitoring", "observability", "tracing"},
		Topics:   []string{"monitoring", "observability", "tracing", "metrics", "logging"},
	},
	{
		Name:     "Data Engineering",
		Keywords: []string{"data pipeline", "etl", "data processing"},
		Topics:   []string{"data", "etl", "pipeline", "streaming", "kafka", "spark"},
	},
	{
		Name:     "Security & Auth",
		Keywords: []string{"security", "authentication", "authorization"},
		Topics:   []string{"security", "auth", "oauth", "jwt", "encryption"},
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

// Classify 根据仓库的 topics 和 description 分类生态
// 返回匹配度最高的生态名称，无匹配返回空字符串
func (c *EcosystemClassifier) Classify(topics []string, description string) string {
	topicSet := make(map[string]bool, len(topics))
	for _, t := range topics {
		topicSet[t] = true
	}

	bestMatch := ""
	bestScore := 0

	for _, rule := range c.rules {
		score := 0

		// Topic 匹配（权重更高）
		for _, t := range rule.Topics {
			if topicSet[t] {
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
	// Fallback: 使用第一个 topic + "Ecosystem"
	if len(topics) > 0 {
		return topics[0] + " Ecosystem"
	}
	return "Unknown Ecosystem"
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
