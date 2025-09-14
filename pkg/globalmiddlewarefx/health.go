package globalmiddlewarefx

import (
	"net/http"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

func healthMiddleware(logger pdlog.Logger) func(http.Handler) http.Handler {
	logger.Debug("register healthz middleware").Send()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" {
				logger.Debug("Health check request received")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"status":"ok"}`))
				if err != nil {
					logger.Error("Failed to write health check response").Err(err).Send()
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
