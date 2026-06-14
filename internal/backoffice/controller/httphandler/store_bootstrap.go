package httphandler

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type StoreBootstrapHandler struct {
	token  string
	stores storectx.StoreUsecase
}

func NewStoreBootstrapHandler(cfg boconfig.Config, stores storectx.StoreUsecase) *StoreBootstrapHandler {
	return &StoreBootstrapHandler{
		token:  strings.TrimSpace(cfg.InternalServiceToken),
		stores: stores,
	}
}

func (h *StoreBootstrapHandler) RegisterRoutes() pdhttp.RouteRegistrar {
	return func(router *gin.Engine) {
		router.POST("/internal/backoffice/v1/stores:bootstrap", h.BootstrapStore)
	}
}

func (h *StoreBootstrapHandler) BootstrapStore(ctx *gin.Context) {
	providedToken := ctx.GetHeader("X-Backoffice-Service-Token")
	if h.token == "" || subtle.ConstantTimeCompare([]byte(providedToken), []byte(h.token)) != 1 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid service token"})
		return
	}

	var request BootstrapStoreRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestCtx := toolkit.WithTenantID(ctx.Request.Context(), strings.TrimSpace(request.WorkspaceID))
	requestCtx = toolkit.WithUserID(requestCtx, strings.TrimSpace(request.OwnerID))
	store, err := h.stores.BootstrapStore(requestCtx, request.toCommand())
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, newBootstrapStoreResponse(store))
}
