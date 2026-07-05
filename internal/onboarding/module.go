package onboarding

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	kvstorebridge "github.com/tuannm99/podzone/internal/onboarding/controller/eventhandler/kvstorebridge"
	"github.com/tuannm99/podzone/internal/onboarding/controller/httphandler"
	infrascontroller "github.com/tuannm99/podzone/internal/onboarding/controller/httphandler/infrasmanager"
	storecontroller "github.com/tuannm99/podzone/internal/onboarding/controller/httphandler/store"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager"
	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	infrasoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport"
	"github.com/tuannm99/podzone/internal/onboarding/domain/store"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport"
	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/backofficeclient"
	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/iamclient"
	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/messaging/publisher"
	"github.com/tuannm99/podzone/internal/onboarding/infrastructure/messaging/worker"
	placementprovider "github.com/tuannm99/podzone/internal/onboarding/infrastructure/provisioning/provider"
	placementrouter "github.com/tuannm99/podzone/internal/onboarding/infrastructure/provisioning/router"
	infrarepository "github.com/tuannm99/podzone/internal/onboarding/infrastructure/repository/infrasmanager"
	storerepository "github.com/tuannm99/podzone/internal/onboarding/infrastructure/repository/store"
	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdworker"
)

const (
	onboardingConsumerRuntimePath = "messaging.onboarding.consumers.kv_store_bridge"
	onboardingConsumerName        = "onboarding.kv-store-bridge"
)

var Module = fx.Options(
	StoreCtrlProvider,
	InfrasCtrlProvider,

	fx.Provide(
		onboardingconfig.NewAuthConfig,
		httphandler.NewAuthentication,
		iamclient.NewAccessAuthorizer,
		fx.Annotate(provideCORSMiddleware, fx.ResultTags(`group:"gin-middleware"`)),
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

	fx.Invoke(func(
		lc fx.Lifecycle,
		store *infrarepository.MongoStore,
		cfg onboardingconfig.StoreProvisioningConfig,
		log pdlog.Logger,
	) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.Info("Ensuring onboarding infrastructure inventory...")
				if err := store.EnsureIndexes(ctx); err != nil {
					return err
				}
				return store.EnsureConfiguredResourceInventory(ctx, cfg)
			},
		})
	}),

	fx.Invoke(func(lc fx.Lifecycle, log pdlog.Logger, w *worker.OutboxWorker) {
		pdworker.StartWorker(lc, log, w)
	}),

	fx.Invoke(func(lc fx.Lifecycle, log pdlog.Logger, w *worker.ConsumerWorker) {
		pdworker.StartWorker(lc, log, w)
	}),

	fx.Invoke(func(lc fx.Lifecycle, log pdlog.Logger, w *worker.StoreProvisioningWorker) {
		pdworker.StartWorker(lc, log, w)
	}),
)

var (
	InfrasCtrlProvider = fx.Provide(
		// --- Infrastructure layer ---
		publisher.NewKVStorePublisher,
		fx.Annotate(
			func(producer pdkafka.Producer) messaging.Publisher {
				return messagingkafka.NewPublisher(producer)
			},
			fx.ParamTags(`name:"kafka-onboarding-producer"`),
		),
		fx.Annotate(
			NewRuntimeConfig,
			fx.ResultTags(`name:"onboarding-kv-store-bridge-runtime"`),
		),
		fx.Annotate(
			func(log pdlog.Logger, cfg messaging.ConsumerRuntimeConfig) messaging.Observer {
				return messaging.NewLoggingObserver(log, onboardingConsumerName, cfg)
			},
			fx.ParamTags(``, `name:"onboarding-kv-store-bridge-runtime"`),
		),
		fx.Annotate(
			kvstorebridge.NewRegistry,
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
		infrarepository.NewMongoStore,
		func(store *infrarepository.MongoStore) infrasoutputport.ConnectionStore {
			return store
		},
		func(store *infrarepository.MongoStore) infrasoutputport.PlacementRepository {
			return store
		},
		func(store *infrarepository.MongoStore) infrasoutputport.PlacementPlanRepository {
			return store
		},
		func(store *infrarepository.MongoStore) infrasoutputport.ResourceInventoryRepository {
			return store
		},
		func(store *infrarepository.MongoStore) messaging.OutboxStore {
			return store
		},
		placementprovider.NewProvider,
		func(provider *placementprovider.Provider) infrasoutputport.PlacementPlanner {
			return provider
		},
		func(provider *placementprovider.Provider) infrasoutputport.StorageProvisioner {
			return provider
		},
		fx.Annotate(
			placementrouter.NewPlacementRouteReader,
			fx.As(new(infrasoutputport.PlacementRouteReader)),
			fx.As(new(infrasoutputport.PlacementRouteWriter)),
		),
		worker.NewOutboxWorker,
		fx.Annotate(
			worker.NewConsumerWorker,
			fx.ParamTags(``, ``, ``, `name:"onboarding-kv-store-bridge-runtime"`, ``, ``),
		),

		// --- Domain layer ---
		fx.Annotate(infrasmanager.NewInteractorWithParams, fx.As(new(infrasinputport.Usecase))),
		func(authorizer *iamclient.AccessAuthorizer) infrasoutputport.AccessAuthorizer {
			return authorizer
		},

		// --- HTTP handler layer ---
		fx.Annotate(
			infrascontroller.NewController,
			fx.As(new(Controller)),
			fx.ResultTags(`group:"onboarding-controllers"`),
		),
	)

	StoreCtrlProvider = fx.Provide(
		// --- Infrastructure layer ---
		onboardingconfig.NewStoreProvisioningConfig,
		onboardingconfig.NewStoreProvisioningDomainConfig,
		func(authorizer *iamclient.AccessAuthorizer) storeoutputport.AccessAuthorizer {
			return authorizer
		},
		fx.Annotate(
			backofficeclient.NewStoreFinalizer,
			fx.As(new(storeoutputport.OperationalStoreFinalizer)),
		),
		fx.Annotate(
			storerepository.New,
			fx.As(new(storeoutputport.StoreRepository)),
		),
		worker.NewStoreProvisioningWorker,

		// --- Domain layer ---
		fx.Annotate(store.NewStoreInteractor, fx.As(new(storeinputport.Usecase))),

		// --- HTTP handler layer ---
		fx.Annotate(
			storecontroller.NewController,
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
	Auth        *httphandler.Authentication
	Controllers []Controller `group:"onboarding-controllers"`
}

func RegisterHTTPRoutes(p RegisterRoutesParams) pdhttp.RouteRegistrar {
	p.Logger.Info("Registering Onboarding HTTP handler")
	return func(r *gin.Engine) {
		v1 := r.Group("/onboarding/v1")
		v1.Use(p.Auth.RequireUser)
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
