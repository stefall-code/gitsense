package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gitsense/gitsense/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

// CacheStatus 缓存命中状态，供 handler 设置 X-Cache Header
type CacheStatus string

const (
	CacheHit    CacheStatus = "HIT"
	CacheMiss   CacheStatus = "MISS"
	CacheBypass CacheStatus = "BYPASS"
)

// Stats 缓存统计数据
type Stats struct {
	Hit    int64 `json:"hit"`
	Miss   int64 `json:"miss"`
	Set    int64 `json:"set"`
	Delete int64 `json:"delete"`
}

// StatsResponse 缓存统计响应
type StatsResponse struct {
	RedisConnected bool    `json:"redis_connected"`
	Hit            int64   `json:"hit,omitempty"`
	Miss           int64   `json:"miss,omitempty"`
	Set            int64   `json:"set,omitempty"`
	Delete         int64   `json:"delete,omitempty"`
	HitRate        float64 `json:"hit_rate,omitempty"`
}

// Client Redis 缓存客户端，自动降级
type Client struct {
	rdb   *redis.Client
	avail bool
	stats struct {
		hit    atomic.Int64
		miss   atomic.Int64
		set    atomic.Int64
		delete atomic.Int64
	}
}

// NewClient 创建 Redis 客户端，连接失败自动降级
func NewClient(cfg config.RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis unavailable, fallback enabled")
		return &Client{rdb: rdb, avail: false}
	}

	log.Println("Redis connected")
	return &Client{rdb: rdb, avail: true}
}

// IsAvailable Redis 是否可用
func (c *Client) IsAvailable() bool {
	return c.avail
}

// Get 获取缓存值，返回 (value, status, error)
func (c *Client) Get(ctx context.Context, key string) (string, CacheStatus, error) {
	if !c.avail {
		return "", CacheBypass, redis.Nil
	}
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		c.stats.miss.Add(1)
		log.Printf("[cache] MISS %s", shortKey(key))
		return "", CacheMiss, redis.Nil
	}
	if err != nil {
		log.Printf("[cache] error get %s: %v", key, err)
		return "", CacheMiss, err
	}
	c.stats.hit.Add(1)
	log.Printf("[cache] HIT %s", shortKey(key))
	return val, CacheHit, nil
}

// Set 设置缓存值
func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if !c.avail {
		return nil
	}
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		log.Printf("[cache] error set %s: %v", key, err)
		return err
	}
	c.stats.set.Add(1)
	log.Printf("[cache] SET %s", shortKey(key))
	return nil
}

// Del 删除缓存
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if !c.avail {
		return nil
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		log.Printf("[cache] error del %v: %v", keys, err)
		return err
	}
	for _, k := range keys {
		c.stats.delete.Add(1)
		log.Printf("[cache] DEL %s", shortKey(k))
	}
	return nil
}

// GetJSON 获取缓存并反序列化为 JSON，返回缓存状态
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) (CacheStatus, error) {
	val, status, err := c.Get(ctx, key)
	if err != nil {
		return status, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return CacheMiss, err
	}
	return status, nil
}

// SetJSON 序列化为 JSON 并设置缓存
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}
	return c.Set(ctx, key, string(data), ttl)
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

// GetStats 获取缓存统计
func (c *Client) GetStats() StatsResponse {
	if !c.avail {
		return StatsResponse{RedisConnected: false}
	}
	hit := c.stats.hit.Load()
	miss := c.stats.miss.Load()
	total := hit + miss
	var hitRate float64
	if total > 0 {
		hitRate = float64(hit) / float64(total)
	}
	return StatsResponse{
		RedisConnected: true,
		Hit:            hit,
		Miss:           miss,
		Set:            c.stats.set.Load(),
		Delete:         c.stats.delete.Load(),
		HitRate:        hitRate,
	}
}

// shortKey 将缓存 key 缩短显示：discovery:owner/repo → discovery owner/repo
func shortKey(key string) string {
	if idx := strings.Index(key, ":"); idx >= 0 {
		return key[:idx] + " " + key[idx+1:]
	}
	return key
}

// --- 缓存 Key 常量 ---

const (
	// CacheVersion 缓存版本号，改 schema 时递增即可自动失效所有旧缓存
	CacheVersion  = "v2"
	KeyDiscovery  = "discovery"   // discovery:{owner}/{repo}
	KeyEcosystems = "ecosystems"  // ecosystems:list
	KeyEcosystem  = "ecosystem"   // ecosystem:{name}
	KeyTrending   = "trending"    // trending:{name}
)

// DiscoveryKey 生成 Discovery 缓存 Key
func DiscoveryKey(owner, repo string) string {
	return fmt.Sprintf("%s:%s:%s/%s", KeyDiscovery, CacheVersion, owner, repo)
}

// EcosystemKey 生成 Ecosystem 详情缓存 Key
func EcosystemKey(name string) string {
	return fmt.Sprintf("%s:%s:%s", KeyEcosystem, CacheVersion, name)
}

// TrendingKey 生成 Trending 缓存 Key
func TrendingKey(name string) string {
	return fmt.Sprintf("%s:%s:%s", KeyTrending, CacheVersion, name)
}

// EcosystemsListKey 生成 Ecosystem 列表缓存 Key
func EcosystemsListKey() string {
	return fmt.Sprintf("%s:%s:list", KeyEcosystems, CacheVersion)
}

// --- TTL 常量 ---

const (
	DiscoveryTTL = 30 * time.Minute
	EcosystemTTL = 1 * time.Hour
	TrendingTTL  = 30 * time.Minute
)
