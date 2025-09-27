package onboarding

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager"
	"github.com/tuannm99/podzone/internal/onboarding/store"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
)

var Module = fx.Options(
	StoreCtrlProvider,
	InfrasCtrlProvider,

	fx.Provide(
		fx.Annotate(
			RegisterHTTPRoutes,
			fx.ResultTags(`group:"gin-routes"`),
		),
	),
)

var (
	InfrasCtrlProvider = fx.Provide(
		fx.Annotate(
			infrasmanager.NewController,
			fx.As(new(Controller)),
			fx.ResultTags(`group:"onboarding-controllers"`),
		),
	)

	StoreCtrlProvider = fx.Provide(
		store.NewStoreService,
		fx.Annotate(
			store.NewStoreController,
			fx.As(new(Controller)),
			fx.ResultTags(`group:"onboarding-controllers"`),
		),
	)
)

type Controller interface {
	RegisterRoutes(r *gin.RouterGroup)
}

type RegisterRoutesParams struct {
	fx.In

	Logger      pdlog.Logger
	Controllers []Controller `group:"onboarding-controllers"`
}

func RegisterHTTPRoutes(p RegisterRoutesParams) pdhttp.RouteRegistrar {
	p.Logger.Info("Registering Onboarding HTTP handler")
	return func(r *gin.Engine) {
		v1 := r.Group("/onboarding/v1")
		for _, ctrl := range p.Controllers {
			ctrl.RegisterRoutes(v1)
		}
	}
}
