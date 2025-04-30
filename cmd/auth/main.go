package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcclientfx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/postgresfx"
	"github.com/tuannm99/podzone/pkg/redisfx"
	"github.com/tuannm99/podzone/pkg/toolkit"

	"github.com/tuannm99/podzone/services/auth"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,

		postgresfx.ModuleFor(
			"auth",
			toolkit.FallbackEnv("PG_USERS_URI", "postgres://postgres:postgres@localhost:5432/auth"),
		),
		redisfx.Module,

		grpcfx.Module,
		grpcclientfx.Module,
		grpcgatewayfx.Module,
		globalmiddlewarefx.Module,

		auth.Module,
	)
	app.Run()
}
