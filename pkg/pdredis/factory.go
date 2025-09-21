package pdredis

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var RedisFactory = func(ctx context.Context, cfg Config) (*redis.Client, error) {
	redisURL, err := url.Parse(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("invalid redis uri %s: %w", cfg.URI, err)
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

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", redisURL.Host, err)
	}
	return client, nil
}

var NoopRedisFactory = func(ctx context.Context, _ Config) (*redis.Client, error) {
	return redis.NewClient(&redis.Options{Addr: "localhost:0"}), nil
}
