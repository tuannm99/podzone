package pdglobalmiddleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func ginLoggerMiddleware(logger pdlog.Logger) pdhttp.Middleware {
	return func(r *gin.Engine) {
		r.Use(func(c *gin.Context) {
			start := time.Now()

			c.Next()

			if c.Request.URL.Path == "/healthz" || c.Request.URL.Path == "/readyz" {
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

func ginHealthRoute(logger pdlog.Logger) pdhttp.RouteRegistrar {
	logger.Debug("Register healthz handler")
	return func(r *gin.Engine) {
		r.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
}

// ReadinessChecker is an optional dependency check a service can supply to
// get a real /readyz instead of always-ready. See PZEP-0019 (PS10,
// docs/07-problems/platform-scaling-2026-07-23.md): the shared /healthz
// above is liveness-only by design (always 200) - a real readiness check
// needs to know which dependencies a given service actually has, which
// this package cannot know generically. NewReadyRoute is a plain
// constructor, not auto-wired by CommonGinMiddlewareModule - a service
// wires it into its own fx.Provide (group:"gin-routes") with its own
// checker, matching how internal/backoffice does it for its Mongo
// dependency.
type ReadinessChecker func(ctx context.Context) error

// NewReadyRoute registers /readyz. A nil checker makes it behave exactly
// like /healthz (always ready) - safe default for a service not yet ready
// to supply a real check.
func NewReadyRoute(logger pdlog.Logger, checker ReadinessChecker) pdhttp.RouteRegistrar {
	logger.Debug("Register readyz handler")
	return func(r *gin.Engine) {
		r.GET("/readyz", func(c *gin.Context) {
			if checker != nil {
				if err := checker(c.Request.Context()); err != nil {
					logger.Error("readyz check failed", "error", err)
					c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
					return
				}
			}
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
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
