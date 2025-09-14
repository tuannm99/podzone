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
	logger.Debug("register logging middleware").Send()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			if r.URL.Path != "/healthz" {
				logger.Info("HTTP Request").
					With("method", r.Method).
					With("path", r.URL.Path).
					With("query", r.URL.RawQuery).
					With("status", rw.statusCode).
					With("user_agent", r.UserAgent()).
					With("remote_addr", r.RemoteAddr).
					With("duration", duration).
					Send()
			}
		})
	}
}
