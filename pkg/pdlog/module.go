package pdlog

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func ModuleFor(appName string) fx.Option {
	return fx.Options(
		fx.Provide(func(v *viper.Viper) (Logger, error) {
			var cfg Config
			if sub := v.Sub("logger"); sub != nil {
				_ = sub.Unmarshal(&cfg) // keep defaults if missing
			}
			if cfg.Provider == "" {
				cfg.Provider = "zap"
			}
			if cfg.Level == "" {
				cfg.Level = "debug"
			}
			if cfg.Env == "" {
				cfg.Env = "dev"
			}
			cfg.AppName = appName

			f := Registry.Get()
			if ff, ok := Registry.Lookup(cfg.Provider); ok {
				f = ff
			}
			return f(context.Background(), cfg)
		}),
		fx.Invoke(registerLifecycle),
	)
}

// Flush logger on shutdown
func registerLifecycle(lc fx.Lifecycle, log Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := log.Sync(); err != nil {
				log.Warn("logger sync failed").Err(err).Send()
			}
			return nil
		},
	})
}
