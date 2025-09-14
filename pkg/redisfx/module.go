package redisfx

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func ModuleFor(name string, conStr string) fx.Option {
	uriName := fmt.Sprintf("%s-redis-uri", name)
	resultName := fmt.Sprintf("redis-%s", name)

	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return conStr },
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, uriName)),
			),
		),

		fx.Provide(
			fx.Annotate(
				func(logger pdlog.Logger, uri string) (*redis.Client, error) {
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

					logger.Info("Initializing Redis connection").With("addr", redisUrl.Host).Send()

					client := redis.NewClient(&redis.Options{
						Addr:     redisUrl.Host,
						Password: pass,
						DB:       db,
					})

					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					err = client.Ping(ctx).Err()
					cancel()

					if err != nil {
						return nil, fmt.Errorf("failed to connect to Redis at %s: %w", redisUrl.Host, err)
					}

					logger.Info("Successfully connected to Redis").With("addr", redisUrl.Host).Send()
					return client, nil
				},
				fx.ParamTags(``, fmt.Sprintf(`name:"%s"`, uriName)),
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, resultName)),
			),
		),

		fx.Invoke(
			fx.Annotate(
				func(lc fx.Lifecycle, logger pdlog.Logger, client *redis.Client) {
					lc.Append(fx.Hook{
						OnStop: func(ctx context.Context) error {
							logger.Info("Closing Redis client").With("name", name).Send()
							return client.Close()
						},
					})
				},
				fx.ParamTags(``, ``, fmt.Sprintf(`name:"%s"`, resultName)),
			),
		),
	)
}
