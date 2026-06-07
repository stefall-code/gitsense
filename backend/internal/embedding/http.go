package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HTTPProvider 调用远程 Embedding Service（Python FastAPI sidecar）
type HTTPProvider struct {
	baseURL    string
	dimensions int
	httpClient *http.Client
}

// NewHTTPProvider 创建 HTTP Provider
func NewHTTPProvider(baseURL string, dimensions int) *HTTPProvider {
	if dimensions <= 0 {
		dimensions = 384 // MiniLM 默认维度
	}
	return &HTTPProvider{
		baseURL:    baseURL,
		dimensions: dimensions,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *HTTPProvider) Dimensions() int { return p.dimensions }

func (p *HTTPProvider) Generate(ctx context.Context, text string) ([]float32, error) {
	results, err := p.GenerateBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("embedding service: no embedding returned")
	}
	return results[0], nil
}

func (p *HTTPProvider) GenerateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := map[string]interface{}{
		"texts": texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/embed/batch", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request to embedding service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
		Dimensions int         `json:"dimensions"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("embedding service: expected %d embeddings, got %d", len(texts), len(result.Embeddings))
	}

	log.Printf("[http-provider] generated %d embeddings from %s (dims=%d)", len(result.Embeddings), p.baseURL, result.Dimensions)
	return result.Embeddings, nil
}

// HealthCheck 检查 Embedding Service 是否可用
func (p *HTTPProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("embedding service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("embedding service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
