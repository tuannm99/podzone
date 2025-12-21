package main

import (
	"context"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/fx"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/tuannm99/podzone/internal/backoffice"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdcontext"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

// TenantMiddleware is a GraphQL middleware that extracts tenant_id from request headers
type TenantMiddleware struct{}

func (m *TenantMiddleware) ExtensionName() string {
	return "TenantMiddleware"
}

func (m *TenantMiddleware) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (m *TenantMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	reqCtx := graphql.GetOperationContext(ctx)
	if reqCtx == nil {
		return next(ctx)
	}

	// Extract tenant_id from request header
	tenantID := reqCtx.Headers.Get("X-Tenant-ID")
	if tenantID == "" {
		return graphql.ErrorResponse(ctx, "tenant_id is required")
	}

	ctx = pdcontext.WithTenantID(ctx, tenantID)

	return next(ctx)
}

func graphqlHandler(resolver *resolver.Resolver) gin.HandlerFunc {
	h := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	}))

	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	h.Use(&TenantMiddleware{})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	handlerURI := toolkit.GetEnv("GRAPHQL_QUERY_URI", "/query")
	h := playground.Handler("GraphQL", handlerURI)

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func startServer(lc fx.Lifecycle, resolver *resolver.Resolver, logger pdlog.Logger) {
	port := toolkit.GetEnv("PORT", "8000")

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Tenant-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/query", graphqlHandler(resolver))
	r.GET("/", playgroundHandler())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting server", "port", port)
			go func() {
				_ = r.Run(":" + port)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down server")
			return nil
		},
	})
}

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(
		pdconfig.Module,
		pdlog.Module,

		// TODO: make pdtenent pool db manager

		backoffice.Module,

		fx.Invoke(startServer),
	)
}
