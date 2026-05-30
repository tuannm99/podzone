package onboarding

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager"
	infrascontroller "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/controller"
	consulbridge "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/controller/eventhandler/consulbridge"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/infrastructure/publisher"
	infrarepository "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/infrastructure/repository"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/infrastructure/worker"
	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/inputport"
	infrasoutputport "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/outputport"
	"github.com/tuannm99/podzone/internal/onboarding/store"
	storecontroller "github.com/tuannm99/podzone/internal/onboarding/store/controller"
	storerepository "github.com/tuannm99/podzone/internal/onboarding/store/infrastructure/repository"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/store/inputport"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/store/outputport"
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

	fx.Invoke(func(lc fx.Lifecycle, store storeoutputport.StoreRepository, log pdlog.Logger) {
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
		func() map[entity.InfraType]entity.InfraProvisioner {
			return map[entity.InfraType]entity.InfraProvisioner{}
		},
		publisher.NewConsulPublisher,
		fx.Annotate(
			func(producer pdkafka.Producer) messaging.Publisher {
				return messagingkafka.NewPublisher(producer)
			},
			fx.ParamTags(`name:"kafka-onboarding-producer"`),
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
		fx.Annotate(
			infrarepository.NewMongoStore,
			fx.As(new(infrasoutputport.ConnectionStore), new(messaging.OutboxStore)),
		),
		worker.NewOutboxWorker,
		fx.Annotate(
			worker.NewConsumerWorker,
			fx.ParamTags(``, ``, ``, `name:"onboarding-consul-bridge-runtime"`, ``, ``),
		),

		// Infras API
		fx.Annotate(infrasmanager.NewInteractor, fx.As(new(infrasinputport.Usecase))),
		fx.Annotate(
			infrascontroller.NewController,
			fx.As(new(Controller)),
			fx.ResultTags(`group:"onboarding-controllers"`),
		),
	)

	StoreCtrlProvider = fx.Provide(
		fx.Annotate(
			storerepository.New,
			fx.As(new(storeoutputport.StoreRepository)),
		),
		fx.Annotate(store.NewStoreInteractor, fx.As(new(storeinputport.Usecase))),
		fx.Annotate(
			storecontroller.NewStoreController,
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
