package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpc"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpprof"
	"github.com/tuannm99/podzone/pkg/pdredis"
	"github.com/tuannm99/podzone/pkg/pdsql"

	"github.com/tuannm99/podzone/internal/auth"
)

var connOpts = fx.Options(
	pdsql.ModuleFor("auth"),
	pdredis.ModuleFor("auth"),

	auth.Module,
)

func main() {
	newAppContainer(connOpts).Run()
}

func newAppContainer(extra ...fx.Option) *fx.App {
	_ = godotenv.Load()

	return fx.New(
		pdconfig.Module,
		pdlog.Module,
		pdpprof.Module,
		pdglobalmiddleware.CommonGRPCModule,
		pdgrpc.Module,

		fx.Options(extra...),
	)
}
