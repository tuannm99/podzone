package domain

import "context"

type TenantRepository interface {
	Create(ctx context.Context, tenant Tenant) (*Tenant, error)
	GetByID(ctx context.Context, tenantID string) (*Tenant, error)
}

type RoleRepository interface {
	GetByName(ctx context.Context, name string) (*Role, error)
	RoleHasPermission(ctx context.Context, roleID uint64, permission string) (bool, error)
}

type MembershipRepository interface {
	Upsert(ctx context.Context, membership Membership) error
	GetByTenantAndUser(ctx context.Context, tenantID string, userID uint) (*Membership, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Membership, error)
	ListByUser(ctx context.Context, userID uint) ([]Membership, error)
	Delete(ctx context.Context, tenantID string, userID uint) error
}
