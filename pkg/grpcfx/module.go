package grpcfx

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Module = fx.Options(
	fx.Provide(newGRPCServer),
	fx.Invoke(startGrpcServer),
)

func newGRPCServer() *grpc.Server {
	return grpc.NewServer()
}
