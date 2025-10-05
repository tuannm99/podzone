package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdmongo"
	"github.com/tuannm99/podzone/pkg/pdredis"
)

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(
		pdconfig.Module,
		pdlog.Module,

		pdredis.ModuleFor("catalog"),
		pdmongo.ModuleFor("catalog"),

		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,
	)
}
