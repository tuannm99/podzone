package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/httpfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/mongofx"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/onboarding"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,
		mongofx.ModuleFor("onboarding", toolkit.GetEnv("MONGO_ONBOARDING_URI", "mongodb://localhost:27017/onboarding")),

		globalmiddlewarefx.CommonGinMiddlewareModule,
		httpfx.Module,
		onboarding.Module,
	)

	app.Run()
}
