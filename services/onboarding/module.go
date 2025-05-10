package onboarding

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(),
	fx.Invoke(
		RegisterHTTPRoutes,
	),
)

func RegisterHTTPRoutes(logger *zap.Logger) {
	logger.Info("Registering Onboarding HTTP handler")
}
