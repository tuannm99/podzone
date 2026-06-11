package backoffice

import (
	"go.uber.org/fx"

	backofficecatalog "github.com/tuannm99/podzone/internal/backoffice/application/catalog"
	backofficeoperations "github.com/tuannm99/podzone/internal/backoffice/application/operations"
	backofficestore "github.com/tuannm99/podzone/internal/backoffice/application/store"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	partnerdirectory "github.com/tuannm99/podzone/internal/backoffice/infrastructure/partnerdirectory"
	catalogrepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/catalog"
	routingrepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/routing"
	storerepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/store"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/storeaccess"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/tenancy"
	"github.com/tuannm99/podzone/pkg/ddd"
	dddinprocess "github.com/tuannm99/podzone/pkg/ddd/inprocess"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
)

// Module provides all backoffice dependencies (GraphQL, domain, infra)
var Module = fx.Options(
	fx.Provide(
		func() *pdtenantdb.Config { return &pdtenantdb.Config{} },
		boconfig.NewConfigFromKoanf,
		fx.Annotate(NewTenantAuthorizer, fx.As(new(TenantAuthorizer))),

		fx.Annotate(NewTenantBootstrapper, fx.As(new(TenantBootstrapper)), fx.As(new(tenancy.Readiness))),
		fx.Annotate(storeaccess.New, fx.As(new(storeaccess.Access))),
		fx.Annotate(tenancy.New, fx.As(new(tenancy.Runtime))),
		fx.Annotate(partnerdirectory.New, fx.As(new(routingctx.PartnerDirectory))),
		fx.Annotate(dddinprocess.NewNoopEventDispatcher, fx.As(new(ddd.EventDispatcher))),
		fx.Annotate(ddd.NewUUIDGenerator, fx.As(new(ddd.IDGenerator))),
		fx.Annotate(ddd.NewSystemClock, fx.As(new(ddd.Clock))),

		// --- Infrastructure layer ---
		fx.Annotate(storerepo.New, fx.As(new(storectx.StoreRepository))),
		fx.Annotate(catalogrepo.New, fx.As(new(catalogctx.ProductSetupRepository))),
		fx.Annotate(
			routingrepo.New,
			fx.As(new(routingctx.OrderRoutingRepository)),
			fx.As(new(orderctx.CustomerOrderQueryRepository)),
		),

		// --- Domain layer ---
		fx.Annotate(backofficestore.NewInteractor, fx.As(new(storectx.StoreUsecase))),
		fx.Annotate(backofficecatalog.NewInteractor, fx.As(new(catalogctx.ProductSetupUsecase))),
		fx.Annotate(backofficeoperations.NewOrderRoutingInteractor, fx.As(new(backofficeoperations.OrderRoutingUsecase))),

		// --- GraphQL resolver root ---
		resolver.NewResolver,
	),

	pdtenantdb.Module,
	graphqlModule,
)
