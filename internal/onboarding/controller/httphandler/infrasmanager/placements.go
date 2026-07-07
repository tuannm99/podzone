package infrasmanager

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func (c *Controller) GetTenantPlacementStatus(ctx *gin.Context) {
	tenantID, ok := placementTenantIDFromPath(ctx)
	if !ok {
		return
	}
	status, err := c.service.GetTenantPlacementStatus(ctx.Request.Context(), tenantID)
	if err != nil {
		writePlacementError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, status)
}

func (c *Controller) ReconcileTenantPlacement(ctx *gin.Context) {
	tenantID, ok := placementTenantIDFromPath(ctx)
	if !ok {
		return
	}
	resp, err := c.service.ReconcileTenantPlacement(
		ctx.Request.Context(),
		tenantID,
		toolkit.ExtractActorFromGinCtx(ctx),
	)
	if err != nil {
		writePlacementError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func placementTenantIDFromPath(ctx *gin.Context) (string, bool) {
	tenantID := strings.TrimSpace(ctx.Param("tenantId"))
	if tenantID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_tenant_id",
			"message": "tenant id is required",
		})
		return "", false
	}
	return tenantID, true
}

func writePlacementError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrInvalidInput):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "message": err.Error()})
	case errors.Is(err, entity.ErrPlacementNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "placement_not_found", "message": err.Error()})
	default:
		writeInfrastructureError(ctx, err)
	}
}
