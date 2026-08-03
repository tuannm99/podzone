package pdgrpcgateway

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdserver"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger  pdlog.Logger
	Handler http.Handler
}

func startHTTPGateway(p Params) {
	httpPort := toolkit.GetEnv("GW_HTTP_PORT", "8088")

	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			o := strings.ToLower(origin)

			if strings.HasPrefix(o, "http://localhost:") ||
				strings.HasPrefix(o, "https://localhost:") ||
				strings.HasPrefix(o, "http://127.0.0.1:") ||
				strings.HasPrefix(o, "https://127.0.0.1:") {
				return true
			}

			allowedSuffixes := []string{
				".podzone.io", ".podzone.dev",
				".podzone.local", ".tuannm.uk",
			}
			for _, suffix := range allowedSuffixes {
				if strings.HasSuffix(o, suffix) {
					return true
				}
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
	// /healthz: liveness only, static - grpc-gateway is a stateless proxy
	// with no dependency of its own to check, unlike Gin services'
	// /readyz (PS10, docs/09-pzep/PZEP-0019-staging-k8s-survivability-track-a.md).
	// This service doesn't go through pdglobalmiddleware.CommonGinMiddlewareModule
	// (it's not a Gin app - p.Handler is grpc-gateway's own runtime mux),
	// so it needs its own healthz route rather than reusing that module.
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/", p.Handler)

	handler := c.Handler(mux)

	pdserver.RegisterHTTPServer(
		p.Lifecycle,
		p.Logger,
		":"+httpPort,
		handler,
		pdserver.WithComponent("grpc-gateway"),
	)
}
