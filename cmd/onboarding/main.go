package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdmongo"
)

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(
		pdlog.ModuleFor("podzone_onboarding"),
		pdconfig.Module(),

		pdmongo.ModuleFor("onboarding"),

		pdglobalmiddleware.CommonGinMiddlewareModule,
		pdhttp.Module,
		onboarding.Module,
	)
}
