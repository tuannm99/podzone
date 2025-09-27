package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpostgres"
	"github.com/tuannm99/podzone/pkg/pdredis"

	"github.com/tuannm99/podzone/internal/auth"
)

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(
		pdconfig.Module,
		pdlog.Module,
		pdpostgres.ModuleFor("auth"),
		pdredis.Module(
			pdredis.ViperLoaderFor("auth"),
			pdredis.WithProvider("real", pdredis.RealProvider),
			pdredis.WithProvider("mock", pdredis.MockProvider),
			pdredis.WithFallback(pdredis.RealProvider),
			pdredis.WithName("auth"),
		),
		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,
		auth.Module,
	)
}
