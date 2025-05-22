package services

import (
	"go.uber.org/fx"
)

// Module exports all service components
var Module = fx.Options(
	fx.Provide(
		NewStoreService,
	),
)
