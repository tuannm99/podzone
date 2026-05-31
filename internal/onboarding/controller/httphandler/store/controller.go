package store

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	storedomain "github.com/tuannm99/podzone/internal/onboarding/domain/store"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

type Controller struct {
	logger  pdlog.Logger
	service storeinputport.Usecase
}

type ControllerParams struct {
	fx.In

	Logger       pdlog.Logger
	StoreUsecase storeinputport.Usecase
}

func NewController(params ControllerParams) *Controller {
	return &Controller{
		logger:  params.Logger,
		service: params.StoreUsecase,
	}
}

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	requests := r.Group("/requests")
	{
		requests.POST("", c.CreateStoreRequest)
		requests.GET("", c.ListStoreRequests)
		requests.GET("/:id", c.GetStoreRequest)
		requests.PATCH("/:id/status", c.UpdateStoreRequestStatus)
		requests.POST("/:id/approve", c.ApproveStoreRequest)
		requests.POST("/:id/reject", c.RejectStoreRequest)
	}

	legacy := r.Group("/stores")
	{
		legacy.POST("", c.CreateStoreRequest)
		legacy.GET("", c.ListStoreRequests)
		legacy.GET("/:id", c.GetStoreRequest)
		legacy.PATCH("/:id/status", c.UpdateStoreRequestStatus)
		legacy.POST("/:id/approve", c.ApproveStoreRequest)
		legacy.POST("/:id/reject", c.RejectStoreRequest)
	}
}

type CreateStoreRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
	RequestedBy string `json:"requested_by" binding:"required"`
	Name        string `json:"name"         binding:"required"`
	Subdomain   string `json:"subdomain"    binding:"required"`
}

func (c *Controller) CreateStoreRequest(ctx *gin.Context) {
	var req CreateStoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request, err := c.service.CreateStoreRequest(ctx, req.Name, req.Subdomain, req.WorkspaceID, req.RequestedBy)
	if err != nil {
		if err == storedomain.ErrSubdomainTaken {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, storedomain.ErrWorkspaceIDRequired) || errors.Is(err, storedomain.ErrRequestedByRequired) ||
			errors.Is(err, storedomain.ErrNameRequired) || errors.Is(err, storedomain.ErrSubdomainRequired) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create store"})
		return
	}

	ctx.JSON(http.StatusCreated, request)
}

func (c *Controller) ListStoreRequests(ctx *gin.Context) {
	workspaceID := ctx.Query("workspace_id")

	requests, err := c.service.ListStoreRequests(ctx, workspaceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list stores"})
		return
	}

	ctx.JSON(http.StatusOK, requests)
}

func (c *Controller) GetStoreRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	request, err := c.service.GetStoreRequest(ctx, id)
	if err != nil {
		if err == storedomain.ErrStoreNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store request not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get store"})
		return
	}

	ctx.JSON(http.StatusOK, request)
}

type UpdateStoreStatusRequest struct {
	Status storeinputport.RequestStatus `json:"status" binding:"required"`
}

func (c *Controller) UpdateStoreRequestStatus(ctx *gin.Context) {
	id := ctx.Param("id")

	var req UpdateStoreStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.service.UpdateStoreRequestStatus(ctx, id, req.Status)
	if err != nil {
		switch err {
		case storedomain.ErrStoreNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store request not found"})
		case storedomain.ErrInvalidStatus:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update store status"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *Controller) ApproveStoreRequest(ctx *gin.Context) {
	if err := c.service.ApproveStoreRequest(ctx, ctx.Param("id")); err != nil {
		switch err {
		case storedomain.ErrStoreNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store request not found"})
		case storedomain.ErrInvalidStatus:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve store request"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *Controller) RejectStoreRequest(ctx *gin.Context) {
	if err := c.service.RejectStoreRequest(ctx, ctx.Param("id")); err != nil {
		switch err {
		case storedomain.ErrStoreNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store request not found"})
		case storedomain.ErrInvalidStatus:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject store request"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}
