package main

import (
	"github.com/joho/godotenv"

	"github.com/tuannm99/podzone/pkg/appfx"
	"github.com/tuannm99/podzone/pkg/grpcclientfx"
	"github.com/tuannm99/podzone/pkg/grpcfx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/redisfx"
)

func main() {
	_ = godotenv.Load()
	appfx.Run(
		logfx.Module,
		redisfx.Module,
		grpcfx.Module,
		grpcclientfx.Module,
		grpcgatewayfx.Module,
	)
}
