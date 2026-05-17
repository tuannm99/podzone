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

type PolicyRepository interface {
	CreatePolicy(ctx context.Context, policy Policy, statements []PolicyStatement) (*Policy, []PolicyStatement, error)
	GetPolicyByName(ctx context.Context, name string) (*Policy, error)
	GetPolicyStatements(ctx context.Context, policyID uint64) ([]PolicyStatement, error)
	ListPolicies(ctx context.Context, scope string) ([]Policy, error)
	DeletePolicy(ctx context.Context, policyID uint64) error
	ListRoleStatements(ctx context.Context, roleID uint64) ([]PolicyStatement, error)
	ListPlatformUserStatements(ctx context.Context, userID uint) ([]PolicyStatement, error)
	ListTenantUserStatements(ctx context.Context, tenantID string, userID uint) ([]PolicyStatement, error)
	ListPlatformGroupStatements(ctx context.Context, userID uint) ([]PolicyStatement, error)
	ListTenantGroupStatements(ctx context.Context, tenantID string, userID uint) ([]PolicyStatement, error)
	AttachPlatformUserPolicy(ctx context.Context, userID uint, policyID uint64) error
	DetachPlatformUserPolicy(ctx context.Context, userID uint, policyID uint64) error
	ListPlatformUserPolicies(ctx context.Context, userID uint) ([]Policy, error)
	AttachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyID uint64) error
	DetachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyID uint64) error
	ListTenantUserPolicies(ctx context.Context, tenantID string, userID uint) ([]Policy, error)
}

type GroupRepository interface {
	CreateGroup(ctx context.Context, group Group) (*Group, error)
	ListGroups(ctx context.Context, scope string, tenantID string) ([]Group, error)
	DeleteGroup(ctx context.Context, groupID uint64) error
	AddMember(ctx context.Context, groupID uint64, userID uint) error
	RemoveMember(ctx context.Context, groupID uint64, userID uint) error
	ListMembers(ctx context.Context, groupID uint64) ([]uint, error)
	AttachPolicy(ctx context.Context, groupID uint64, policyID uint64) error
	DetachPolicy(ctx context.Context, groupID uint64, policyID uint64) error
	ListPolicies(ctx context.Context, groupID uint64) ([]Policy, error)
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
