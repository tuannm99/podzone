package grpcfx

import (
	"context"
	"log"
	"net"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GrpcPortFx = string

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger   pdlog.Logger
	GRPC     *grpc.Server
	GrpcPort GrpcPortFx
}

func startGrpcServer(p Params) {
	grpcPort := p.GrpcPort

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", ":"+grpcPort)
			if err != nil {
				log.Fatal("Failed to listen on gRPC port %w", err)
				return err
			}

			go func() {
				p.Logger.Info("Starting gRPC server").With("port", grpcPort).Send()
				if err := p.GRPC.Serve(lis); err != nil {
					log.Fatal("Failed to serve gRPC %w", err)
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
