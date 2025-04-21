package grpcgatewayfx

import (
	"context"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func startHTTPGateway(
	lc fx.Lifecycle,
	grpcServeMuxOpts []runtime.ServeMuxOption,
	logger *zap.Logger,
) {
	httpPort := os.Getenv("GW_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	grpcGatewayMux := runtime.NewServeMux(grpcServeMuxOpts...)

	gwServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: grpcGatewayMux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("gRPC-Gateway started",
					zap.String("address", "http://0.0.0.0:"+httpPort))
				if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("Failed to start HTTP server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down HTTP Gateway server")
			return gwServer.Shutdown(ctx)
		},
	})
}
