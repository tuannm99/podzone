package outputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant entity.Tenant) (*entity.Tenant, error)
	GetByID(ctx context.Context, tenantID string) (*entity.Tenant, error)
	AttachOrganization(ctx context.Context, tenantID string, orgID string) error
	DetachOrganization(ctx context.Context, tenantID string) error
}

type PlatformMembershipRepository interface {
	Upsert(ctx context.Context, userID uint, roleID uint64, status string) error
	ListByUser(ctx context.Context, userID uint) ([]entity.PlatformMembership, error)
	ListRoleIDsByUser(ctx context.Context, userID uint) ([]uint64, error)
	Delete(ctx context.Context, userID uint, roleID uint64) error
}

type MembershipRepository interface {
	Upsert(ctx context.Context, membership entity.Membership) error
	GetByTenantAndUser(ctx context.Context, tenantID string, userID uint) (*entity.Membership, error)
	ListByTenant(ctx context.Context, tenantID string) ([]entity.Membership, error)
	ListByUser(ctx context.Context, userID uint) ([]entity.Membership, error)
	Delete(ctx context.Context, tenantID string, userID uint) error
}

type InviteRepository interface {
	Create(ctx context.Context, invite entity.TenantInvite) error
	GetByID(ctx context.Context, inviteID string) (*entity.TenantInvite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.TenantInvite, error)
	ListByTenant(ctx context.Context, tenantID string) ([]entity.TenantInvite, error)
	MarkAccepted(ctx context.Context, inviteID string, acceptedByUserID uint, acceptedAt time.Time) error
	MarkRevoked(ctx context.Context, inviteID string, revokedAt time.Time) error
}
