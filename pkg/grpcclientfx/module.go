package grpcclientfx

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Module = fx.Options(
	fx.Provide(newGRPCClient),
)

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
	Host      string
	Port      string
}

func newGRPCClient(p Params) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		p.Host+":"+p.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		p.Logger.Error("Failed to connect to gRPC server", zap.Error(err))
		return nil, err
	}

	p.Logger.Info("gRPC client connected", zap.String("host", p.Host), zap.String("port", p.Port))

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Closing gRPC client connection")
			return conn.Close()
		},
	})

	return conn, nil
}
