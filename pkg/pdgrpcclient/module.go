package pdgrpcclient

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlog"
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
	Logger    pdlog.Logger
	Host      string
	Port      string
}

func newGRPCClient(p Params) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		p.Host+":"+p.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		p.Logger.Error("Failed to connect to gRPC server").Err(err)
		return nil, err
	}

	p.Logger.Info("gRPC client connected").With("host", p.Host).With("port", p.Port).Send()

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Closing gRPC client connection")
			return conn.Close()
		},
	})

	return conn, nil
}
