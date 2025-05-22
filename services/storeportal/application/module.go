package application

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/services/storeportal/application/services"
	"github.com/tuannm99/podzone/services/storeportal/domain/repositories"
)

// Module provides application services
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			func(repo repositories.StoreRepository) *services.StoreService {
				return services.NewStoreService(repo)
			},
			fx.ParamTags(`name:"store-repository"`),
		),
	),
)
