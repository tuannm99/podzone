package logging

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new structured logger
func NewLogger(level string, env string) (*zap.Logger, error) {
	var config zap.Config

	// Parse log level
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	// Configure based on environment
	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.Level = zap.NewAtomicLevelAt(logLevel)

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// WithContext adds context fields to a logger
func WithContext(logger *zap.Logger, fields ...zapcore.Field) *zap.Logger {
	return logger.With(fields...)
}

// LoggerMiddleware creates a logging middleware that logs HTTP requests
func LoggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture the status code
			rw := &responseWriter{w, http.StatusOK}

			// Process the request
			next.ServeHTTP(rw, r)

			// Log the request
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

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
