package grpcfx

import (
	"context"
	"net"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	GRPC   *grpc.Server
	Logger *zap.Logger
}

func startGrpcServer(p Params) {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", ":"+grpcPort)
			if err != nil {
				p.Logger.Fatal("Failed to listen on gRPC port",
					zap.String("port", grpcPort),
					zap.Error(err))
				return err
			}

			go func() {
				p.Logger.Info("Starting gRPC server", zap.String("port", grpcPort))
				if err := p.GRPC.Serve(lis); err != nil {
					p.Logger.Fatal("Failed to serve gRPC", zap.Error(err))
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
