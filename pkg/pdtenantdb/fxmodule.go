package pdtenantdb

import (
	"context"
	"time"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"pdtenantdb",
	fx.Provide(
		NewDefaultKVClusterRegistry,
		NewManager,
		NewKVPlacementResolver,
	),
	fx.Invoke(func(lc fx.Lifecycle, m Manager, cfg *Config) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				ttl := cfg.DedicatedIdleTTL
				if ttl == 0 {
					ttl = 30 * time.Minute
				}
				go func() {
					ticker := time.NewTicker(ttl / 2)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case now := <-ticker.C:
							m.CloseIdleDedicated(now)
						}
					}
				}()
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return m.CloseAll()
			},
		})
	}),
)
