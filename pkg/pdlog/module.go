package pdlog

import (
	"context"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
)

func ModuleFor(appName string) fx.Option {
	return fx.Options(
		fx.Provide(func() (Logger, error) {
			return NewFrom(
				toolkit.GetEnv("LOG_PROVIDER", "zap"),
				context.Background(),
				WithLevel(toolkit.GetEnv("DEFAULT_LOG_LEVEL", "debug")),
				WithEnv(toolkit.GetEnv("APP_ENV", "dev")),
				WithAppName(appName),
			)
		}),
		fx.Invoke(registerLifecycle),
	)
}

// registerLifecycle hook for running logger.Sync() when app shutdown
func registerLifecycle(lc fx.Lifecycle, log Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return log.Sync()
		},
	})
}
