package store

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// StoreController handles HTTP requests for store management
type StoreController struct {
	logger  *zap.Logger
	service *StoreService
}

// StoreControllerParams contains dependencies for Controller
type StoreControllerParams struct {
	fx.In

	Logger       *zap.Logger
	StoreService *StoreService
}

// NewStoreController creates a new store controller
func NewStoreController(params StoreControllerParams) *StoreController {
	return &StoreController{
		logger:  params.Logger,
		service: params.StoreService,
	}
}

// RegisterRoutes registers the HTTP routes for store management
func (c *StoreController) RegisterRoutes(r *gin.RouterGroup) {
	stores := r.Group("/stores")
	{
		stores.POST("", c.CreateStore)
		stores.GET("", c.ListStores)
		stores.GET("/:id", c.GetStore)
		stores.PATCH("/:id/status", c.UpdateStoreStatus)
	}
}

// CreateStoreRequest represents the request body for creating a store
type CreateStoreRequest struct {
	Name      string `json:"name"      binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
}

// CreateStore handles store creation
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

// ListStores handles listing stores for the current user
func (c *StoreController) ListStores(ctx *gin.Context) {
	// TODO: Get owner ID from auth context
	ownerID := "temp-owner-id"

	stores, err := c.service.GetStoresByOwner(ctx, ownerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list stores"})
		return
	}

	ctx.JSON(http.StatusOK, stores)
}

// GetStore handles retrieving a single store
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

// UpdateStoreStatusRequest represents the request body for updating store status
type UpdateStoreStatusRequest struct {
	Status StoreStatus `json:"status" binding:"required"`
}

// UpdateStoreStatus handles updating a store's status
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
