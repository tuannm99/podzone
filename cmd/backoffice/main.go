package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/backoffice"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdgraphql"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	kvstores "github.com/tuannm99/podzone/pkg/pdkvstores"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

var connOpts = fx.Options(
	kvstores.Module,
	backoffice.Module,
)

func main() {
	_ = godotenv.Load()
	newAppContainer(connOpts).Run()
}

func newAppContainer(extra ...fx.Option) *fx.App {
	return fx.New(
		pdconfig.Module,
		pdlog.Module,

		// <-- use shared gin/http server
		pdhttp.Module,
		pdgraphql.Module,

		fx.Options(extra...),
	)
}
