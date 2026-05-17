package domain

import "context"

type IAMUsecase interface {
	CreateTenant(ctx context.Context, ownerUserID uint, cmd CreateTenantCmd) (*Tenant, error)
	CreatePolicy(ctx context.Context, input CreatePolicyInput) (*Policy, []PolicyStatement, error)
	GetPolicy(ctx context.Context, name string) (*Policy, []PolicyStatement, error)
	ListPolicies(ctx context.Context, scope string) ([]Policy, error)
	DeletePolicy(ctx context.Context, name string) error
	CreateGroup(ctx context.Context, input CreateGroupInput) (*Group, error)
	ListGroups(ctx context.Context, scope string, tenantID string) ([]Group, error)
	DeleteGroup(ctx context.Context, groupID uint64) error
	AddGroupMember(ctx context.Context, groupID uint64, userID uint) error
	RemoveGroupMember(ctx context.Context, groupID uint64, userID uint) error
	ListGroupMembers(ctx context.Context, groupID uint64) ([]uint, error)
	AttachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error
	DetachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error
	ListGroupPolicies(ctx context.Context, groupID uint64) ([]Policy, error)
	CheckPlatformPermission(ctx context.Context, userID uint, permission string) (bool, error)
	RequirePlatformPermission(ctx context.Context, userID uint, permission string) error
	AddPlatformRole(ctx context.Context, userID uint, roleName string) error
	ListPlatformRoles(ctx context.Context, userID uint) ([]PlatformMembership, error)
	RemovePlatformRole(ctx context.Context, userID uint, roleName string) error
	AttachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error
	DetachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error
	ListPlatformUserPolicies(ctx context.Context, userID uint) ([]Policy, error)
	AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error
	AttachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyName string) error
	DetachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyName string) error
	ListTenantUserPolicies(ctx context.Context, tenantID string, userID uint) ([]Policy, error)
	CreateInvite(
		ctx context.Context,
		tenantID, email, roleName string,
		invitedByUserID uint,
	) (*TenantInvite, string, error)
	GetInvite(ctx context.Context, inviteID string) (*TenantInvite, error)
	ListTenantInvites(ctx context.Context, tenantID string) ([]TenantInvite, error)
	RevokeInvite(ctx context.Context, inviteID string) error
	AcceptInvite(ctx context.Context, inviteToken string, userID uint, email string) (*Membership, error)
	GetMembership(ctx context.Context, tenantID string, userID uint) (*Membership, error)
	ListUserTenants(ctx context.Context, userID uint) ([]Membership, error)
	ListTenantMembers(ctx context.Context, tenantID string) ([]Membership, error)
	RemoveMember(ctx context.Context, tenantID string, userID uint) error
	CheckPermission(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
	RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error
}
