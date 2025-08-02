package infrasmanager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type InfrasController struct {
	logger *zap.Logger
}

type StoreControllerParams struct {
	fx.In

	Logger *zap.Logger
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
