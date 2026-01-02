package pdgraphql

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type registerParams struct {
	fx.In
	Router *gin.Engine
	Cfg    Config

	Query      gin.HandlerFunc `name:"pdgraphql-query"      optional:"true"`
	Playground gin.HandlerFunc `name:"pdgraphql-playground" optional:"true"`
}

func RegisterGraphQLRoutes(p registerParams) error {
	if !p.Cfg.Enabled {
		return nil
	}

	if p.Query == nil {
		return fmt.Errorf(`pdgraphql enabled but missing gin.HandlerFunc name:"pdgraphql-query"`)
	}
	p.Router.POST(p.Cfg.QueryPath, p.Query)

	if p.Cfg.Playground.Enabled {
		if p.Playground == nil {
			return fmt.Errorf(`pdgraphql enabled but missing gin.HandlerFunc name:"pdgraphql-playground"`)
		}
		p.Router.GET(p.Cfg.Playground.Path, p.Playground)
	}
	return nil
}

var Module = fx.Options(
	fx.Provide(NewConfigFromKoanf),
	fx.Invoke(RegisterGraphQLRoutes),
)
