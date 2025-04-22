package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/grpcclientfx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/middlewarefx"
	"github.com/tuannm99/podzone/pkg/redisfx"

	"github.com/tuannm99/podzone/services/auth"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,
		redisfx.Module,
		grpcfx.Module,
		grpcclientfx.Module,
		grpcgatewayfx.Module,
		middlewarefx.Module,
		auth.Module,
	)
	app.Run()
}
