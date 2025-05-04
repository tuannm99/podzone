package globalmiddlewarefx

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			loggerMiddleware,
			fx.ResultTags(`group:"http-middleware"`),
		),
		fx.Annotate(
			healthMiddleware,
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
