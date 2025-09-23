package pdlogv2

import (
	"context"
	"strings"

	"go.uber.org/fx"
)

// Module wires: Loader -> Config -> Factory -> Logger, then lifecycle Sync on stop.
func Module(loader Loader, opts ...Option) fx.Option {
	return fx.Options(
		fx.Provide(func() Config {
			cfg := loader()
			// final safety defaults if loader omitted some fields
			if cfg.Provider == "" {
				cfg.Provider = "zap"
			}
			if cfg.Level == "" {
				cfg.Level = "info"
			}
			if cfg.Env == "" {
				cfg.Env = "prod"
			}
			return cfg
		}),
		fx.Provide(func() *factoryOptions { return NewFactory(opts...) }),
		fx.Provide(func(f *factoryOptions, cfg Config) (Logger, error) {
			return f.Make(context.Background(), cfg)
		}),
		fx.Invoke(registerLifecycle),
	)
}

func registerLifecycle(lc fx.Lifecycle, log Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := log.Sync(); err != nil {
				// Ignore benign stderr sync errors seen in CI on some platforms
				if strings.Contains(err.Error(), "sync /dev/stderr") {
					return nil
				}
				log.Warn("logger sync failed", "error", err)
			}
			return nil
		},
	})
}
