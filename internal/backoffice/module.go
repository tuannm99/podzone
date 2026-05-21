package backoffice

import (
	"go.uber.org/fx"

	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	backofficecatalog "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	backofficerouting "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	backofficestore "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	partnerdirectory "github.com/tuannm99/podzone/internal/backoffice/infrastructure/partnerdirectory"
	catalogrepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/catalog"
	routingrepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/routing"
	storerepo "github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository/store"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
)

// Module provides all backoffice dependencies (GraphQL, domain, infra)
var Module = fx.Options(
	fx.Provide(
		func() *pdtenantdb.Config { return &pdtenantdb.Config{} },
		boconfig.NewConfigFromKoanf,
		fx.Annotate(NewTenantAuthorizer, fx.As(new(TenantAuthorizer))),
		fx.Annotate(NewTenantBootstrapper, fx.As(new(TenantBootstrapper))),
		fx.Annotate(partnerdirectory.New, fx.As(new(outputport.PartnerDirectory))),

		// --- Infrastructure layer ---
		fx.Annotate(storerepo.New, fx.As(new(outputport.StoreRepository))),
		fx.Annotate(catalogrepo.New, fx.As(new(outputport.ProductSetupRepository))),
		fx.Annotate(routingrepo.New, fx.As(new(outputport.OrderRoutingRepository))),

		// --- Domain layer ---
		fx.Annotate(backofficestore.NewStoreInteractor, fx.As(new(inputport.StoreUsecase))),
		fx.Annotate(backofficecatalog.NewProductSetupInteractor, fx.As(new(inputport.ProductSetupUsecase))),
		fx.Annotate(backofficerouting.NewOrderRoutingInteractor, fx.As(new(inputport.OrderRoutingUsecase))),

		// --- GraphQL resolver root ---
		resolver.NewResolver,
	),

	pdtenantdb.Module,
	graphqlModule,
)
