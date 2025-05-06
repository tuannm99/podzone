package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/grpcgateway"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,
		globalmiddlewarefx.CommonHttpModule,
		grpcgatewayfx.Module,

		fx.Provide(
			fx.Annotate(
				func() grpcgateway.GatewayRegistrar {
					return &grpcgateway.AuthRegistrar{
						AddrVal: toolkit.FallbackEnv("AUTH_GRPC_ADDR", "localhost:50051"),
					}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() grpcgateway.GatewayRegistrar {
					return &grpcgateway.OrderRegistrar{
						AddrVal: toolkit.FallbackEnv("ORDER_GRPC_ADDR", "localhost:50052"),
					}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),

		fx.Invoke(grpcgateway.RegisterGWHandlers),
	)

	app.Run()
}
