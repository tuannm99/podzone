package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdmongo"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		pdlog.ModuleFor("podzone_onboarding"),

		pdmongo.ModuleFor("onboarding", toolkit.GetEnv("MONGO_ONBOARDING_URI", "mongodb://localhost:27017/onboarding")),

		pdglobalmiddleware.CommonGinMiddlewareModule,
		pdhttp.Module,
		onboarding.Module,
	)

	app.Run()
}
