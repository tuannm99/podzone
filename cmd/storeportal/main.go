package main

import (
	"context"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/tuannm99/podzone/pkg/contextfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/mongofx"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/storeportal"
	"github.com/tuannm99/podzone/services/storeportal/interfaces/graphql/generated"
)

// TenantMiddleware is a GraphQL middleware that extracts tenant_id from request headers
type TenantMiddleware struct{}

// ExtensionName returns the name of the extension
func (m *TenantMiddleware) ExtensionName() string {
	return "TenantMiddleware"
}

// Validate is called when adding the extension to the server
func (m *TenantMiddleware) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptResponse is called for each response
func (m *TenantMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	// Get the request context
	reqCtx := graphql.GetOperationContext(ctx)
	if reqCtx == nil {
		return next(ctx)
	}

	// Extract tenant_id from request header
	tenantID := reqCtx.Headers.Get("X-Tenant-ID")
	if tenantID == "" {
		// Return error if tenant_id is not provided
		return graphql.ErrorResponse(ctx, "tenant_id is required")
	}

	// Add tenant_id to the context
	ctx = contextfx.WithTenantID(ctx, tenantID)

	return next(ctx)
}

func graphqlHandler(resolver generated.ResolverRoot) gin.HandlerFunc {
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
	handlerURI := toolkit.FallbackEnv("GRAPHQL_QUERY_URI", "/query")
	h := playground.Handler("GraphQL", handlerURI)

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func startServer(lc fx.Lifecycle, resolver generated.ResolverRoot, logger *zap.Logger) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

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
			logger.Info("Starting server", zap.String("port", port))
			go r.Run(":" + port)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down server")
			return nil
		},
	})
}

func main() {
	app := fx.New(
		logfx.Module,

		// Provide MongoDB connection
		mongofx.ModuleFor("storeportal", toolkit.FallbackEnv("MONGODB_URI", "mongodb://localhost:27017")),

		// Include the storeportal module
		storeportal.Module,

		// Start the server
		fx.Invoke(startServer),
	)

	app.Run()
}
