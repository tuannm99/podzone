package workerfx

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Worker interface {
	Run(context.Context) error
}

func StartWorker(lc fx.Lifecycle, logger *zap.Logger, w Worker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := w.Run(ctx); err != nil {
					logger.Error("Worker exited", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Worker stopping...")
			return nil
		},
	})
}
