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
		pdconfig.Module,
		pdlog.Module,

		pdmongo.Module(
			pdmongo.ViperLoaderFor("onboarding"), // mongo.onboarding.*
			pdmongo.WithProvider("real", pdmongo.RealProvider),
			pdmongo.WithProvider("mock", pdmongo.MockProvider),
			pdmongo.WithFallback(pdmongo.RealProvider),
			pdmongo.WithName("onboarding"), // provide name:"mongo-onboarding"
		),

		pdglobalmiddleware.CommonGinMiddlewareModule,
		pdhttp.Module,
		onboarding.Module,
	)
}
