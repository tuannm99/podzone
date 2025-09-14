package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
	"github.com/tuannm99/podzone/pkg/mongofx"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/redisfx"
	"github.com/tuannm99/podzone/pkg/toolkit"
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

		fx.Provide(func() grpcfx.GrpcPortFx {
			return toolkit.GetEnv("GRPC_PORT", "50052")
		}),

		redisfx.ModuleFor("catalog", toolkit.GetEnv("CATALOG_REDIS_ADDR", "redis://localhost:6379/1")),
		mongofx.ModuleFor(
			"catalog",
			toolkit.GetEnv("MONGO_CATALOG_URI", "mongodb://localhost:27017/catalog"),
		),

		globalmiddlewarefx.CommonGRPCModule,
		grpcfx.Module,
	)
	app.Run()
}
