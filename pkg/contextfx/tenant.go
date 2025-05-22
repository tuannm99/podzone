package contextfx

import (
	"context"
	"errors"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// TenantIDKey is the key for tenant_id in context
	TenantIDKey ContextKey = "tenant_id"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrUnauthorized   = errors.New("unauthorized access")
)

// WithTenantID adds tenant_id to the context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves tenant_id from the context
func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	return tenantID, ok
}
