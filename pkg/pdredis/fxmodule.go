package pdredis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type RedisClient interface {
	Ping(ctx context.Context) error
	Close() error
}

type redisClientAdapter struct {
	c *redis.Client
}

func (a redisClientAdapter) Ping(ctx context.Context) error {
	return a.c.Ping(ctx).Err()
}

func (a redisClientAdapter) Close() error {
	return a.c.Close()
}

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}

	nameParamTag := `name:"pdredis-` + name + `"`
	configResultTag := `name:"redis-` + name + `-config"`
	clientResultTag := `name:"redis-` + name + `"`

	return fx.Options(
		fx.Supply(fx.Annotate(name, fx.ResultTags(nameParamTag))),

		fx.Provide(
			fx.Annotate(
				GetConfigFromKoanf,
				fx.ParamTags(nameParamTag, ``),
				fx.ResultTags(configResultTag),
			),
			fx.Annotate(
				NewClientFromConfig,
				fx.ParamTags(configResultTag),
				fx.ResultTags(clientResultTag),
			),

			// Provide redis.Cmdable (for app usage)
			fx.Annotate(
				func(c *redis.Client) redis.Cmdable { return c },
				fx.ParamTags(clientResultTag),
				fx.ResultTags(clientResultTag),
			),

			// Provide RedisClient adapter (for lifecycle mocking)
			fx.Annotate(
				func(c *redis.Client) RedisClient { return redisClientAdapter{c: c} },
				fx.ParamTags(clientResultTag),
				fx.ResultTags(clientResultTag),
			),
		),

		fx.Invoke(
			fx.Annotate(registerLifecycle, fx.ParamTags(``, clientResultTag, ``, configResultTag)),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, client RedisClient, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Pinging Redis...", "uri", cfg.URI)

			_ = time.Second

			return client.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Redis client")
			return client.Close()
		},
	})
}
