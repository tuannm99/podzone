package pdmongo

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/fx"
)

type moduleOptions struct {
	name      string
	providers map[string]ProviderFn
	fallback  ProviderFn
}

type Option func(*moduleOptions)

func WithProvider(id string, f ProviderFn) Option {
	return func(o *moduleOptions) { o.providers[id] = f }
}

func WithFallback(f ProviderFn) Option { return func(o *moduleOptions) { o.fallback = f } }

// tag: name:"mongo-<name>"
func WithName(name string) Option { return func(o *moduleOptions) { o.name = name } }

// ViperLoaderFor  mongo.<name>
func ViperLoaderFor(name string) func(*viper.Viper) Config {
	return func(v *viper.Viper) Config {
		base := "mongo." + name
		c := Config{
			Provider: v.GetString(base + ".provider"),
			URI:      v.GetString(base + ".uri"),
		}
		if c.Provider == "" {
			c.Provider = "real"
		}
		return c
	}
}

// Module: Viper -> Config -> Provider -> *mongo.Client + lifecycle
func Module(loader func(*viper.Viper) Config, opts ...Option) fx.Option {
	base := &moduleOptions{
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

	resultTag := fmt.Sprintf(`name:"%s"`, "mongo-"+base.name)

	return fx.Options(
		fx.Provide(
			fx.Annotate(func(v *viper.Viper, lc fx.Lifecycle, log pdlog.Logger) (*mongo.Client, error) {
				cfg := loader(v)

				prov := base.fallback
				if p, ok := base.providers[cfg.Provider]; ok {
					prov = p
				}

				cl, err := prov(context.Background(), cfg)
				if err != nil {
					return nil, err
				}

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						if cfg.Provider == "mock" {
							log.Info("Using Mongo mock provider", "name", base.name, "uri", cfg.URI)
							return nil
						}
						log.Info("Pinging Mongo...", "name", base.name, "dsn", cfg.URI)
						tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
						defer cancel()
						if err := cl.Ping(tctx, readpref.Primary()); err != nil {
							return fmt.Errorf("mongo ping failed: %w", err)
						}
						log.Info("Mongo is reachable", "name", base.name)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						if cfg.Provider == "mock" {
							log.Info("Skipping Mongo Disconnect for mock", "name", base.name)
							return nil
						}
						log.Info("Closing Mongo client", "name", base.name)
						tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
						defer cancel()
						return cl.Disconnect(tctx)
					},
				})

				return cl, nil
			}, fx.ResultTags(resultTag)),
		),
	)
}
