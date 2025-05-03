package grpcgatewayfx

import (
	"context"
	"net/http"
	"os"

	"github.com/rs/cors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger  *zap.Logger
	Handler http.Handler
}

func startHTTPGateway(p Params) {
	httpPort := os.Getenv("GW_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	handler := cors.Default().Handler(p.Handler)

	gwServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: handler,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				p.Logger.Info("gRPC-Gateway started",
					zap.String("address", "http://0.0.0.0:"+httpPort))
				if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					p.Logger.Fatal("Failed to start HTTP server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Shutting down HTTP Gateway server")
			return gwServer.Shutdown(ctx)
		},
	})
}
