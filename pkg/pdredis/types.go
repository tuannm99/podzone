package pdredis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// Config đọc từ viper: redis.<name>.*
type Config struct {
	Provider string `mapstructure:"provider"` // "real" | "mock"
	URI      string `mapstructure:"uri"`
}

// ProviderFn tạo *redis.Client theo Config
type ProviderFn func(ctx context.Context, cfg Config) (*redis.Client, error)
