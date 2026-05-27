package scope

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrTenantContextNotFound = errors.New("backoffice: tenant context not found")
	ErrStoreContextNotFound  = errors.New("backoffice: store context not found")
)

type (
	tenantContextKey struct{}
	storeContextKey  struct{}
)

type TenantContext struct {
	TenantID      string
	UserID        string
	SessionID     string
	ClusterName   string
	DBName        string
	SchemaName    string
	PlacementMode string
}

type StoreContext struct {
	StoreID string
}

func WithTenantContext(ctx context.Context, tenant TenantContext) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenant)
}

func GetTenantContext(ctx context.Context) (TenantContext, error) {
	tenant, ok := ctx.Value(tenantContextKey{}).(TenantContext)
	if !ok || strings.TrimSpace(tenant.TenantID) == "" {
		return TenantContext{}, ErrTenantContextNotFound
	}
	return tenant, nil
}

func WithStoreContext(ctx context.Context, store StoreContext) context.Context {
	return context.WithValue(ctx, storeContextKey{}, store)
}

func GetStoreContext(ctx context.Context) (StoreContext, error) {
	store, ok := ctx.Value(storeContextKey{}).(StoreContext)
	if !ok || strings.TrimSpace(store.StoreID) == "" {
		return StoreContext{}, ErrStoreContextNotFound
	}
	return store, nil
}

func CurrentStoreID(ctx context.Context) string {
	store, err := GetStoreContext(ctx)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(store.StoreID)
}
