package toolkit

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var ErrTenantNotFound = errors.New("tenant not found")

type ContextKey string

const (
	TenantIDKey ContextKey = "tenant_id"
)

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

func GetTenantID(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	if !ok {
		return "", ErrTenantNotFound
	}
	return tenantID, nil
}

func GetTenantIDFromGinCtx(ctx *gin.Context) (string, bool) {
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

func ExtractActorFromGinCtx(ctx *gin.Context) map[string]string {
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
