package pdredis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type options struct {
	name      string
	providers map[string]ProviderFn
	fallback  ProviderFn
}

type Option func(*options)

func WithProvider(id string, f ProviderFn) Option { return func(o *options) { o.providers[id] = f } }

func WithFallback(f ProviderFn) Option { return func(o *options) { o.fallback = f } }

func WithName(name string) Option { return func(o *options) { o.name = name } }

// ViperLoaderFor read redis.<name>
func ViperLoaderFor(name string) func(*viper.Viper) Config {
	return func(v *viper.Viper) Config {
		base := "redis." + name
		var cfg Config
		if sub := v.Sub(base); sub != nil {
			_ = sub.Unmarshal(&cfg)
		}
		return cfg
	}
}

// Module: Viper -> Config -> chọn Provider -> *redis.Client + lifecycle
func Module(loader func(*viper.Viper) Config, opts ...Option) fx.Option {
	base := &options{
		providers: map[string]ProviderFn{
			"real": RealProvider,
			"mock": MockProvider,
		},
		fallback: RealProvider,
		name:     "default",
	}
	for _, opt := range opts {
		opt(base)
	}

	resultTag := fmt.Sprintf(`name:"%s"`, "redis-"+base.name)

	return fx.Options(
		fx.Provide(
			fx.Annotate(func(v *viper.Viper, lc fx.Lifecycle, log pdlog.Logger) (*redis.Client, error) {
				cfg := loader(v)

				prov := base.fallback
				if p, ok := base.providers[cfg.Provider]; ok {
					prov = p
				}

				client, err := prov(context.Background(), cfg)
				if err != nil {
					return nil, err
				}

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						if cfg.Provider == "mock" {
							log.Info("Redis mock ready", "name", base.name, "uri", cfg.URI)
							return nil
						}
						log.Info("Pinging Redis...", "name", base.name, "uri", cfg.URI)
						return client.Ping(ctx).Err()
					},
					OnStop: func(ctx context.Context) error {
						log.Info("Closing Redis client", "name", base.name)
						return client.Close()
					},
				})

				return client, nil
			}, fx.ResultTags(resultTag)),
		),
	)
}
