package globalmiddlewarefx

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
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

func loggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	logger.Debug("register grpc logger interceptor")
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		logger.Info("gRPC Request",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Bool("error", err != nil),
		)

		return resp, err
	}
}
