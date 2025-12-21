package pdgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/knadh/koanf/v2"
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
	Config *koanf.Koanf
}

func startGrpcServer(p Params) {
	grpcPort := p.Config.Int("grpc.port")
	if grpcPort == 0 {
		grpcPort = 50051 // sensible default (or keep 0 if you want random port)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
			if err != nil {
				return fmt.Errorf("failed to listen on gRPC port %d: %w", grpcPort, err)
			}

			// Register health service BEFORE Serve (best practice)
			healthServer := health.NewServer()
			grpc_health_v1.RegisterHealthServer(p.GRPC, healthServer)
			healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

			go func() {
				p.Logger.Info("Starting gRPC server", "address", lis.Addr().String())
				if err := p.GRPC.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					p.Logger.Error("gRPC server stopped with error", "error", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Shutting down gRPC server")
			p.GRPC.GracefulStop()
			return nil
		},
	})
}
