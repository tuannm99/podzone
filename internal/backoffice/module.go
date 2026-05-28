package backoffice

import (
	"go.uber.org/fx"

	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	backofficecatalog "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	cataloginputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/inputport"
	catalogoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/outputport"
	backofficerouting "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	backofficestore "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	storeinputport "github.com/tuannm99/podzone/internal/backoffice/domain/store/inputport"
	storeoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/store/outputport"
	partnerdirectory "github.com/tuannm99/podzone/internal/backoffice/infrastructure/partnerdirectory"
	catalogrepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/catalog"
	routingrepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/routing"
	storerepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/store"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/storeaccess"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/tenancy"
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
		fx.Annotate(partnerdirectory.New, fx.As(new(routingoutputport.PartnerDirectory))),

		// --- Infrastructure layer ---
		fx.Annotate(storerepo.New, fx.As(new(storeoutputport.StoreRepository))),
		fx.Annotate(catalogrepo.New, fx.As(new(catalogoutputport.ProductSetupRepository))),
		fx.Annotate(routingrepo.New, fx.As(new(routingoutputport.OrderRoutingRepository))),

		// --- Domain layer ---
		fx.Annotate(backofficestore.NewStoreInteractor, fx.As(new(storeinputport.StoreUsecase))),
		fx.Annotate(backofficecatalog.NewProductSetupInteractor, fx.As(new(cataloginputport.ProductSetupUsecase))),
		fx.Annotate(backofficerouting.NewOrderRoutingInteractor, fx.As(new(routinginputport.OrderRoutingUsecase))),

		// --- GraphQL resolver root ---
		resolver.NewResolver,
	),

	pdtenantdb.Module,
	graphqlModule,
)
