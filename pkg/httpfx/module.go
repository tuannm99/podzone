package httpfx

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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
	Middlewares     []Middleware     `group:"http-middleware"`
	RegistrarRoutes []RouteRegistrar `group:"http-routes"`
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

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			httpPort := os.Getenv("HTTP_PORT")
			if httpPort == "" {
				httpPort = "8080"
			}

			addr := fmt.Sprintf(":%s", httpPort)

			go func() {
				if err := router.Run(addr); err != nil && err != http.ErrServerClosed {
					panic(fmt.Sprintf("failed to start HTTP server: %v", err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return router
}

func StartHTTPServer(_ *gin.Engine) {
	// no code here, just triggers fx.Lifecycle
}
