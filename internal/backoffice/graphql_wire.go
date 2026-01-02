package backoffice

import (
	"go.uber.org/fx"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	"github.com/tuannm99/podzone/pkg/pdgraphql"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/vektah/gqlparser/v2/ast"
)

func provideCORSMiddleware() pdhttp.Middleware {
	return func(r *gin.Engine) {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"POST", "GET", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "X-Tenant-ID"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	}
}

type gqlRegistrarParams struct {
	fx.In
	Cfg      pdgraphql.Config
	Resolver *resolver.Resolver
}

func graphQLRegistrar(p gqlRegistrarParams) pdhttp.RouteRegistrar {
	return func(r *gin.Engine) {
		if !p.Cfg.Enabled {
			return
		}

		schema := generated.NewExecutableSchema(generated.Config{Resolvers: p.Resolver})
		srv := handler.New(schema)

		// transports
		srv.AddTransport(transport.Options{})
		srv.AddTransport(transport.GET{})
		srv.AddTransport(transport.POST{})

		// cache & extensions
		srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
		srv.Use(extension.Introspection{})
		srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

		// app-specific extension
		srv.Use(&TenantMiddleware{})

		r.POST(p.Cfg.QueryPath, gin.HandlerFunc(func(c *gin.Context) {
			srv.ServeHTTP(c.Writer, c.Request)
		}))

		if p.Cfg.Playground.Enabled {
			h := playground.Handler("GraphQL", p.Cfg.QueryPath)
			r.GET(p.Cfg.Playground.Path, gin.HandlerFunc(func(c *gin.Context) {
				h.ServeHTTP(c.Writer, c.Request)
			}))
		}
	}
}

var graphqlModule = fx.Options(
	fx.Provide(
		fx.Annotate(graphQLRegistrar, fx.ResultTags(`group:"gin-routes"`)),
		fx.Annotate(provideCORSMiddleware, fx.ResultTags(`group:"gin-middleware"`)),
	),
)
