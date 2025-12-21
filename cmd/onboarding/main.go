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

var connOpts = fx.Options(
	pdmongo.ModuleFor("onboarding"),
	onboarding.Module,
)

func main() {
	newAppContainer(connOpts).Run()
}

func newAppContainer(extra ...fx.Option) *fx.App {
	_ = godotenv.Load()
	return fx.New(
		pdconfig.Module,
		pdlog.Module,
		pdglobalmiddleware.CommonGinMiddlewareModule,
		pdhttp.Module,

		fx.Options(extra...),
	)
}
