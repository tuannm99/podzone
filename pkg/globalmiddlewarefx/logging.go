package globalmiddlewarefx

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	logger.Debug("register logging middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", rw.statusCode),
				zap.String("user_agent", r.UserAgent()),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Duration("duration", duration),
			)
		})
	}
}
