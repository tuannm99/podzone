package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdmongo"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func main() {
	_ = godotenv.Load()

	logger, err := pdlog.NewFrom(
		toolkit.GetEnv("LOG_BACKEND", "zap"),
		context.Background(),
		pdlog.WithLevel(toolkit.GetEnv("DEFAULT_LOG_LEVEL", "debug")),
		pdlog.WithEnv(toolkit.GetEnv("APP_ENV", "dev")),
		pdlog.WithAppName(toolkit.GetEnv("APP_NAME", "podzone_onboarding")),
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

		pdmongo.ModuleFor("onboarding", toolkit.GetEnv("MONGO_ONBOARDING_URI", "mongodb://localhost:27017/onboarding")),

		pdglobalmiddleware.CommonGinMiddlewareModule,
		pdhttp.Module,
		onboarding.Module,
	)

	app.Run()
}
