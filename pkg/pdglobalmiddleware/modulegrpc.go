package pdglobalmiddleware

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var CommonGRPCModule = fx.Options(
	fx.Provide(
		fx.Annotate(
			loggerInterceptor,
			fx.ResultTags(`group:"grpc-middleware"`),
		),
	),
	fx.Provide(NewUnaryInterceptor),
)

type GRPCParams struct {
	fx.In

	Middlewares []grpc.UnaryServerInterceptor `group:"grpc-middleware"`
}

type GRPCResult struct {
	fx.Out

	Interceptor grpc.ServerOption
}

func NewUnaryInterceptor(p GRPCParams) GRPCResult {
	chain := grpc.ChainUnaryInterceptor(p.Middlewares...)
	return GRPCResult{Interceptor: chain}
}

func loggerInterceptor(logger pdlog.Logger) grpc.UnaryServerInterceptor {
	logger.Debug("register grpc logger interceptor").Send()
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Skip logging for health checks
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return resp, err
		}

		logger.Info("gRPC Request").
			With("method", info.FullMethod).
			With("duration", duration).
			With("error", err != nil).
			Send()

		return resp, err
	}
}
