package onboarding

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/eventstore"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/publisher"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/worker"
	"github.com/tuannm99/podzone/internal/onboarding/store"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdworker"
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

	fx.Invoke(func(lc fx.Lifecycle, store core.ConnectionStore, log pdlog.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Info("Ensuring onboarding mongo indexes...")
				return store.EnsureIndexes(ctx)
			},
		})
	}),

	fx.Invoke(func(lc fx.Lifecycle, log pdlog.Logger, w *worker.OutboxWorker) {
		pdworker.StartWorker(lc, log, w)
	}),
)

var (
	InfrasCtrlProvider = fx.Provide(
		// Wire Infras Core
		func() map[core.InfraType]core.InfraProvisioner {
			return map[core.InfraType]core.InfraProvisioner{}
		},
		publisher.NewConsulPublisher,
		fx.Annotate(
			func() string {
				return "onboarding"
			},
			fx.ResultTags(`name:"mongo-onboarding-db"`),
		),
		fx.Annotate(eventstore.NewMongoStore, fx.As(new(core.ConnectionStore))),
		core.NewInfraManager,
		worker.NewOutboxWorker,

		// Infras API
		infrasmanager.NewService,
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
