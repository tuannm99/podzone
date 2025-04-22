package middlewarefx

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			func(logger *zap.Logger) func(http.Handler) http.Handler {
				return loggerMiddleware(logger)
			},
			fx.ResultTags(`group:"http-middleware"`),
		),
		fx.Annotate(
			func(logger *zap.Logger) func(http.Handler) http.Handler {
				return healthMiddleware(logger)
			},
			fx.ResultTags(`group:"http-middleware"`),
		),
	),
	fx.Provide(NewChainedHandler),
)

type Params struct {
	fx.In

	Mux         *runtime.ServeMux
	Middlewares []func(http.Handler) http.Handler `group:"http-middleware"`
}

type Result struct {
	fx.Out

	Handler http.Handler
}

func NewChainedHandler(p Params) Result {
	final := http.Handler(p.Mux)
	for i := len(p.Middlewares) - 1; i >= 0; i-- {
		final = p.Middlewares[i](final)
	}
	return Result{Handler: final}
}
