package pdtenantdb

import "go.uber.org/fx"

var Module = fx.Module("pdtenantdb", fx.Provide(NewManager))
