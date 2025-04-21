package kafkafx

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Invoke(RegisterLifecycle),
)

func NewClient() {}

type Queue interface{}

func RegisterLifecycle(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Redis connection")
			return nil
		},
	})
}
