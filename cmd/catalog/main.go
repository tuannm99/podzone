package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdmongo"
	"github.com/tuannm99/podzone/pkg/pdredis"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		pdlog.ModuleFor("podzone_admin_catalog"),

		fx.Provide(func() pdgrpc.GrpcPortFx {
			return toolkit.GetEnv("GRPC_PORT", "50052")
		}),

		pdredis.ModuleFor("catalog", toolkit.GetEnv("CATALOG_REDIS_ADDR", "redis://localhost:6379/1")),
		pdmongo.ModuleFor(
			"catalog",
			toolkit.GetEnv("MONGO_CATALOG_URI", "mongodb://localhost:27017/catalog"),
		),

		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,
	)
	app.Run()
}
