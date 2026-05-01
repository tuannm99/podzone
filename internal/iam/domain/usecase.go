package domain

import "context"

type IAMUsecase interface {
	CreateTenant(ctx context.Context, ownerUserID uint, cmd CreateTenantCmd) (*Tenant, error)
	AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error
	GetMembership(ctx context.Context, tenantID string, userID uint) (*Membership, error)
	ListUserTenants(ctx context.Context, userID uint) ([]Membership, error)
	ListTenantMembers(ctx context.Context, tenantID string) ([]Membership, error)
	RemoveMember(ctx context.Context, tenantID string, userID uint) error
	CheckPermission(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
	RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error
}
