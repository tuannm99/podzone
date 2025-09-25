package pdpostgres

import (
	"context"

	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func ViperLoaderFor(name string) func(*viper.Viper) Config {
	return func(v *viper.Viper) Config {
		base := "postgres." + name
		var cfg Config
		if sub := v.Sub(base); sub != nil {
			_ = sub.Unmarshal(&cfg)
		}
		return cfg
	}
}

type factoryOptions struct {
	fallback  ProviderFn
	providers map[string]ProviderFn
	name      string
}

type Option func(*factoryOptions)

func WithProvider(id string, p ProviderFn) Option {
	return func(o *factoryOptions) {
		if o.providers == nil {
			o.providers = map[string]ProviderFn{}
		}
		o.providers[id] = p
	}
}

func WithFallback(p ProviderFn) Option {
	return func(o *factoryOptions) { o.fallback = p }
}

func WithName(name string) Option {
	return func(o *factoryOptions) { o.name = name }
}

func Module(loader func(*viper.Viper) Config, opts ...Option) fx.Option {
	var fo factoryOptions
	for _, f := range opts {
		f(&fo)
	}
	if fo.fallback == nil {
		fo.fallback = RealProvider
	}
	name := fo.name
	if name == "" {
		name = "default"
	}
	resultTag := `name:"gorm-` + name + `"`

	return fx.Provide(
		fx.Annotate(func(v *viper.Viper, lc fx.Lifecycle, log pdlogv2.Logger) (*gorm.DB, error) {
			cfg := loader(v)
			provider := fo.fallback
			if p, ok := fo.providers[cfg.Provider]; ok {
				provider = p
			}
			db, err := provider(context.Background(), cfg)
			if err != nil {
				return nil, err
			}

			isMock := cfg.Provider == "mock"
			logCtx := log.With("name", name)

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if isMock {
						logCtx.Info("Postgres mock ready", "uri", cfg.URI)
						return nil
					}
					logCtx.Info("Pinging Postgres...", "uri", cfg.URI)
					sqlDB, _ := db.DB()
					return sqlDB.PingContext(ctx)
				},
				OnStop: func(ctx context.Context) error {
					if isMock {
						logCtx.Info("Skipping Postgres close for mock")
						return nil
					}
					logCtx.Info("Closing Postgres connection")
					sqlDB, _ := db.DB()
					return sqlDB.Close()
				},
			})
			return db, nil
		}, fx.ResultTags(resultTag)),
	)
}
