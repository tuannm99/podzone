package pdgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger pdlog.Logger
	GRPC   *grpc.Server
	Config *viper.Viper
}

func startGrpcServer(p Params) {
	grpcPort := p.Config.GetString("grpc.port")
	if grpcPort == "" {
		grpcPort = "0"
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", ":"+grpcPort)
			if err != nil {
				return fmt.Errorf("failed to listen on gRPC port %s: %w", grpcPort, err)
			}

			go func() {
				p.Logger.Info("Starting gRPC server", "address", lis.Addr().String())
				if err := p.GRPC.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					p.Logger.Error("gRPC server stopped with error", "error", err)
				}
			}()

			// Register health service
			healthServer := health.NewServer()
			grpc_health_v1.RegisterHealthServer(p.GRPC, healthServer)
			healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Shutting down gRPC server")
			p.GRPC.GracefulStop()
			return nil
		},
	})
}
