package backoffice

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/tuannm99/podzone/pkg/postgresfx"
	"github.com/tuannm99/podzone/services/backoffice/handlers/graphql/resolver"
	"github.com/tuannm99/podzone/services/backoffice/repository"
	"github.com/tuannm99/podzone/services/backoffice/service"
)

// Module provides backoffice services
var Module = fx.Options(
	fx.Provide(
		func(logger *zap.Logger) *postgresfx.TenantDBManager {
			config := &postgresfx.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "backoffice",
				SSLMode:  "disable",
			}
			return postgresfx.NewTenantDBManager(config, logger)
		},
		repository.NewStoreRepository,
		service.NewStoreService,
		resolver.NewResolver,
	),
)
