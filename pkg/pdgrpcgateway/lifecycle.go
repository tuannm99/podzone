package pdgrpcgateway

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

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

	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			o := strings.ToLower(origin)

			if strings.HasPrefix(o, "http://localhost:") ||
				strings.HasPrefix(o, "https://localhost:") ||
				strings.HasPrefix(o, "http://127.0.0.1:") ||
				strings.HasPrefix(o, "https://127.0.0.1:") {
				return true
			}

			if strings.Contains(o, "tuannm") || strings.Contains(o, "podzone") {
				return true
			}

			return false
		},

		AllowedMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Authorization", "Content-Type",
			"X-Tenant-ID", "X-Request-ID", "X-User",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	})
	handler := c.Handler(p.Handler)

	gwServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: handler,
	}

	errCh := make(chan error, 1)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				p.Logger.Info("HTTP Gateway started",
					"address", "http://0.0.0.0:"+httpPort,
				)

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
