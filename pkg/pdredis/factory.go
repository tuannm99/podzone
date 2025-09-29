package pdredis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func NewClientFromConfig(cfg *Config) (*redis.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil redis config")
	}
	switch cfg.Provider {
	case "mock":
		return RedisProvider(context.Background(), *cfg)
	case "", "real":
		return MockProvider(context.Background(), *cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}
