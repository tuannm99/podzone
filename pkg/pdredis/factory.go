package pdredis

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func NewClientFromConfig(cfg *Config) (*redis.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil redis config")
	}

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
