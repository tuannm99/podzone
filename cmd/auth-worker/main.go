package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	authiamprojection "github.com/tuannm99/podzone/internal/auth/infrastructure/messaging/iamprojection"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpprof"
	"github.com/tuannm99/podzone/pkg/pdsql"
)

var connOpts = fx.Options(
	pdsql.ModuleFor("auth"),
	pdkafka.ModuleFor("auth"),
	authiamprojection.Module,
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
		fx.Options(extra...),
	)
}
