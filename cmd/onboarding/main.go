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
		globalmiddlewarefx.CommonHttpModule,
		httpfx.Module,
		mongofx.ModuleFor("config", toolkit.FallbackEnv("MONGO_CONFIG_URI", "mongodb://localhost:27017/config")),

		onboarding.Module,
	)

	app.Run()
}
