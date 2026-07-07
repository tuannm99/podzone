package infrasmanager

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	infrasinputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	infrasoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport"
	"github.com/tuannm99/podzone/pkg/collection"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
)

type Controller struct {
	logger  pdlog.Logger
	service infrasinputport.Usecase
	authz   infrasoutputport.AccessAuthorizer
}

type ControllerParams struct {
	fx.In

	Logger        pdlog.Logger
	InfrasUsecase infrasinputport.Usecase
	Authorizer    infrasoutputport.AccessAuthorizer
}

func NewController(params ControllerParams) *Controller {
	return &Controller{
		logger:  params.Logger,
		service: params.InfrasUsecase,
		authz:   params.Authorizer,
	}
}

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	infras := r.Group("/infras")
	{
		// Connections
		infras.GET("/connections", c.requireInfrastructureRead, c.ListConnections)
		infras.GET("/connections/:infraType/:name", c.requireInfrastructureRead, c.GetConnection)
		infras.POST("/connections", c.requireInfrastructureManage, c.UpsertConnection)
		infras.DELETE(
			"/connections/:infraType/:name",
			c.requireInfrastructureManage,
			c.DeleteConnection,
		)

		// Events (history)
		infras.GET("/events", c.requireInfrastructureRead, c.ListEvents)

		placements := infras.Group("/placements")
		{
			placements.GET("/:tenantId/status", c.requireInfrastructureRead, c.GetTenantPlacementStatus)
			placements.POST("/:tenantId/reconcile", c.requireInfrastructureManage, c.ReconcileTenantPlacement)
		}

		resources := infras.Group("/resources")
		{
			resources.GET("/database-clusters", c.requireInfrastructureRead, c.ListDatabaseClusters)
			resources.PUT("/database-clusters/:name", c.requireInfrastructureManage, c.UpsertDatabaseCluster)
			resources.POST(
				"/database-clusters/:name/health-check",
				c.requireInfrastructureManage,
				c.CheckDatabaseClusterHealth,
			)
			resources.DELETE("/database-clusters/:name", c.requireInfrastructureManage, c.DeleteDatabaseCluster)
			resources.GET("/kubernetes-clusters", c.requireInfrastructureRead, c.ListKubernetesClusters)
			resources.PUT("/kubernetes-clusters/:name", c.requireInfrastructureManage, c.UpsertKubernetesCluster)
			resources.DELETE("/kubernetes-clusters/:name", c.requireInfrastructureManage, c.DeleteKubernetesCluster)
			resources.GET("/runtime-pools", c.requireInfrastructureRead, c.ListRuntimePools)
			resources.PUT("/runtime-pools/:name", c.requireInfrastructureManage, c.UpsertRuntimePool)
			resources.DELETE("/runtime-pools/:name", c.requireInfrastructureManage, c.DeleteRuntimePool)
		}
	}
}

func (c *Controller) requireInfrastructureRead(ctx *gin.Context) {
	c.authorize(ctx, c.authz.AuthorizeInfrastructureRead)
}

func (c *Controller) requireInfrastructureManage(ctx *gin.Context) {
	c.authorize(ctx, c.authz.AuthorizeInfrastructureManage)
}

func (c *Controller) authorize(
	ctx *gin.Context,
	check func(context.Context, string) error,
) {
	requestCtx := ctx.Request.Context()
	userID, err := toolkit.GetUserID(requestCtx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "authenticated user is required",
		})
		return
	}
	if err := check(requestCtx, userID); err != nil {
		statusCode := http.StatusServiceUnavailable
		if errors.Is(err, entity.ErrAccessDenied) {
			statusCode = http.StatusForbidden
		}
		ctx.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	ctx.Next()
}

func (c *Controller) UpsertConnection(ctx *gin.Context) {
	tenantID, ok := tenantIDFromRequest(ctx)
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
	tenantID, ok := tenantIDFromRequest(ctx)
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
	tenantID, ok := tenantIDFromRequest(ctx)
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
	tenantID, ok := tenantIDFromRequest(ctx)
	if !ok {
		return
	}

	includeDeleted := ctx.Query("include_deleted") == "true"
	query, err := collection.ParseURLValues(ctx.Request.URL.Query(), "collection.")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query = withLegacyFilter(query, "infraType", ctx.Query("infra_type"))

	page, err := c.service.ListConnections(ctx, tenantID, includeDeleted, query)
	if err != nil {
		writeInfrastructureError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, infrasinputport.ListConnectionsResponse{
		Items:    page.Items,
		PageInfo: page.Info(),
	})
}

func (c *Controller) ListEvents(ctx *gin.Context) {
	tenantID, ok := tenantIDFromRequest(ctx)
	if !ok {
		return
	}

	query, err := collection.ParseURLValues(ctx.Request.URL.Query(), "collection.")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query = withLegacyFilter(query, "infraType", ctx.Query("infra_type"))
	query = withLegacyFilter(query, "name", ctx.Query("name"))
	query = withLegacyFilter(query, "correlationId", ctx.Query("correlation_id"))

	page, err := c.service.ListEvents(ctx, tenantID, query)
	if err != nil {
		writeInfrastructureError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, infrasinputport.ListEventsResponse{
		Items:    page.Items,
		PageInfo: page.Info(),
	})
}

func withLegacyFilter(query collection.Query, field string, value string) collection.Query {
	if value == "" {
		return query
	}
	query.Filters = append(query.Filters, collection.Filter{
		Field:    field,
		Operator: collection.FilterEqual,
		Values:   []string{value},
	})
	return query
}

func writeInfrastructureError(ctx *gin.Context, err error) {
	if errors.Is(err, collection.ErrInvalidQuery) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal_error",
		"message": err.Error(),
	})
}

func tenantIDFromRequest(ctx *gin.Context) (string, bool) {
	tenantID, err := toolkit.GetTenantID(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated workspace is required"})
		return "", false
	}
	return tenantID, true
}
