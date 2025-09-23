package backoffice

import (
	"go.uber.org/fx"
	// "github.com/tuannm99/podzone/internal/backoffice/handlers/graphql/resolver"
	// "github.com/tuannm99/podzone/internal/backoffice/repository"
	// "github.com/tuannm99/podzone/internal/backoffice/service"
	// "github.com/tuannm99/podzone/pkg/pdlogv2"
	// "github.com/tuannm99/podzone/pkg/pdpostgres"
)

// Module provides backoffice services
var Module = fx.Options(
	fx.Provide(
	// func(logger pdlogv2.Logger) *pdpostgres.TenantDBManager {
	// 	config := &pdpostgres.Config{
	// 		Host:     "localhost",
	// 		Port:     5432,
	// 		User:     "postgres",
	// 		Password: "postgres",
	// 		DBName:   "backoffice",
	// 		SSLMode:  "disable",
	// 	}
	// 	return pdpostgres.NewTenantDBManager(config, logger)
	// },
	// repository.NewStoreRepository,
	// service.NewStoreService,
	// resolver.NewResolver,
	),
)
