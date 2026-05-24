package outputport

import "context"

type IAMProjectionRepository interface {
	UpsertTenant(ctx context.Context, tenantID string, slug string, name string) error
	UpsertTenantMembership(ctx context.Context, tenantID string, userID uint, roleName string, status string) error
}
