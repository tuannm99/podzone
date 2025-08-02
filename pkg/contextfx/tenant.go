package contextfx

import (
	"context"
	"errors"
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
