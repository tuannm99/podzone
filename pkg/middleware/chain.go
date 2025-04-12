package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func Default(logger *zap.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		loggerMiddleware(logger),
		healthMiddleware(logger),
	}
}
