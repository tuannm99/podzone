package globalmiddlewarefx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/pkg/httpfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ginLoggerMiddleware(logger *zap.Logger) httpfx.Middleware {
	return func(r *gin.Engine) {
		r.Use(func(c *gin.Context) {
			start := time.Now()

			c.Next()

			duration := time.Since(start)
			status := c.Writer.Status()
			method := c.Request.Method
			path := c.Request.URL.Path

			if c.Request.URL.Path != "/healthz" {
				logger.Info("HTTP request",
					zap.Int("status", status),
					zap.String("method", method),
					zap.String("path", path),
					zap.Duration("duration", duration),
				)
			}
		})
	}
}

func ginHealthRoute(logger *zap.Logger) httpfx.RouteRegistrar {
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
