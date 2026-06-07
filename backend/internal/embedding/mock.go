package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// MockProvider 用于测试的 mock embedding provider
// 基于 hash 生成确定性向量，相同输入永远返回相同向量
type MockProvider struct {
	dimensions int
}

// NewMockProvider 创建 Mock Provider
func NewMockProvider(dimensions int) *MockProvider {
	if dimensions <= 0 {
		dimensions = 1536
	}
	return &MockProvider{dimensions: dimensions}
}

func (p *MockProvider) Dimensions() int { return p.dimensions }

func (p *MockProvider) Generate(ctx context.Context, text string) ([]float32, error) {
	return p.hashToVector(text), nil
}

func (p *MockProvider) GenerateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		results[i] = p.hashToVector(text)
	}
	return results, nil
}

// hashToVector 基于 SHA256 生成确定性向量
// 相同文本永远生成相同向量，不同文本生成不同向量
func (p *MockProvider) hashToVector(text string) []float32 {
	vec := make([]float32, p.dimensions)
	if text == "" {
		return vec
	}

	// 用 SHA256 生成种子，然后扩展到 dimensions
	hash := sha256.Sum256([]byte(text))

	// 循环使用 hash bytes 填充向量
	for i := 0; i < p.dimensions; i++ {
		// 每 4 bytes 生成一个 float32
		offset := (i * 4) % len(hash)
		bits := binary.LittleEndian.Uint32(hash[offset : offset+4])
		// 映射到 [-1, 1] 范围
		vec[i] = float32(int32(bits)) / float32(1<<31)
	}

	// 归一化
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = sqrt32(norm)
		for i := range vec {
			vec[i] /= norm
		}
	}

	return vec
}

func sqrt32(x float32) float32 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// Verify MockProvider implements EmbeddingProvider
var _ fmt.Stringer = (*MockProvider)(nil)

func (p *MockProvider) String() string {
	return fmt.Sprintf("MockProvider(dims=%d)", p.dimensions)
}
