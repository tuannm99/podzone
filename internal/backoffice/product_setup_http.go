package backoffice

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/pkg/pdhttp"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
)

type productSetupHandler struct {
	cfg        boconfig.Config
	authorizer TenantAuthorizer
	usecase    inputport.ProductSetupUsecase
}

type productSetupRouteParams struct {
	fx.In
	Cfg        boconfig.Config
	Authorizer TenantAuthorizer
	Usecase    inputport.ProductSetupUsecase
}

type productSetupJWTClaims struct {
	UserID         uint   `json:"user_id"`
	ActiveTenantID string `json:"active_tenant_id"`
	SessionID      string `json:"session_id"`
	Key            string `json:"key"`
	jwt.StandardClaims
}

type createDraftRequest struct {
	Name        string `json:"name"`
	Partner     string `json:"partner"`
	BaseCost    string `json:"baseCost"`
	RetailPrice string `json:"retailPrice"`
	Status      string `json:"status"`
	Notes       string `json:"notes"`
}

type promoteCandidateRequest struct {
	DraftID            string                              `json:"draftId"`
	Channel            string                              `json:"channel"`
	VariantColor       string                              `json:"variantColor"`
	VariantSize        string                              `json:"variantSize"`
	ArtworkChecklist   entity.ProductSetupArtworkChecklist `json:"artworkChecklist"`
	MerchandisingNotes string                              `json:"merchandisingNotes"`
}

type updateCandidateStatusRequest struct {
	Status string `json:"status"`
}

func NewProductSetupRoutes(p productSetupRouteParams) pdhttp.RouteRegistrar {
	h := &productSetupHandler{
		cfg:        p.Cfg,
		authorizer: p.Authorizer,
		usecase:    p.Usecase,
	}
	return h.registerRoutes
}

func (h *productSetupHandler) registerRoutes(r *gin.Engine) {
	group := r.Group("/backoffice/v1/product-setup")
	group.GET("", h.withTenantAccess("store_config:read", h.getSnapshot))
	group.POST("/drafts", h.withTenantAccess("store_config:update", h.createDraft))
	group.POST("/candidates/promote", h.withTenantAccess("store_config:update", h.promoteCandidate))
	group.PATCH("/candidates/:id/status", h.withTenantAccess("store_config:update", h.updateCandidateStatus))
}

func (h *productSetupHandler) withTenantAccess(
	permission string,
	handler func(*gin.Context, context.Context),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, err := h.authorizedContext(c, permission)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return
		}
		handler(c, ctx)
	}
}

func (h *productSetupHandler) authorizedContext(c *gin.Context, permission string) (context.Context, error) {
	userID, tenantID, sessionID, err := h.identityFromAuthorization(c.GetHeader("Authorization"))
	if err != nil {
		return nil, err
	}
	ctx := toolkit.WithTenantID(c.Request.Context(), tenantID)
	ctx = toolkit.WithUserID(ctx, userID)
	if err := h.authorizer.AuthorizeTenant(ctx, sessionID, userID, tenantID); err != nil {
		return nil, err
	}
	if err := h.authorizer.RequirePermission(ctx, userID, tenantID, permission); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (h *productSetupHandler) identityFromAuthorization(header string) (string, string, string, error) {
	if header == "" {
		return "", "", "", fmt.Errorf("authorization bearer token is required")
	}
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", "", "", fmt.Errorf("authorization header must use bearer token")
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	if tokenStr == "" {
		return "", "", "", fmt.Errorf("authorization bearer token is required")
	}

	claims := &productSetupJWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(h.cfg.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", "", fmt.Errorf("invalid authorization token")
	}
	if h.cfg.Auth.JWTKey != "" && claims.Key != h.cfg.Auth.JWTKey {
		return "", "", "", fmt.Errorf("invalid authorization token")
	}
	if claims.UserID == 0 {
		return "", "", "", fmt.Errorf("authorization token missing user_id")
	}
	if strings.TrimSpace(claims.ActiveTenantID) == "" {
		return "", "", "", fmt.Errorf("authorization token missing active_tenant_id")
	}
	if claims.SessionID == "" {
		return "", "", "", fmt.Errorf("authorization token missing session_id")
	}
	return strconv.FormatUint(uint64(claims.UserID), 10), claims.ActiveTenantID, claims.SessionID, nil
}

func (h *productSetupHandler) getSnapshot(c *gin.Context, ctx context.Context) {
	snapshot, err := h.usecase.GetSnapshot(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

func (h *productSetupHandler) createDraft(c *gin.Context, ctx context.Context) {
	var req createDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	draft, err := h.usecase.CreateDraft(ctx, inputport.CreateProductSetupDraftCmd{
		Name:        req.Name,
		Partner:     req.Partner,
		BaseCost:    req.BaseCost,
		RetailPrice: req.RetailPrice,
		Status:      req.Status,
		Notes:       req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, draft)
}

func (h *productSetupHandler) promoteCandidate(c *gin.Context, ctx context.Context) {
	var req promoteCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	candidate, err := h.usecase.PromoteCandidate(ctx, inputport.PromoteProductSetupCandidateCmd{
		DraftID:            req.DraftID,
		Channel:            req.Channel,
		VariantColor:       req.VariantColor,
		VariantSize:        req.VariantSize,
		ArtworkChecklist:   req.ArtworkChecklist,
		MerchandisingNotes: req.MerchandisingNotes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, candidate)
}

func (h *productSetupHandler) updateCandidateStatus(c *gin.Context, ctx context.Context) {
	var req updateCandidateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	candidate, err := h.usecase.UpdateCandidateStatus(ctx, c.Param("id"), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}
