package pdglobalmiddleware

import (
	"net/http"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggerMiddleware(logger pdlog.Logger) func(next http.Handler) http.Handler {
	logger.Debug("register logging middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			if r.URL.Path == "/healthz" {
				return
			}

			duration := time.Since(start)
			logger.Info("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", rw.statusCode,
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
				"duration", duration,
			)
		})
	}
}
