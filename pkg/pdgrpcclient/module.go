package pdgrpcclient

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Module = fx.Options(
	fx.Provide(newGRPCClient),
)

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    pdlogv2.Logger
	Host      string
	Port      string
}

func newGRPCClient(p Params) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		p.Host+":"+p.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		p.Logger.Error("Failed to connect to gRPC server", "error", err)
		return nil, err
	}

	p.Logger.Info("gRPC client connected", "host", p.Host, "port", p.Port)

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Closing gRPC client connection")
			return conn.Close()
		},
	})

	return conn, nil
}
