package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// Client 封装 GitHub REST API v3 调用
// 包含限流保护、重试、指数退避
// 内部分离 Core API 与 Search API Rate Limiter
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string

	// Core API 限流 (repos, readmes, topics, languages)
	coreLimitRemaining int
	coreLimitReset     time.Time

	// Search API 限流 (search/repositories)
	searchLimitRemaining int
	searchLimitReset     time.Time

	// 重试
	maxRetries int
}

// NewClient 创建新的 GitHub API Client
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:      token,
		baseURL:    "https://api.github.com",
		maxRetries: 3,
	}
}

// RepoInfo GitHub 仓库基本信息
type RepoInfo struct {
	FullName    string   `json:"full_name"`
	Owner       string   `json:"login"`  // 从 owner.login 解析
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Stars       int      `json:"stargazers_count"`
	Topics      []string `json:"topics"`
	UpdatedAt   string   `json:"updated_at"`  // GitHub updated_at (综合活跃度)
	PushedAt    string   `json:"pushed_at"`   // GitHub pushed_at (开发活跃度)
}

// LanguageMap 语言占比 map (e.g. {"Go": 80.5, "Python": 19.5})
type LanguageMap map[string]float64

// FetchRepository 获取仓库元信息
func (c *Client) FetchRepository(ctx context.Context, owner, repo string) (*RepoInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch repository %s/%s: %w", owner, repo, err)
	}

	// GitHub API 返回的 owner 是对象，需要特殊处理
	var raw struct {
		FullName    string   `json:"full_name"`
		Owner       struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Language    string   `json:"language"`
		Stars       int      `json:"stargazers_count"`
		Topics      []string `json:"topics"`
		UpdatedAt   string   `json:"updated_at"`
		PushedAt    string   `json:"pushed_at"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse repository response: %w", err)
	}

	return &RepoInfo{
		FullName:    raw.FullName,
		Owner:       raw.Owner.Login,
		Name:        raw.Name,
		Description: raw.Description,
		Language:    raw.Language,
		Stars:       raw.Stars,
		Topics:      raw.Topics,
		UpdatedAt:   raw.UpdatedAt,
		PushedAt:    raw.PushedAt,
	}, nil
}

// SearchRepositories 通过 GitHub Search API 搜索仓库
func (c *Client) SearchRepositories(ctx context.Context, query string, perPage int) ([]*RepoInfo, error) {
	url := fmt.Sprintf("%s/search/repositories?q=%s&per_page=%d&sort=stars", c.baseURL, query, perPage)

	// Search API 限流检查
	if err := c.checkSearchRateLimit(); err != nil {
		c.waitForSearchRateLimitReset()
	}

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("search repositories: %w", err)
	}

	var result struct {
		Items []struct {
			FullName    string   `json:"full_name"`
			Owner       struct {
				Login string `json:"login"`
			} `json:"owner"`
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Language    string   `json:"language"`
			Stars       int      `json:"stargazers_count"`
			Forks       int      `json:"forks_count"`
			Topics      []string `json:"topics"`
			UpdatedAt   string   `json:"updated_at"`
			PushedAt    string   `json:"pushed_at"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	var repos []*RepoInfo
	for _, item := range result.Items {
		repos = append(repos, &RepoInfo{
			FullName:    item.FullName,
			Owner:       item.Owner.Login,
			Name:        item.Name,
			Description: item.Description,
			Language:    item.Language,
			Stars:       item.Stars,
			Topics:      item.Topics,
			UpdatedAt:   item.UpdatedAt,
			PushedAt:    item.PushedAt,
		})
	}

	return repos, nil
}

// FetchReadme 获取仓库 README 内容（解码 Base64）
func (c *Client) FetchReadme(ctx context.Context, owner, repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/readme", c.baseURL, owner, repo)

	// 请求 Accept: application/vnd.github.raw+json 直接获取原始内容
	body, err := c.doRequestWithAccept(ctx, url, "application/vnd.github.raw+json")
	if err != nil {
		// README 不存在不算错误
		if isNotFoundError(err) {
			log.Printf("[github] readme not found for %s/%s", owner, repo)
			return "", nil
		}
		return "", fmt.Errorf("fetch readme %s/%s: %w", owner, repo, err)
	}

	return string(body), nil
}

// FetchTopics 获取仓库 topics（已在 FetchRepository 中包含，此方法用于单独获取）
func (c *Client) FetchTopics(ctx context.Context, owner, repo string) ([]string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/topics", c.baseURL, owner, repo)

	body, err := c.doRequestWithAccept(ctx, url, "application/vnd.github.mercy-preview+json")
	if err != nil {
		return nil, fmt.Errorf("fetch topics %s/%s: %w", owner, repo, err)
	}

	var result struct {
		Names []string `json:"names"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse topics response: %w", err)
	}

	return result.Names, nil
}

// FetchLanguages 获取仓库语言占比
func (c *Client) FetchLanguages(ctx context.Context, owner, repo string) (LanguageMap, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/languages", c.baseURL, owner, repo)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch languages %s/%s: %w", owner, repo, err)
	}

	// GitHub 返回 {"Go": 12345, "Python": 6789}
	var rawBytes map[string]int64
	if err := json.Unmarshal(body, &rawBytes); err != nil {
		return nil, fmt.Errorf("parse languages response: %w", err)
	}

	// 计算百分比
	var total int64
	for _, v := range rawBytes {
		total += v
	}

	result := make(LanguageMap, len(rawBytes))
	if total == 0 {
		return result, nil
	}

	for lang, bytes := range rawBytes {
		result[lang] = float64(bytes) / float64(total) * 100
	}

	return result, nil
}

// GetPrimaryLanguage 从语言占比中获取主要语言
func GetPrimaryLanguage(lm LanguageMap) string {
	var maxLang string
	var maxPct float64
	for lang, pct := range lm {
		if pct > maxPct {
			maxLang = lang
			maxPct = pct
		}
	}
	return maxLang
}

// doRequest 执行带限流保护和重试的 HTTP 请求
func (c *Client) doRequest(ctx context.Context, url string) ([]byte, error) {
	return c.doRequestWithAccept(ctx, url, "application/vnd.github+json")
}

// doRequestWithAccept 执行带指定 Accept header 的请求
func (c *Client) doRequestWithAccept(ctx context.Context, url, accept string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// 限流检查：等待而非直接放弃
		if err := c.checkRateLimit(); err != nil {
			if rle, ok := err.(*RateLimitError); ok {
				log.Printf("[github] pre-request rate limit, waiting until %v", rle.ResetAt)
				c.WaitForRateLimitReset()
				// 不消耗重试次数，限流等待后继续
				attempt--
				continue
			}
			lastErr = err
			break
		}

		// 构建请求
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Accept", accept)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			if shouldRetry(attempt, c.maxRetries) {
				c.backoff(attempt)
				continue
			}
			break
		}

		// 更新限流信息
		c.updateRateLimit(resp)

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			if shouldRetry(attempt, c.maxRetries) {
				c.backoff(attempt)
				continue
			}
			break
		}

		// 处理状态码
		switch {
		case resp.StatusCode == http.StatusOK:
			return body, nil
		case resp.StatusCode == http.StatusNotFound:
			return nil, &NotFoundError{URL: url}
		case resp.StatusCode == http.StatusForbidden:
			// 可能是限流
			if c.coreLimitRemaining <= 0 {
				lastErr = &RateLimitError{ResetAt: c.coreLimitReset}
				if shouldRetry(attempt, c.maxRetries) {
					c.WaitForRateLimitReset()
					continue
				}
				break
			}
			lastErr = fmt.Errorf("forbidden: %s", string(body))
			if shouldRetry(attempt, c.maxRetries) {
				c.backoff(attempt)
				continue
			}
		case resp.StatusCode == http.StatusTooManyRequests:
			lastErr = &RateLimitError{ResetAt: c.coreLimitReset}
			if shouldRetry(attempt, c.maxRetries) {
				c.WaitForRateLimitReset()
				continue
			}
		default:
			lastErr = fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
			if shouldRetry(attempt, c.maxRetries) {
				c.backoff(attempt)
				continue
			}
		}

		break
	}

	return nil, lastErr
}

// updateRateLimit 从响应头更新限流信息（自动区分 Core/Search）
func (c *Client) updateRateLimit(resp *http.Response) {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")

	// 通过请求路径判断是 Core 还是 Search
	// Search API: /search/...  Core API: /repos/...
	isSearch := false
	if reqURL := resp.Request.URL.Path; len(reqURL) > 0 {
		isSearch = strings.Contains(reqURL, "/search/")
	}

	if remaining != "" {
		var val int
		fmt.Sscanf(remaining, "%d", &val)
		if isSearch {
			c.searchLimitRemaining = val
		} else {
			c.coreLimitRemaining = val
		}
	}
	if reset != "" {
		var unix int64
		fmt.Sscanf(reset, "%d", &unix)
		t := time.Unix(unix, 0)
		if isSearch {
			c.searchLimitReset = t
		} else {
			c.coreLimitReset = t
		}
	}
}

// checkRateLimit 检查 Core API 限流
func (c *Client) checkRateLimit() error {
	if c.coreLimitRemaining <= 0 && !c.coreLimitReset.IsZero() {
		if time.Now().Before(c.coreLimitReset) {
			return &RateLimitError{ResetAt: c.coreLimitReset}
		}
	}
	return nil
}

// checkSearchRateLimit 检查 Search API 限流
func (c *Client) checkSearchRateLimit() error {
	if c.searchLimitRemaining <= 0 && !c.searchLimitReset.IsZero() {
		if time.Now().Before(c.searchLimitReset) {
			return &RateLimitError{ResetAt: c.searchLimitReset}
		}
	}
	return nil
}

// WaitForRateLimitReset 等待 Core API 限流重置（公开方法）
func (c *Client) WaitForRateLimitReset() {
	c.waitForReset(&c.coreLimitReset)
}

// waitForSearchRateLimitReset 等待 Search API 限流重置
func (c *Client) waitForSearchRateLimitReset() {
	c.waitForReset(&c.searchLimitReset)
}

func (c *Client) waitForReset(resetAt *time.Time) {
	if resetAt.IsZero() {
		time.Sleep(60 * time.Second)
		return
	}
	waitDur := time.Until(*resetAt) + time.Second
	if waitDur > 0 {
		log.Printf("[github] rate limited, waiting %v until reset", waitDur)
		time.Sleep(waitDur)
	}
}

// GetRateLimitInfo 获取当前限流信息（供 status API 使用）
func (c *Client) GetRateLimitInfo() (coreRemaining, searchRemaining int, coreReset, searchReset time.Time) {
	return c.coreLimitRemaining, c.searchLimitRemaining, c.coreLimitReset, c.searchLimitReset
}

// backoff 指数退避
func (c *Client) backoff(attempt int) {
	delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	// 最大 30 秒
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	log.Printf("[github] retry attempt %d, backoff %v", attempt+1, delay)
	time.Sleep(delay)
}

func shouldRetry(attempt, maxRetries int) bool {
	return attempt < maxRetries
}

// --- 错误类型 ---

// RateLimitError GitHub API 限流错误
type RateLimitError struct {
	ResetAt time.Time
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("github api rate limited, resets at %s", e.ResetAt.Format(time.RFC3339))
}

// NotFoundError 资源不存在
type NotFoundError struct {
	URL string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.URL)
}

func isNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// IsRateLimitError 判断是否为限流错误
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// TruncateStrategy 截断策略接口
// MVP 实现按 bytes 截断，未来可扩展按 token 截断
type TruncateStrategy interface {
	Truncate(s string, maxBytes int) string
	Name() string
}

// ByteTruncator 按 bytes 截断（MVP 默认实现）
type ByteTruncator struct{}

func (t *ByteTruncator) Name() string { return "byte" }

func (t *ByteTruncator) Truncate(s string, maxBytes int) string {
	if len([]byte(s)) <= maxBytes {
		return s
	}
	// 按 bytes 截断，保证不截断在多字节字符中间
	b := []byte(s)
	if maxBytes < len(b) {
		// 确保不截断在 UTF-8 多字节字符中间
		for maxBytes > 0 && !utf8.RuneStart(b[maxBytes]) {
			maxBytes--
		}
		b = b[:maxBytes]
	}
	result := string(b)
	// 找到最后一个完整词
	if idx := strings.LastIndex(result, " "); idx > len(result)/2 {
		result = result[:idx]
	}
	return strings.TrimSpace(result)
}

// defaultTruncator 全局默认截断器
var defaultTruncator TruncateStrategy = &ByteTruncator{}

// CleanReadme 清洗 README 内容
// 去除 Markdown 噪声，按 bytes 截断到最大长度
func CleanReadme(raw string, maxBytes int) string {
	return CleanReadmeWithStrategy(raw, maxBytes, defaultTruncator)
}

// CleanReadmeWithStrategy 使用指定截断策略清洗 README
func CleanReadmeWithStrategy(raw string, maxBytes int, strategy TruncateStrategy) string {
	if raw == "" {
		return ""
	}

	cleaned := raw

	// 去除代码块
	cleaned = stripCodeBlocks(cleaned)

	// 去除图片链接 ![alt](url)
	cleaned = stripImageLinks(cleaned)

	// 去除链接，保留文字 [text](url) → text
	cleaned = stripLinks(cleaned)

	// 去除 HTML 标签
	cleaned = stripHTMLTags(cleaned)

	// 去除 Markdown 标题标记 # → 空格
	cleaned = stripHeadingMarkers(cleaned)

	// 压缩连续空白
	cleaned = compressWhitespace(cleaned)

	// 截断
	cleaned = strategy.Truncate(cleaned, maxBytes)

	return cleaned
}

// BuildEmbeddingText 构建 embedding 输入文本
// 优先级：README > Description + Topics fallback
func BuildEmbeddingText(readme, description string, topics []string) string {
	if readme != "" {
		text := readme
		if description != "" {
			text = description + "\n\n" + text
		}
		return text
	}

	// fallback: description + topics
	parts := []string{}
	if description != "" {
		parts = append(parts, description)
	}
	if len(topics) > 0 {
		parts = append(parts, "Topics: "+strings.Join(topics, ", "))
	}
	return strings.Join(parts, "\n\n")
}

// --- README 清洗辅助函数 ---

func stripCodeBlocks(s string) string {
	// 去除 ```...``` 代码块，保留代码块内文字（去掉标记）
	result := strings.Builder{}
	inBlock := false
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inBlock = !inBlock
			continue
		}
		if !inBlock {
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return result.String()
}

func stripImageLinks(s string) string {
	// ![alt](url) → 空
	for {
		start := strings.Index(s, "![")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "](")
		if end == -1 {
			break
		}
		closeParen := strings.Index(s[start+end+2:], ")")
		if closeParen == -1 {
			break
		}
		s = s[:start] + s[start+end+2+closeParen+1:]
	}
	return s
}

func stripLinks(s string) string {
	// [text](url) → text
	for {
		start := strings.Index(s, "[")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "](")
		if end == -1 {
			break
		}
		closeParen := strings.Index(s[start+end+2:], ")")
		if closeParen == -1 {
			break
		}
		text := s[start+1 : start+end]
		s = s[:start] + text + s[start+end+2+closeParen+1:]
	}
	return s
}

func stripHTMLTags(s string) string {
	for {
		start := strings.Index(s, "<")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+1:]
	}
	return s
}

func stripHeadingMarkers(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, "#")
		if trimmed != line && len(trimmed) > 0 && trimmed[0] == ' ' {
			lines[i] = trimmed[1:]
		} else if trimmed != line {
			lines[i] = trimmed
		}
	}
	return strings.Join(lines, "\n")
}

func compressWhitespace(s string) string {
	// 压缩连续空行
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	// 压缩行内连续空格
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		for strings.Contains(line, "  ") {
			line = strings.ReplaceAll(line, "  ", " ")
		}
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}
