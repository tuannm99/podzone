package main

import (
	"context"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/backoffice"
	"github.com/tuannm99/podzone/services/backoffice/handlers/graphql/generated"
	"github.com/tuannm99/podzone/services/backoffice/handlers/graphql/resolver"
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

	ctx = contextfx.WithTenantID(ctx, tenantID)

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
	handlerURI := toolkit.FallbackEnv("GRAPHQL_QUERY_URI", "/query")
	h := playground.Handler("GraphQL", handlerURI)

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func startServer(lc fx.Lifecycle, resolver *resolver.Resolver, logger *zap.Logger) {
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
	app := fx.New(
		logfx.Module,

		// Provide MongoDB connection
		fx.Provide(
			func() string {
				return toolkit.FallbackEnv("MONGODB_PORTAL_URI", "mongodb://localhost:27017")
			},
			func(uri string) *mongo.Client {
				client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
				if err != nil {
					panic(err)
				}
				return client
			},
		),

		backoffice.Module,

		fx.Invoke(startServer),
	)

	app.Run()
}
