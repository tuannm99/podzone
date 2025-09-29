package pdworker

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type Worker interface {
	Run(context.Context) error
}

func StartWorker(lc fx.Lifecycle, logger pdlog.Logger, w Worker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := w.Run(ctx); err != nil {
					logger.Error("Worker exited", "error", err)
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
