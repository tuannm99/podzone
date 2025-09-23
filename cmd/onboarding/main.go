package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"github.com/tuannm99/podzone/pkg/pdlogv2/provider"
	"github.com/tuannm99/podzone/pkg/pdmongo"
)

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(
		pdconfig.Module,

		pdlogv2.Module(
			pdlogv2.Defaults("podzone_onboarding"),
			pdlogv2.WithProvider("zap", provider.ZapFactory),
			pdlogv2.WithProvider("slog", provider.SlogFactory),
			pdlogv2.WithProvider("mock", provider.MockFactory),
			pdlogv2.WithFallback(provider.ZapFactory),
		),

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
