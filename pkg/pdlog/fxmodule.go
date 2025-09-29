package pdlog

import (
	"context"

	"go.uber.org/fx"
)

// Module wires logger configuration and sync lifecycle
var Module = fx.Options(
	fx.Provide(
		GetLogConfigFromViper,
		NewLogger,
	),
	fx.Invoke(registerLoggerLifecycle),
)

func registerLoggerLifecycle(lc fx.Lifecycle, log Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			Cleanup(log)
			return nil
		},
	})
}
