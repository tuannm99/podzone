package pdgrpcgateway

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdserver"
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

	pdserver.RegisterHTTPServer(
		p.Lifecycle,
		p.Logger,
		":"+httpPort,
		handler,
		pdserver.WithComponent("grpc-gateway"),
	)
}
