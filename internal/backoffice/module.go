package backoffice

import (
	"go.uber.org/fx"

	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	interactor "github.com/tuannm99/podzone/internal/backoffice/domain"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	"github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
)

// Module provides all backoffice dependencies (GraphQL, domain, infra)
var Module = fx.Options(
	fx.Provide(
		func() *pdtenantdb.Config { return &pdtenantdb.Config{} },
		boconfig.NewConfigFromKoanf,
		fx.Annotate(NewTenantAuthorizer, fx.As(new(TenantAuthorizer))),

		// --- Infrastructure layer ---
		fx.Annotate(repository.NewStoreRepository, fx.As(new(outputport.StoreRepository))),
		fx.Annotate(repository.NewProductSetupRepository, fx.As(new(outputport.ProductSetupRepository))),

		// --- Domain layer ---
		fx.Annotate(interactor.NewStoreInteractor, fx.As(new(inputport.StoreUsecase))),
		fx.Annotate(interactor.NewProductSetupInteractor, fx.As(new(inputport.ProductSetupUsecase))),

		// --- GraphQL resolver root ---
		resolver.NewResolver,
		fx.Annotate(NewProductSetupRoutes, fx.ResultTags(`group:"gin-routes"`)),
	),

	pdtenantdb.Module,
	graphqlModule,
)
