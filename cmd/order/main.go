package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcclientfx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/mongofx"
	"github.com/tuannm99/podzone/pkg/postgresfx"
	"github.com/tuannm99/podzone/pkg/redisfx"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,

		postgresfx.ModuleFor("users", toolkit.FallbackEnv("PG_USERS_URI", "postgres://localhost:5432/users")),
		postgresfx.ModuleFor(
			"analytics",
			toolkit.FallbackEnv("PG_ANALYTICS_URI", "postgres://localhost:5432/analytics"),
		),
		mongofx.ModuleFor("audit", toolkit.FallbackEnv("MONGO_AUDIT_URI", "mongodb://localhost:27017/audit")),

		redisfx.Module,
		grpcfx.Module,
		grpcclientfx.Module,
		grpcgatewayfx.Module,
		globalmiddlewarefx.Module,
	)
	app.Run()
}
