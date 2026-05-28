package onboarding

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager"
	consulbridge "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/controller/eventhandler/consulbridge"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/eventstore"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/publisher"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/worker"
	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/inputport"
	"github.com/tuannm99/podzone/internal/onboarding/store"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/store/inputport"
	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdworker"
)

const (
	onboardingConsumerRuntimePath = "messaging.onboarding.consumers.consul_bridge"
	onboardingConsumerName        = "onboarding.consul-bridge"
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

	fx.Invoke(func(lc fx.Lifecycle, log pdlog.Logger, w *worker.ConsumerWorker) {
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
			func(producer pdkafka.Producer) messaging.Publisher {
				return messagingkafka.NewPublisher(producer)
			},
			fx.ParamTags(`name:"kafka-onboarding-producer"`),
		),
		fx.Annotate(
			func(store core.ConnectionStore) messaging.OutboxStore {
				return core.NewOutboxStoreAdapter(store)
			},
		),
		fx.Annotate(
			NewRuntimeConfig,
			fx.ResultTags(`name:"onboarding-consul-bridge-runtime"`),
		),
		fx.Annotate(
			func(log pdlog.Logger, cfg messaging.ConsumerRuntimeConfig) messaging.Observer {
				return messaging.NewLoggingObserver(log, onboardingConsumerName, cfg)
			},
			fx.ParamTags(``, `name:"onboarding-consul-bridge-runtime"`),
		),
		fx.Annotate(
			consulbridge.NewRegistry,
			fx.As(new(messaging.Handler)),
		),
		fx.Annotate(
			worker.NewConsumerGroupRunner,
			fx.ParamTags(`name:"kafka-onboarding-consumer-group-factory"`, `name:"kafka-onboarding-config"`),
		),
		fx.Annotate(
			func() string {
				return "onboarding"
			},
			fx.ResultTags(`name:"mongo-onboarding-db"`),
		),
		fx.Annotate(eventstore.NewMongoStore, fx.As(new(core.ConnectionStore))),
		core.NewInfraManager,
		worker.NewOutboxWorker,
		fx.Annotate(
			worker.NewConsumerWorker,
			fx.ParamTags(``, ``, ``, `name:"onboarding-consul-bridge-runtime"`, ``, ``),
		),

		// Infras API
		fx.Annotate(infrasmanager.NewService, fx.As(new(infrasinputport.Usecase))),
		fx.Annotate(
			infrasmanager.NewController,
			fx.As(new(Controller)),
			fx.ResultTags(`group:"onboarding-controllers"`),
		),
	)

	StoreCtrlProvider = fx.Provide(
		fx.Annotate(store.NewStoreService, fx.As(new(storeinputport.Usecase))),
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

func NewRuntimeConfig(k *koanf.Koanf) messaging.ConsumerRuntimeConfig {
	cfg := messaging.LoadConsumerRuntimeConfig(
		k,
		onboardingConsumerRuntimePath,
		messaging.DefaultConsumerRuntimeConfig(onboardingConsumerName),
	)
	cfg.Idempotency.ConsumerName = onboardingConsumerName
	return cfg
}
