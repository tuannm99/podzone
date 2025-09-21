package pdredis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
)

type Config struct {
	URI string `mapstructure:"uri"`
}

var Registry = toolkit.NewRegistry[*redis.Client, Config]("real")

func init() {
	Registry.Register("real", RedisFactory)
	Registry.Register("noop", NoopRedisFactory)
}

func ModuleFor(name string) fx.Option {
	tag := fmt.Sprintf(`name:"%s"`, "redis-"+name)

	return fx.Provide(
		fx.Annotate(func(v *viper.Viper, lc fx.Lifecycle, logger pdlog.Logger) (*redis.Client, error) {
			// expect: redis.<name>.uri
			sub := v.Sub("redis")
			if sub == nil {
				return nil, fmt.Errorf("missing config block: redis")
			}
			sub = sub.Sub(name)
			if sub == nil {
				return nil, fmt.Errorf("missing config block: redis.%s", name)
			}

			var cfg Config
			if err := sub.Unmarshal(&cfg); err != nil {
				return nil, fmt.Errorf("unmarshal redis.%s failed: %w", name, err)
			}

			factory := Registry.Get()
			client, err := factory(context.Background(), cfg)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					logger.Info("Closing Redis client").With("name", name).Send()
					return client.Close()
				},
			})
			return client, nil
		}, fx.ResultTags(tag)),
	)
}
