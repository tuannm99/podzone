package outputport

import "context"

type TenantMembershipProjection struct {
	TenantID string `db:"tenant_id"`
	UserID   uint   `db:"user_id"`
	RoleName string `db:"role_name"`
	Status   string `db:"status"`
}

type IAMProjectionRepository interface {
	UpsertTenant(ctx context.Context, tenantID string, slug string, name string) error
	UpsertTenantMembership(ctx context.Context, tenantID string, userID uint, roleName string, status string) error
	GetTenantMembership(ctx context.Context, tenantID string, userID uint) (*TenantMembershipProjection, error)
}
