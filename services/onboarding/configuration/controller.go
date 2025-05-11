package configuration

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type OnboardingControllerParams struct {
	fx.In
	Logger *zap.Logger
}

func NewOnboardingController(p OnboardingControllerParams) *OnboardingController {
	return &OnboardingController{
		logger: p.Logger,
	}
}

type OnboardingController struct {
	logger *zap.Logger
}

func (c *OnboardingController) GetAll(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "Okkkkkkkk"})
}
