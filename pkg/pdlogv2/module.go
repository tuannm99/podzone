package pdlogv2

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Init using Fx
func Module(loader Loader, opts ...Option) fx.Option {
	return fx.Options(
		fx.Provide(func(v *viper.Viper) Config {
			return loader(v)
		}),

		fx.Provide(func() *factoryOptions {
			return NewFactory(opts...)
		}),

		fx.Provide(func(f *factoryOptions, cfg Config) (Logger, error) {
			return f.ByProvider(context.Background(), cfg)
		}),

		// Sync logger when application stop
		fx.Invoke(func(lc fx.Lifecycle, log Logger) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					if err := log.Sync(); err != nil && !strings.Contains(err.Error(), "sync /dev/stderr") {
						log.Warn("logger sync failed", "error", err)
					}
					return nil
				},
			})
		}),
	)
}
