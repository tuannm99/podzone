package httphandler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/runtime/identity"
	"github.com/tuannm99/podzone/pkg/pdauthn"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type Authentication struct {
	verifier     *pdauthn.Verifier
	serviceToken string
}

func NewAuthentication(cfg onboardingconfig.AuthConfig) *Authentication {
	return &Authentication{
		verifier:     pdauthn.NewVerifier(cfg.Authn),
		serviceToken: strings.TrimSpace(cfg.ServiceToken),
	}
}

func (a *Authentication) RequireUser(ctx *gin.Context) {
	if a.authenticateService(ctx) {
		ctx.Next()
		return
	}

	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization bearer token is required"})
		return
	}

	claims, err := a.verifier.ClaimsFromTokenString(strings.TrimSpace(header[len("Bearer "):]))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if claims.UserID == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization token missing user_id"})
		return
	}

	requestCtx := toolkit.WithUserID(ctx.Request.Context(), strconv.FormatUint(uint64(claims.UserID), 10))
	tenantID := strings.TrimSpace(claims.ActiveTenantID)
	requestedTenantID := strings.TrimSpace(ctx.GetHeader("X-Tenant-ID"))
	if tenantID != "" && requestedTenantID != "" && tenantID != requestedTenantID {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tenant header does not match active session"})
		return
	}
	if tenantID == "" {
		tenantID = requestedTenantID
	}
	if tenantID != "" {
		requestCtx = toolkit.WithTenantID(requestCtx, tenantID)
	}
	requestCtx = metadata.AppendToOutgoingContext(requestCtx, "authorization", header)
	ctx.Request = ctx.Request.WithContext(requestCtx)
	ctx.Next()
}

func (a *Authentication) authenticateService(ctx *gin.Context) bool {
	if a.serviceToken == "" || ctx.GetHeader("X-Onboarding-Service-Token") != a.serviceToken {
		return false
	}
	userID := strings.TrimSpace(ctx.GetHeader("X-User-ID"))
	tenantID := strings.TrimSpace(ctx.GetHeader("X-Tenant-ID"))
	if userID == "" || tenantID == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "service identity headers are required"})
		return true
	}
	requestCtx := toolkit.WithUserID(ctx.Request.Context(), userID)
	requestCtx = toolkit.WithTenantID(requestCtx, tenantID)
	requestCtx = identity.WithTrustedService(requestCtx)
	ctx.Request = ctx.Request.WithContext(requestCtx)
	return true
}
