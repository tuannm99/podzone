package pdgrpcgateway

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/cors"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger  pdlog.Logger
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

	errCh := make(chan error, 1)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				p.Logger.Info("HTTP Gateway started").
					With("address", "http://0.0.0.0:"+httpPort).
					Send()

				if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- err
				}
			}()

			select {
			case err := <-errCh:
				return fmt.Errorf("start HTTP gateway failed: %w", err)
			default:
				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Shutting down HTTP Gateway server")
			return gwServer.Shutdown(ctx)
		},
	})
}
