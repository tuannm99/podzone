package pdredis

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-redis/redis/v8"
)

func RedisProvider(ctx context.Context, cfg Config) (*redis.Client, error) {
	redisURL, err := url.Parse(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("invalid redis uri %q: %w", cfg.URI, err)
	}

	pass, _ := redisURL.User.Password()
	db := 0
	if redisURL.Path != "" {
		if parsed, err := strconv.Atoi(redisURL.Path[1:]); err == nil {
			db = parsed
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL.Host,
		Password: pass,
		DB:       db,
	})
	return client, nil
}

func MockProvider(context.Context, Config) (*redis.Client, error) {
	return redis.NewClient(&redis.Options{Addr: "localhost:0"}), nil
}
