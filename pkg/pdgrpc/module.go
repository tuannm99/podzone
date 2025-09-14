package pdgrpc

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var Module = fx.Options(
	fx.Provide(newGRPCServer),
	fx.Invoke(startGrpcServer),
)

func newGRPCServer(opts grpc.ServerOption) *grpc.Server {
	server := grpc.NewServer(opts)
	reflection.Register(server)
	return server
}
