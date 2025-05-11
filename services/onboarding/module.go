package onboarding

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/tuannm99/podzone/pkg/httpfx"
	"github.com/tuannm99/podzone/services/onboarding/configuration"
)

var Module = fx.Options(
	fx.Provide(
		configuration.NewOnboardingController,
	),
	fx.Provide(
		fx.Annotate(RegisterHTTPRoutes, fx.ResultTags(`group:"gin-routes"`)),
	),
)

type RegisterRoutesParams struct {
	fx.In

	Logger         *zap.Logger
	OnboardingCtrl *configuration.OnboardingController
}

func RegisterHTTPRoutes(p RegisterRoutesParams) httpfx.RouteRegistrar {
	p.Logger.Info("Registering Onboarding HTTP handler")
	return func(r *gin.Engine) {
		v1 := r.Group("/onboarding/v1")

		v1.GET("/configs", p.OnboardingCtrl.GetAll)

		v1.GET("/stores", p.OnboardingCtrl.GetAll)
	}
}
