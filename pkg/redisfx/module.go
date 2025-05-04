package redisfx

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ModuleFor(name string, conStr string) fx.Option {
	uriName := fmt.Sprintf("%s-redis-uri", name)
	resultName := fmt.Sprintf("redis-%s", name)

	return fx.Options(
		// Provide the Redis URI as named string
		fx.Provide(
			fx.Annotate(
				func() string { return conStr },
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, uriName)),
			),
		),

		// Provide the Redis client using URI
		fx.Provide(
			fx.Annotate(
				func(logger *zap.Logger, uri string) (*redis.Client, error) {
					redisUrl, err := url.Parse(uri)
					if err != nil {
						return nil, fmt.Errorf("invalid redis uri %s: %w", uri, err)
					}

					pass, _ := redisUrl.User.Password()
					db := 0
					if redisUrl.Path != "" {
						if parsed, err := strconv.Atoi(redisUrl.Path[1:]); err == nil {
							db = parsed
						}
					}

					logger.Info("Initializing Redis connection", zap.String("addr", redisUrl.Host))

					client := redis.NewClient(&redis.Options{
						Addr:     redisUrl.Host,
						Password: pass,
						DB:       db,
					})

					for attempt := 1; attempt <= 3; attempt++ {
						ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
						err = client.Ping(ctx).Err()
						cancel()

						if err == nil {
							logger.Info("Successfully connected to Redis", zap.String("addr", redisUrl.Host))
							break
						}

						if attempt < 3 {
							retryDelay := time.Duration(attempt) * time.Second
							logger.Warn("Redis connection failed, retrying...",
								zap.String("addr", redisUrl.Host),
								zap.Int("attempt", attempt),
								zap.Duration("retry_delay", retryDelay),
								zap.Error(err))
							time.Sleep(retryDelay)
						}
					}

					if err != nil {
						logger.Error("Failed to connect to Redis after multiple attempts",
							zap.String("addr", redisUrl.Host),
							zap.Int("attempts", 3),
							zap.Error(err))
						return nil, fmt.Errorf("failed to connect to Redis at %s: %w", redisUrl.Host, err)
					}

					return client, nil
				},
				fx.ParamTags(``, fmt.Sprintf(`name:"%s"`, uriName)),
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, resultName)),
			),
		),

		fx.Invoke(
			fx.Annotate(
				func(lc fx.Lifecycle, logger *zap.Logger, client *redis.Client) {
					lc.Append(fx.Hook{
						OnStop: func(ctx context.Context) error {
							logger.Info("Closing Redis client", zap.String("name", name))
							return client.Close()
						},
					})
				},
				fx.ParamTags(``, ``, fmt.Sprintf(`name:"%s"`, resultName)),
			),
		),
	)
}
