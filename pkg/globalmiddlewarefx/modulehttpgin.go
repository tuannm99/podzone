package globalmiddlewarefx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/pkg/httpfx"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func ginLoggerMiddleware(logger pdlog.Logger) httpfx.Middleware {
	return func(r *gin.Engine) {
		r.Use(func(c *gin.Context) {
			start := time.Now()

			c.Next()

			duration := time.Since(start)
			status := c.Writer.Status()
			method := c.Request.Method
			path := c.Request.URL.Path

			if c.Request.URL.Path != "/healthz" {
				logger.Info("HTTP request").
					With("status", status).
					With("method", method).
					With("path", path).
					With("duration", duration).
					Send()
			}
		})
	}
}

func ginHealthRoute(logger pdlog.Logger) httpfx.RouteRegistrar {
	logger.Debug("Register healthz handler").Send()
	return func(r *gin.Engine) {
		r.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
}

var CommonGinMiddlewareModule = fx.Options(
	fx.Provide(
		fx.Annotate(
			ginLoggerMiddleware,
			fx.ResultTags(`group:"gin-middleware"`),
		),
		// fx.Annotate(
		// 	GinCORSMiddleware,
		// 	fx.ResultTags(`group:"gin-middleware"`),
		// ),
		fx.Annotate(
			ginHealthRoute,
			fx.ResultTags(`group:"gin-routes"`),
		),
	),
)
