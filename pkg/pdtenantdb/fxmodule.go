package pdtenantdb

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"pdtenantdb",
	fx.Provide(
		NewDefaultConsulClusterRegistry,
		NewManager,
		NewStaticPlacementResolver,
	),
	fx.Invoke(func(lc fx.Lifecycle, m Manager) {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error { return m.CloseAll() },
		})
	}),
)
