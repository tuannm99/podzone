package main

import (
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/appfx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/redisfx"
)

func main() {
	appfx.Run(
		fx.Provide(
			logfx.LoggerProvider,
		),
		redisfx.Module,
		grpcfx.Module,
	)
}
