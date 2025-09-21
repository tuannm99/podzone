package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/auth"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpostgres"
	"github.com/tuannm99/podzone/pkg/pdredis"
)

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(
		pdlog.ModuleFor("podzone_admin_auth"),
		pdconfig.Module,

		pdpostgres.ModuleFor("auth"),
		pdredis.ModuleFor("auth"),

		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,
		auth.Module,
	)
}
