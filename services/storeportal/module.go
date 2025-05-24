package storeportal

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/services/storeportal/handlers/graphql"
	"github.com/tuannm99/podzone/services/storeportal/repository"
	"github.com/tuannm99/podzone/services/storeportal/service"
)

// Module provides storeportal services
var Module = fx.Options(
	fx.Provide(
		repository.NewStoreRepository,
		service.NewStoreService,
		graphql.NewResolver,
	),
)
