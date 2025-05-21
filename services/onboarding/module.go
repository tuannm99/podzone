package onboarding

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/tuannm99/podzone/pkg/httpfx"
	"github.com/tuannm99/podzone/services/onboarding/store"
)

var Module = fx.Options(
	fx.Provide(
		store.NewStoreService,
		store.NewStoreController,
	),

	fx.Provide(
		fx.Annotate(RegisterHTTPRoutes, fx.ResultTags(`group:"gin-routes"`)),
	),
)

type RegisterRoutesParams struct {
	fx.In

	Logger    *zap.Logger
	StoreCtrl *store.StoreController
}

func RegisterHTTPRoutes(p RegisterRoutesParams) httpfx.RouteRegistrar {
	p.Logger.Info("Registering Onboarding HTTP handler")
	return func(r *gin.Engine) {
		v1 := r.Group("/onboarding/v1")
		p.StoreCtrl.RegisterRoutes(v1)
	}
}
