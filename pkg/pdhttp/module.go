package pdhttp

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewHTTPServer),
	fx.Invoke(StartHTTPServer),
)

type (
	RouteRegistrar func(*gin.Engine)
	Middleware     func(*gin.Engine)
)

type Params struct {
	fx.In
	Lc              fx.Lifecycle
	Logger          pdlogv2.Logger
	Middlewares     []Middleware     `group:"gin-middleware"`
	RegistrarRoutes []RouteRegistrar `group:"gin-routes"`
}

func NewHTTPServer(p Params) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)

	for _, m := range p.Middlewares {
		m(router)
	}

	for _, r := range p.RegistrarRoutes {
		r(router)
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8000"
	}
	addr := fmt.Sprintf(":%s", httpPort)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
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

	return router
}

func StartHTTPServer(_ *gin.Engine) {
	// only used for triggering fx.Lifecycle
}
