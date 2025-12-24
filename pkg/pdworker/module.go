package pdworker

import (
	"context"
	"sync"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type Worker interface {
	Run(context.Context)
}

func StartWorker(lc fx.Lifecycle, logger pdlog.Logger, w Worker) {
	var (
		runCtx context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup
	)

	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			// IMPORTANT: fx ctx has short started time
			runCtx, cancel = context.WithCancel(context.Background())

			wg.Go(func() {
				w.Run(runCtx)
			})

			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			logger.Info("Worker stopping...")

			if cancel != nil {
				cancel()
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-stopCtx.Done():
				logger.Warn("Worker stop timeout", "error", stopCtx.Err())
				return nil
			case <-done:
				return nil
			}
		},
	})
}
