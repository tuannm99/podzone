package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/auth"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"

	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"github.com/tuannm99/podzone/pkg/pdlogv2/provider"

	"github.com/tuannm99/podzone/pkg/pdpostgres"
	"github.com/tuannm99/podzone/pkg/pdredis"
)

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(

		pdconfig.Module,
		pdlogv2.Module(
			pdlogv2.ViperLoaderFor("logger"),
			pdlogv2.WithProvider("zap", provider.ZapFactory),
			pdlogv2.WithProvider("slog", provider.SlogFactory),
			pdlogv2.WithProvider("mock", provider.MockFactory),
			pdlogv2.WithFallback(provider.ZapFactory),
		),
		pdpostgres.Module(
			pdpostgres.ViperLoaderFor("auth"),
			pdpostgres.WithProvider("real", pdpostgres.RealProvider),
			pdpostgres.WithProvider("mock", pdpostgres.MockProvider),
			pdpostgres.WithFallback(pdpostgres.RealProvider),
			pdpostgres.WithName("auth"),
		),
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
