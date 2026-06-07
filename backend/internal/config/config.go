package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Redis     RedisConfig
	GitHub    GitHubConfig
	Embedding EmbeddingConfig
	Cache     CacheConfig
	Trend     TrendConfig
	Bootstrap BootstrapConfig
}

type ServerConfig struct {
	Port       string
	Mode       string // debug / release / test
	AdminToken string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type GitHubConfig struct {
	Token string
}

type EmbeddingConfig struct {
	Provider   string // openai / local / http / mock
	APIKey     string
	Model      string
	Dimensions int
	Host       string // embedding service host (for local/http provider)
	Port       string // embedding service port
	BatchSize  int    // worker batch size
}

type CacheConfig struct {
	RepoTTL           int // seconds
	RecommendationTTL int // seconds
	EcosystemTTL      int // seconds
}

type TrendConfig struct {
	EnableRanking bool // ENABLE_TREND_RANKING feature flag
	CacheTTL      int  // seconds (default 12h)
}

type BootstrapConfig struct {
	MinStars    int // MIN_REPO_STARS
	MinForks    int // MIN_REPO_FORKS
	ActiveYears int // REPO_ACTIVE_YEARS
	MaxDepth    int // BOOTSTRAP_MAX_DEPTH
	Workers     int // BOOTSTRAP_WORKERS
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:       getEnv("SERVER_PORT", "8080"),
			Mode:       getEnv("GIN_MODE", "debug"),
			AdminToken: getEnv("ADMIN_TOKEN", ""),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "gitsense"),
			Password: getEnv("DB_PASSWORD", "gitsense"),
			Name:     getEnv("DB_NAME", "gitsense"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		GitHub: GitHubConfig{
			Token: getEnv("GITHUB_TOKEN", ""),
		},
		Embedding: EmbeddingConfig{
			Provider:   getEnv("EMBEDDING_PROVIDER", "local"),
			APIKey:     getEnv("OPENAI_API_KEY", ""),
			Model:      getEnv("EMBEDDING_MODEL", "all-MiniLM-L6-v2"),
			Dimensions: getEnvInt("EMBEDDING_DIMENSIONS", 384),
			Host:       getEnv("EMBEDDING_SERVICE_HOST", "embedding-service"),
			Port:       getEnv("EMBEDDING_SERVICE_PORT", "8001"),
			BatchSize:  getEnvInt("EMBEDDING_BATCH_SIZE", 10),
		},
		Cache: CacheConfig{
			RepoTTL:           getEnvInt("CACHE_REPO_TTL", 86400),
			RecommendationTTL: getEnvInt("CACHE_RECOMMENDATION_TTL", 3600),
			EcosystemTTL:      getEnvInt("CACHE_ECOSYSTEM_TTL", 3600),
		},
		Trend: TrendConfig{
			EnableRanking: getEnvBool("ENABLE_TREND_RANKING", true),
			CacheTTL:      getEnvInt("TREND_CACHE_TTL", 43200), // 12h
		},
		Bootstrap: BootstrapConfig{
			MinStars:    getEnvInt("MIN_REPO_STARS", 100),
			MinForks:    getEnvInt("MIN_REPO_FORKS", 10),
			ActiveYears: getEnvInt("REPO_ACTIVE_YEARS", 2),
			MaxDepth:    getEnvInt("BOOTSTRAP_MAX_DEPTH", 2),
			Workers:     getEnvInt("BOOTSTRAP_WORKERS", 3),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
