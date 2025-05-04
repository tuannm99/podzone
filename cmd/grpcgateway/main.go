package main

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,
		globalmiddlewarefx.Module,
		grpcgatewayfx.Module,

		fx.Provide(
			fx.Annotate(
				func() GatewayRegistrar {
					return &AuthRegistrar{
						Addr: toolkit.FallbackEnv("AUTH_GRPC_ADDR", "localhost:50051"),
					}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() GatewayRegistrar {
					return &AuthRegistrar{
						Addr: toolkit.FallbackEnv("ORDER_GRPC_ADDR", "localhost:50051"),
					}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),

		fx.Invoke(RegisterGWHandlers),
	)

	app.Run()
}

type GatewayRegistrar interface {
	Register(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error
}

var _ GatewayRegistrar = (*AuthRegistrar)(nil)

type AuthRegistrar struct {
	Addr string
}

func (r *AuthRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
	return pbAuth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, r.Addr, opts)
}

type GWParams struct {
	fx.In
	Mux        *runtime.ServeMux
	Logger     *zap.Logger
	Registrars []GatewayRegistrar `group:"gateway-registrars"`
}

func RegisterGWHandlers(p GWParams) error {
	p.Logger.Info("Registering HTTP handlers (gRPC-Gateway)")
	ctx := context.Background()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	for _, r := range p.Registrars {
		if err := r.Register(ctx, p.Mux, opts); err != nil {
			p.Logger.Error("Failed to register service", zap.Error(err))
			return err
		}
	}
	return nil
}
