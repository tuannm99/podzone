package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/mongofx"
	"github.com/tuannm99/podzone/pkg/redisfx"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,

		redisfx.ModuleFor("catalog", toolkit.FallbackEnv("CATALOG_REDIS_ADDR", "redis://localhost:6379/1")),
		mongofx.ModuleFor("catalog", toolkit.FallbackEnv("MONGO_CATALOG_URI", "mongodb://localhost:27017/catalog")),

		globalmiddlewarefx.CommonGRPCModule,
		grpcfx.Module,
	)
	app.Run()
}
