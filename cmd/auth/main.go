package main

import (
	"context"
	"log"

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
	logger, err := pdlog.NewFrom(
		toolkit.GetEnv("LOG_BACKEND", "zap"),
		context.Background(),
		pdlog.WithLevel(toolkit.GetEnv("DEFAULT_LOG_LEVEL", "debug")),
		pdlog.WithEnv(toolkit.GetEnv("APP_ENV", "dev")),
		pdlog.WithAppName(toolkit.GetEnv("APP_NAME", "podzone_admin_catalog")),
	)
	if err != nil {
		log.Fatal("error init logger %w", err)
	}

	app := fx.New(
		fx.Provide(func() pdlog.Logger { return logger }),
		fx.Invoke(func(lc fx.Lifecycle, log pdlog.Logger) {
			lc.Append(fx.Hook{
				OnStop: func(context.Context) error { return log.Sync() },
			})
		}),

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
