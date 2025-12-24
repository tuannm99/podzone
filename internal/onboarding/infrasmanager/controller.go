package infrasmanager

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type InfrasController struct {
	logger  pdlog.Logger
	service *Service
}

type ControllerParams struct {
	fx.In

	Logger  pdlog.Logger
	Service *Service
}

func NewController(params ControllerParams) *InfrasController {
	return &InfrasController{
		logger:  params.Logger,
		service: params.Service,
	}
}

func (c *InfrasController) RegisterRoutes(r *gin.RouterGroup) {
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

func (c *InfrasController) UpsertConnection(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}

	var req UpsertConnectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}

	resp, err := c.service.ManualUpsertConnection(tenantID, req, extractActor(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *InfrasController) DeleteConnection(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}

	infraType := core.InfraType(ctx.Param("infraType"))
	name := ctx.Param("name")

	corrID, err := c.service.DeleteConnection(tenantID, infraType, name, extractActor(ctx))
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal_error", "message": err.Error(), "correlation_id": corrID},
		)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true, "correlation_id": corrID})
}

func (c *InfrasController) GetConnection(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}

	infraType := core.InfraType(ctx.Param("infraType"))
	name := ctx.Param("name")

	item, err := c.service.GetConnection(tenantID, infraType, name)
	if err != nil {
		if err == core.ErrConnectionNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (c *InfrasController) ListConnections(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}

	infraType := core.InfraType(ctx.Query("infra_type")) // optional: "" => all
	includeDeleted := ctx.Query("include_deleted") == "true"

	limit := parseInt(ctx.Query("limit"), 50)
	offset := parseInt(ctx.Query("offset"), 0)

	items, err := c.service.ListConnections(tenantID, infraType, includeDeleted, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, ListConnectionsResponse{Items: items})
}

func (c *InfrasController) ListEvents(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}

	infraType := core.InfraType(ctx.Query("infra_type")) // optional
	name := ctx.Query("name")                            // optional
	corrID := ctx.Query("correlation_id")                // optional

	limit := parseInt(ctx.Query("limit"), 50)
	offset := parseInt(ctx.Query("offset"), 0)

	items, err := c.service.ListEvents(tenantID, infraType, name, corrID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, ListEventsResponse{Items: items})
}

func getTenantID(ctx *gin.Context) (string, bool) {
	// Prefer header for multi-tenant gateway.
	tenantID := ctx.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		tenantID = ctx.Query("tenant_id")
	}
	if tenantID == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "missing_tenant", "message": "missing tenant id (X-Tenant-ID or tenant_id)"},
		)
		return "", false
	}
	return tenantID, true
}

func extractActor(ctx *gin.Context) map[string]string {
	actor := map[string]string{
		"user":       ctx.GetHeader("X-User"),
		"request_id": ctx.GetHeader("X-Request-ID"),
		"ip":         ctx.ClientIP(),
		"ua":         ctx.GetHeader("User-Agent"),
	}
	// Remove empty to keep event small
	for k, v := range actor {
		if v == "" {
			delete(actor, k)
		}
	}
	return actor
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

