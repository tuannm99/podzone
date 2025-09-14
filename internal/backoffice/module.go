package backoffice

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/backoffice/handlers/graphql/resolver"
	"github.com/tuannm99/podzone/internal/backoffice/repository"
	"github.com/tuannm99/podzone/internal/backoffice/service"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/postgresfx"
)

// Module provides backoffice services
var Module = fx.Options(
	fx.Provide(
		func(logger pdlog.Logger) *postgresfx.TenantDBManager {
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
