package outputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type TenantCommandRepository interface {
	Create(ctx context.Context, tenant entity.Tenant) (*entity.Tenant, error)
	AttachOrganization(ctx context.Context, tenantID string, orgID string) error
	DetachOrganization(ctx context.Context, tenantID string) error
}

type TenantQueryRepository interface {
	GetByID(ctx context.Context, tenantID string) (*entity.Tenant, error)
}

type TenantRepository interface {
	TenantCommandRepository
	TenantQueryRepository
}

type PlatformMembershipCommandRepository interface {
	Upsert(ctx context.Context, userID uint, roleID uint64, status string) error
	Delete(ctx context.Context, userID uint, roleID uint64) error
}

type PlatformMembershipQueryRepository interface {
	ListByUser(ctx context.Context, userID uint) ([]entity.PlatformMembership, error)
	ListPageByUser(
		ctx context.Context,
		userID uint,
		query collection.Query,
	) (collection.Page[entity.PlatformMembership], error)
	ListRoleIDsByUser(ctx context.Context, userID uint) ([]uint64, error)
}

type PlatformMembershipRepository interface {
	PlatformMembershipCommandRepository
	PlatformMembershipQueryRepository
}

type MembershipCommandRepository interface {
	Upsert(ctx context.Context, membership entity.Membership) error
	Delete(ctx context.Context, tenantID string, userID uint) error
}

type MembershipQueryRepository interface {
	GetByTenantAndUser(ctx context.Context, tenantID string, userID uint) (*entity.Membership, error)
	ListPageByTenant(
		ctx context.Context,
		tenantID string,
		query collection.Query,
	) (collection.Page[entity.Membership], error)
	ListByUser(ctx context.Context, userID uint) ([]entity.Membership, error)
}

type MembershipRepository interface {
	MembershipCommandRepository
	MembershipQueryRepository
}

type InviteCommandRepository interface {
	Create(ctx context.Context, invite entity.TenantInvite) error
	MarkAccepted(ctx context.Context, inviteID string, acceptedByUserID uint, acceptedAt time.Time) error
	MarkRevoked(ctx context.Context, inviteID string, revokedAt time.Time) error
}

type InviteQueryRepository interface {
	GetByID(ctx context.Context, inviteID string) (*entity.TenantInvite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.TenantInvite, error)
	ListPageByTenant(
		ctx context.Context,
		tenantID string,
		query collection.Query,
	) (collection.Page[entity.TenantInvite], error)
}

type InviteRepository interface {
	InviteCommandRepository
	InviteQueryRepository
}
