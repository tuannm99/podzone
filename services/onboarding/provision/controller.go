package provision

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ProvisionController struct {
	logger           *zap.Logger
	provisionService *ProvisionService
}

type ProvisionControllerParams struct {
	fx.In

	Logger           *zap.Logger
	ProvisionService *ProvisionService
}

func NewProvisionController(params ProvisionControllerParams) *ProvisionController {
	return &ProvisionController{
		logger:           params.Logger,
		provisionService: params.ProvisionService,
	}
}

func (c *ProvisionController) RegisterRoutes(r *gin.RouterGroup) {
	p := r.Group("/provisions")
	{
		// p.GET("", c.ListStores)
	}
}
