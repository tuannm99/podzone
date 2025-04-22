package middlewarefx

import (
	"net/http"

	"go.uber.org/zap"
)

func healthMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" {
				logger.Debug("Health check request received")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"status":"ok"}`))
				if err != nil {
					logger.Error("Failed to write health check response", zap.Error(err))
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
