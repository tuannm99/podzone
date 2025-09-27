package pdconfig

import "go.uber.org/fx"

var Module = fx.Provide(NewAppConfig)
