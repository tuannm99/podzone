package pdgrpc

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var Module = fx.Options(
	fx.Provide(newGRPCServer),
	fx.Invoke(startGrpcServer),
)

func newGRPCServer(cfg *koanf.Koanf, opts grpc.ServerOption) *grpc.Server {
	server := grpc.NewServer(opts)
	enableReflection := cfg.Bool("grpc.enable_reflection")
	if !enableReflection && cfg.String("logger.env") == "dev" {
		enableReflection = true
	}
	if enableReflection {
		reflection.Register(server)
	}
	return server
}
