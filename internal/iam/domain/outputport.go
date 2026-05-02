package domain

import (
	"context"
	"time"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant Tenant) (*Tenant, error)
	GetByID(ctx context.Context, tenantID string) (*Tenant, error)
}

type RoleRepository interface {
	GetByName(ctx context.Context, name string) (*Role, error)
	RoleHasPermission(ctx context.Context, roleID uint64, permission string) (bool, error)
}

type PlatformMembershipRepository interface {
	Upsert(ctx context.Context, userID uint, roleID uint64, status string) error
	ListByUser(ctx context.Context, userID uint) ([]PlatformMembership, error)
	ListRoleIDsByUser(ctx context.Context, userID uint) ([]uint64, error)
	Delete(ctx context.Context, userID uint, roleID uint64) error
}

type MembershipRepository interface {
	Upsert(ctx context.Context, membership Membership) error
	GetByTenantAndUser(ctx context.Context, tenantID string, userID uint) (*Membership, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Membership, error)
	ListByUser(ctx context.Context, userID uint) ([]Membership, error)
	Delete(ctx context.Context, tenantID string, userID uint) error
}

type InviteRepository interface {
	Create(ctx context.Context, invite TenantInvite) error
	GetByID(ctx context.Context, inviteID string) (*TenantInvite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*TenantInvite, error)
	ListByTenant(ctx context.Context, tenantID string) ([]TenantInvite, error)
	MarkAccepted(ctx context.Context, inviteID string, acceptedByUserID uint, acceptedAt time.Time) error
	MarkRevoked(ctx context.Context, inviteID string, revokedAt time.Time) error
}
