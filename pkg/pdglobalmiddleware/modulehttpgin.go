package pdglobalmiddleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdlogv2"
	"go.uber.org/fx"
)

func ginLoggerMiddleware(logger pdlogv2.Logger) pdhttp.Middleware {
	return func(r *gin.Engine) {
		r.Use(func(c *gin.Context) {
			start := time.Now()

			c.Next()

			if c.Request.URL.Path == "/healthz" {
				return
			}

			duration := time.Since(start)
			logger.Info("HTTP request",
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"query", c.Request.URL.RawQuery,
				"user_agent", c.Request.UserAgent(),
				"remote_ip", c.ClientIP(),
				"duration", duration,
			)
		})
	}
}

func ginHealthRoute(logger pdlogv2.Logger) pdhttp.RouteRegistrar {
	logger.Debug("Register healthz handler")
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
		fx.Annotate(
			ginHealthRoute,
			fx.ResultTags(`group:"gin-routes"`),
		),
	),
)

