package infrasmanager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pdlog "github.com/tuannm99/podzone/pkg/pdlogv2"
	"go.uber.org/fx"
)

type InfrasController struct {
	logger pdlog.Logger
}

type StoreControllerParams struct {
	fx.In

	Logger pdlog.Logger
}

func NewController(params StoreControllerParams) *InfrasController {
	return &InfrasController{
		logger: params.Logger,
	}
}

func (c *InfrasController) RegisterRoutes(r *gin.RouterGroup) {
	infras := r.Group("/infras")
	{
		infras.GET("/connections", c.ListConnections)
	}
}

func (c *InfrasController) ListConnections(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "ok")
}
