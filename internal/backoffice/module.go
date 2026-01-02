package backoffice

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	interactor "github.com/tuannm99/podzone/internal/backoffice/domain"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	"github.com/tuannm99/podzone/internal/backoffice/infrastructure/repository"
)

// Module provides all backoffice dependencies (GraphQL, domain, infra)
var Module = fx.Options(
	fx.Provide(
		// --- Infrastructure layer ---
		fx.Annotate(repository.NewStoreRepository, fx.As(new(outputport.StoreRepository))),

		// --- Domain layer ---
		fx.Annotate(interactor.NewStoreInteractor, fx.As(new(inputport.StoreUsecase))),

		// --- GraphQL resolver root ---
		resolver.NewResolver,
	),

	graphqlModule,
)
