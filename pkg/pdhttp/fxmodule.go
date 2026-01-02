package pdhttp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type (
	RouteRegistrar func(*gin.Engine)
	Middleware     func(*gin.Engine)
)

var Module = fx.Options(
	fx.Provide(
		NewHttpConfigFromKoanf,
		NewHTTPServer,
	),
	fx.Invoke(
		RegisterMiddlewares,
		RegisterRoutes,
		StartHTTPServer,
	),
)

type NewParams struct {
	fx.In
	Cfg HttpConfig
}

func NewHTTPServer(p NewParams) *gin.Engine {
	router := gin.New()

	// If you want to fully disable trusted proxies, use nil.
	if len(p.Cfg.TrustedProxies) == 0 {
		_ = router.SetTrustedProxies(nil)
	} else {
		_ = router.SetTrustedProxies(p.Cfg.TrustedProxies)
	}

	return router
}

type MWParams struct {
	fx.In
	Router      *gin.Engine
	Middlewares []Middleware `group:"gin-middleware"`
}

func RegisterMiddlewares(p MWParams) {
	for _, m := range p.Middlewares {
		m(p.Router)
	}
}

type RouteParams struct {
	fx.In
	Router *gin.Engine
	Routes []RouteRegistrar `group:"gin-routes"`
}

func RegisterRoutes(p RouteParams) {
	for _, r := range p.Routes {
		r(p.Router)
	}
}

type StartParams struct {
	fx.In
	Lc     fx.Lifecycle
	Logger pdlog.Logger
	Cfg    HttpConfig
	Router *gin.Engine
}

func StartHTTPServer(p StartParams) {
	addr := p.Cfg.Address

	srv := &http.Server{
		Addr:    addr,
		Handler: p.Router,
	}

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				p.Logger.Info("Starting HTTP server", "address", addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					p.Logger.Error("HTTP server stopped with error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
