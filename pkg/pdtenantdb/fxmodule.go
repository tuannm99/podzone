package pdtenantdb

import (
	"context"
	"time"

	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
)

// newKVPlacementResolverFromConfig wires Config.PlacementCacheTTL into the
// KV-backed resolver instead of always using the hardcoded 2-minute
// default - see MT9, docs/07-problems/multitenancy-2026-07-23.md.
func newKVPlacementResolverFromConfig(kv kvstores.KVStore, cfg *Config) PlacementResolver {
	return NewKVPlacementResolverWithTTL(kv, "podzone/tenants", cfg.PlacementCacheTTL)
}

var Module = fx.Module(
	"pdtenantdb",
	fx.Provide(
		NewDefaultKVClusterRegistry,
		NewManager,
		newKVPlacementResolverFromConfig,
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
