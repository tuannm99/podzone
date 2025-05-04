package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
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

		redisfx.ModuleFor(
			"order",
			toolkit.FallbackEnv("REDIS_ADDR", "redis://localhost:6379/0"),
		),
		globalmiddlewarefx.CommonGRPCModule,
		grpcfx.Module,
	)
	app.Run()
}
