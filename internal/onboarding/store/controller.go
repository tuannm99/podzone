package store

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pdlog "github.com/tuannm99/podzone/pkg/pdlogv2"
	"go.uber.org/fx"
)

type StoreController struct {
	logger  pdlog.Logger
	service *StoreService
}

type StoreControllerParams struct {
	fx.In

	Logger       pdlog.Logger
	StoreService *StoreService
}

func NewStoreController(params StoreControllerParams) *StoreController {
	return &StoreController{
		logger:  params.Logger,
		service: params.StoreService,
	}
}

func (c *StoreController) RegisterRoutes(r *gin.RouterGroup) {
	stores := r.Group("/stores")
	{
		stores.POST("", c.CreateStore)
		stores.GET("", c.ListStores)
		stores.GET("/:id", c.GetStore)
		stores.PATCH("/:id/status", c.UpdateStoreStatus)
	}
}

type CreateStoreRequest struct {
	Name      string `json:"name"      binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
}

func (c *StoreController) CreateStore(ctx *gin.Context) {
	var req CreateStoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get owner ID from auth context
	ownerID := "temp-owner-id"

	store, err := c.service.CreateStore(ctx, req.Name, req.Subdomain, ownerID)
	if err != nil {
		if err == ErrSubdomainTaken {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create store"})
		return
	}

	ctx.JSON(http.StatusCreated, store)
}

func (c *StoreController) ListStores(ctx *gin.Context) {
	ownerID := "temp-owner-id"

	stores, err := c.service.GetStoresByOwner(ctx, ownerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list stores"})
		return
	}

	ctx.JSON(http.StatusOK, stores)
}

func (c *StoreController) GetStore(ctx *gin.Context) {
	id := ctx.Param("id")

	store, err := c.service.GetStore(ctx, id)
	if err != nil {
		if err == ErrStoreNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get store"})
		return
	}

	ctx.JSON(http.StatusOK, store)
}

type UpdateStoreStatusRequest struct {
	Status StoreStatus `json:"status" binding:"required"`
}

func (c *StoreController) UpdateStoreStatus(ctx *gin.Context) {
	id := ctx.Param("id")

	var req UpdateStoreStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.service.UpdateStoreStatus(ctx, id, req.Status)
	if err != nil {
		switch err {
		case ErrStoreNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		case ErrInvalidStatus:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update store status"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}
