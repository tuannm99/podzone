package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

// IAMCommandUsecase owns state-changing IAM operations.
type IAMCommandUsecase interface {
	TenantCommandUsecase
	OrganizationCommandUsecase
	PolicyCommandUsecase
	GroupCommandUsecase
	PlatformPrincipalCommandUsecase
	AuthzCommandUsecase
}

// IAMQueryUsecase owns read/evaluation IAM operations.
type IAMQueryUsecase interface {
	TenantQueryUsecase
	OrganizationQueryUsecase
	PolicyQueryUsecase
	GroupQueryUsecase
	PlatformPrincipalQueryUsecase
	AuthzQueryUsecase
}

type TenantCommandUsecase interface {
	CreateTenant(ctx context.Context, ownerUserID uint, cmd entity.CreateTenantCmd) (*entity.Tenant, error)
	AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error
	PutTenantUserInlinePolicy(ctx context.Context, input entity.PutTenantUserInlinePolicyInput) error
	DeleteTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) error
	AttachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyName string) error
	DetachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyName string) error
	PutTenantUserPermissionBoundary(ctx context.Context, tenantID string, userID uint, policyName string) error
	DeleteTenantUserPermissionBoundary(ctx context.Context, tenantID string, userID uint) error
	CreateInvite(
		ctx context.Context,
		tenantID, email, roleName string,
		invitedByUserID uint,
	) (*entity.TenantInvite, string, error)
	RevokeInvite(ctx context.Context, inviteID string) error
	AcceptInvite(ctx context.Context, inviteToken string, userID uint, email string) (*entity.Membership, error)
	RemoveMember(ctx context.Context, tenantID string, userID uint) error
}

type TenantQueryUsecase interface {
	GetTenantUserInlinePolicy(
		ctx context.Context,
		tenantID string,
		userID uint,
		name string,
	) (*entity.UserInlinePolicy, error)
	ListTenantUserInlinePolicies(
		ctx context.Context,
		tenantID string,
		userID uint,
		query collection.Query,
	) (collection.Page[entity.UserInlinePolicy], error)
	ListTenantUserPolicies(
		ctx context.Context,
		tenantID string,
		userID uint,
		query collection.Query,
	) (collection.Page[entity.Policy], error)
	GetTenantUserPermissionBoundary(
		ctx context.Context,
		tenantID string,
		userID uint,
	) (*entity.PermissionBoundary, error)
	GetInvite(ctx context.Context, inviteID string) (*entity.TenantInvite, error)
	ListTenantInvites(
		ctx context.Context,
		tenantID string,
		query collection.Query,
	) (collection.Page[entity.TenantInvite], error)
	GetMembership(ctx context.Context, tenantID string, userID uint) (*entity.Membership, error)
	ListUserTenants(ctx context.Context, userID uint) ([]entity.Membership, error)
	ListTenantMembers(
		ctx context.Context,
		tenantID string,
		query collection.Query,
	) (collection.Page[entity.Membership], error)
}

type OrganizationCommandUsecase interface {
	CreateOrganization(ctx context.Context, name string, slug string) (*entity.Organization, error)
	AttachTenantToOrganization(ctx context.Context, tenantID string, orgID string) error
	DetachTenantFromOrganization(ctx context.Context, tenantID string) error
	AttachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error
	DetachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error
}

type OrganizationQueryUsecase interface {
	ListOrganizations(ctx context.Context, query collection.Query) (collection.Page[entity.Organization], error)
	ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error)
}

type PolicyCommandUsecase interface {
	CreatePolicy(ctx context.Context, input entity.CreatePolicyInput) (*entity.Policy, []entity.PolicyStatement, error)
	CreatePolicyVersion(
		ctx context.Context,
		input entity.CreatePolicyVersionInput,
	) (*entity.PolicyVersion, []entity.PolicyStatement, error)
	DeletePolicyVersion(ctx context.Context, name string, version string) error
	SetDefaultPolicyVersion(ctx context.Context, name string, version string) error
	DeletePolicy(ctx context.Context, name string) error
	PutRoleTrustPolicy(ctx context.Context, input entity.PutRoleTrustPolicyInput) error
	DeleteRoleTrustPolicy(ctx context.Context, roleName string) error
	PutRolePermissionBoundary(ctx context.Context, roleName string, policyName string) error
	DeleteRolePermissionBoundary(ctx context.Context, roleName string) error
}

type PolicyQueryUsecase interface {
	GetPolicy(ctx context.Context, name string) (*entity.Policy, []entity.PolicyStatement, error)
	ListPolicyVersions(
		ctx context.Context,
		name string,
		query collection.Query,
	) (collection.Page[entity.PolicyVersion], error)
	ListPolicies(ctx context.Context, scope string, query collection.Query) (collection.Page[entity.Policy], error)
	ListPolicyAttachments(
		ctx context.Context,
		name string,
		query collection.Query,
	) (collection.Page[entity.PolicyAttachment], error)
	GetRoleTrustPolicy(ctx context.Context, roleName string) ([]entity.RoleTrustStatement, error)
	GetRolePermissionBoundary(ctx context.Context, roleName string) (*entity.RolePermissionBoundary, error)
}

type GroupCommandUsecase interface {
	CreateGroup(ctx context.Context, input entity.CreateGroupInput) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupID uint64) error
	PutGroupInlinePolicy(ctx context.Context, input entity.PutGroupInlinePolicyInput) error
	DeleteGroupInlinePolicy(ctx context.Context, groupID uint64, name string) error
	AddGroupMember(ctx context.Context, groupID uint64, userID uint) error
	RemoveGroupMember(ctx context.Context, groupID uint64, userID uint) error
	AttachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error
	DetachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error
}

type GroupQueryUsecase interface {
	ListGroups(
		ctx context.Context,
		scope string,
		tenantID string,
		query collection.Query,
	) (collection.Page[entity.Group], error)
	GetGroupInlinePolicy(ctx context.Context, groupID uint64, name string) (*entity.GroupInlinePolicy, error)
	ListGroupInlinePolicies(
		ctx context.Context,
		groupID uint64,
		query collection.Query,
	) (collection.Page[entity.GroupInlinePolicy], error)
	ListGroupMembers(ctx context.Context, groupID uint64, query collection.Query) (collection.Page[uint], error)
	ListGroupPolicies(
		ctx context.Context,
		groupID uint64,
		query collection.Query,
	) (collection.Page[entity.Policy], error)
}

type PlatformPrincipalCommandUsecase interface {
	AddPlatformRole(ctx context.Context, userID uint, roleName string) error
	RemovePlatformRole(ctx context.Context, userID uint, roleName string) error
	PutPlatformUserInlinePolicy(ctx context.Context, input entity.PutPlatformUserInlinePolicyInput) error
	DeletePlatformUserInlinePolicy(ctx context.Context, userID uint, name string) error
	AttachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error
	DetachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error
	PutPlatformUserPermissionBoundary(ctx context.Context, userID uint, policyName string) error
	DeletePlatformUserPermissionBoundary(ctx context.Context, userID uint) error
}

type PlatformPrincipalQueryUsecase interface {
	CheckPlatformPermission(ctx context.Context, userID uint, permission string) (bool, error)
	RequirePlatformPermission(ctx context.Context, userID uint, permission string) error
	ListPlatformRoles(
		ctx context.Context,
		userID uint,
		query collection.Query,
	) (collection.Page[entity.PlatformMembership], error)
	GetPlatformUserInlinePolicy(ctx context.Context, userID uint, name string) (*entity.UserInlinePolicy, error)
	ListPlatformUserInlinePolicies(
		ctx context.Context,
		userID uint,
		query collection.Query,
	) (collection.Page[entity.UserInlinePolicy], error)
	ListPlatformUserPolicies(
		ctx context.Context,
		userID uint,
		query collection.Query,
	) (collection.Page[entity.Policy], error)
	GetPlatformUserPermissionBoundary(ctx context.Context, userID uint) (*entity.PermissionBoundary, error)
}

type AuthzCommandUsecase interface {
	AssumeRole(ctx context.Context, input entity.AssumeRoleInput) (*entity.AssumedRole, error)
}

type AuthzQueryUsecase interface {
	CheckPermission(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
	CheckPermissionForResource(
		ctx context.Context,
		tenantID string,
		userID uint,
		permission string,
		resource string,
	) (bool, error)
	RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error
	SimulateAccess(ctx context.Context, input entity.SimulateAccessInput) (*entity.SimulateAccessResult, error)
}
