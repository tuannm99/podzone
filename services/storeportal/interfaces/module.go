package interfaces

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/services/storeportal/application/services"
	"github.com/tuannm99/podzone/services/storeportal/interfaces/graphql"
)

// Module exports all interface layer components
var Module = fx.Options(
	fx.Provide(
		provideGraphQLResolver,
	),
)

// GraphQLParams contains dependencies for GraphQL resolver
type GraphQLParams struct {
	fx.In

	StoreService *services.StoreService
}

// provideGraphQLResolver creates a new GraphQL resolver
func provideGraphQLResolver(params GraphQLParams) *graphql.Resolver {
	return graphql.NewResolver(params.StoreService)
}
