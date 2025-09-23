package pdglobalmiddleware

import (
	"net/http"

	"github.com/tuannm99/podzone/pkg/pdlogv2"
)

func healthMiddleware(logger pdlogv2.Logger) func(http.Handler) http.Handler {
	logger.Debug("register healthz middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" {
				logger.Debug("Health check request received")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
					logger.Error("Failed to write health check response", "error", err)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
