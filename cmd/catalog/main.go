package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"github.com/tuannm99/podzone/pkg/pdlogv2/provider"
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

		pdlogv2.Module(
			pdlogv2.ViperLoaderFor("logger"),
			pdlogv2.WithProvider("zap", provider.ZapFactory),
			pdlogv2.WithProvider("slog", provider.SlogFactory),
			pdlogv2.WithProvider("mock", provider.MockFactory),
			pdlogv2.WithFallback(provider.ZapFactory),
		),

		pdredis.Module(
			pdredis.ViperLoaderFor("catalog"), // redis.catalog.*
			pdredis.WithProvider("real", pdredis.RealProvider),
			pdredis.WithProvider("mock", pdredis.MockProvider),
			pdredis.WithFallback(pdredis.RealProvider),
			pdredis.WithName("catalog"), // provide name:"redis-catalog"
		),

		pdmongo.Module(
			pdmongo.ViperLoaderFor("catalog"), // mongo.catalog.*
			pdmongo.WithProvider("real", pdmongo.RealProvider),
			pdmongo.WithProvider("mock", pdmongo.MockProvider),
			pdmongo.WithFallback(pdmongo.RealProvider),
			pdmongo.WithName("catalog"), // provide name:"mongo-catalog"
		),

		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,
	)
}
