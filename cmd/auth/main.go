package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpostgres"
	"github.com/tuannm99/podzone/pkg/pdredis"
	"github.com/tuannm99/podzone/pkg/toolkit"

	"github.com/tuannm99/podzone/internal/auth"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		pdlog.ModuleFor("podzone_admin_auth"),

		fx.Provide(func() pdgrpc.GrpcPortFx {
			return toolkit.GetEnv("GRPC_PORT", "50051")
		}),

		pdpostgres.ModuleFor(
			"auth",
			toolkit.GetEnv("PG_AUTH_URI", "postgres://postgres:postgres@localhost:5432/auth"),
		),
		pdredis.ModuleFor(
			"auth",
			toolkit.GetEnv("REDIS_ADDR", "redis://localhost:6379/0"),
		),

		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,
		auth.Module,
	)
	app.Run()
}
