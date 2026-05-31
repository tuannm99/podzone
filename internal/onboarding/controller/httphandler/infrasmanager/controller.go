package infrasmanager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
)

type Controller struct {
	logger  pdlog.Logger
	service infrasinputport.Usecase
}

type ControllerParams struct {
	fx.In

	Logger        pdlog.Logger
	InfrasUsecase infrasinputport.Usecase
}

func NewController(params ControllerParams) *Controller {
	return &Controller{
		logger:  params.Logger,
		service: params.InfrasUsecase,
	}
}

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	infras := r.Group("/infras")
	{
		// Connections
		infras.GET("/connections", c.ListConnections)
		infras.GET("/connections/:infraType/:name", c.GetConnection)
		infras.POST("/connections", c.UpsertConnection)
		infras.DELETE("/connections/:infraType/:name", c.DeleteConnection)

		// Events (history)
		infras.GET("/events", c.ListEvents)
	}
}

func (c *Controller) UpsertConnection(ctx *gin.Context) {
	tenantID, ok := toolkit.GetTenantIDFromGinCtx(ctx)
	if !ok {
		return
	}

	var req infrasinputport.UpsertConnectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}

	resp, err := c.service.ManualUpsertConnection(
		ctx, tenantID, req, toolkit.ExtractActorFromGinCtx(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *Controller) DeleteConnection(ctx *gin.Context) {
	tenantID, ok := toolkit.GetTenantIDFromGinCtx(ctx)
	if !ok {
		return
	}

	infraType := entity.InfraType(ctx.Param("infraType"))
	name := ctx.Param("name")

	corrID, err := c.service.DeleteConnection(
		ctx, tenantID, infraType, name, toolkit.ExtractActorFromGinCtx(ctx))
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal_error", "message": err.Error(), "correlation_id": corrID},
		)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true, "correlation_id": corrID})
}

func (c *Controller) GetConnection(ctx *gin.Context) {
	tenantID, ok := toolkit.GetTenantIDFromGinCtx(ctx)
	if !ok {
		return
	}

	infraType := entity.InfraType(ctx.Param("infraType"))
	name := ctx.Param("name")

	item, err := c.service.GetConnection(ctx, tenantID, infraType, name)
	if err != nil {
		if err == entity.ErrConnectionNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (c *Controller) ListConnections(ctx *gin.Context) {
	tenantID, ok := toolkit.GetTenantIDFromGinCtx(ctx)
	if !ok {
		return
	}

	infraType := entity.InfraType(ctx.Query("infra_type")) // optional: "" => all
	includeDeleted := ctx.Query("include_deleted") == "true"

	limit := toolkit.ParseInt(ctx.Query("limit"), 50)
	offset := toolkit.ParseInt(ctx.Query("offset"), 0)

	items, err := c.service.ListConnections(
		ctx, tenantID, infraType, includeDeleted, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, infrasinputport.ListConnectionsResponse{Items: items})
}

func (c *Controller) ListEvents(ctx *gin.Context) {
	tenantID, ok := toolkit.GetTenantIDFromGinCtx(ctx)
	if !ok {
		return
	}

	infraType := entity.InfraType(ctx.Query("infra_type")) // optional
	name := ctx.Query("name")                              // optional
	corrID := ctx.Query("correlation_id")                  // optional

	limit := toolkit.ParseInt(ctx.Query("limit"), 50)
	offset := toolkit.ParseInt(ctx.Query("offset"), 0)

	items, err := c.service.ListEvents(
		ctx, tenantID, infraType, name, corrID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, infrasinputport.ListEventsResponse{Items: items})
}
