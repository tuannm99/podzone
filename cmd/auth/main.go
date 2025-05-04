package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/grpcfx"
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
			toolkit.FallbackEnv("PG_AUTH_URI", "postgres://postgres:postgres@localhost:5432/auth"),
		),
		redisfx.ModuleFor(
			"auth",
			toolkit.FallbackEnv("REDIS_ADDR", "redis://localhost:6379/0"),
		),
		// redisfx.Module,

		grpcfx.Module,
		auth.Module,
	)
	app.Run()
}
