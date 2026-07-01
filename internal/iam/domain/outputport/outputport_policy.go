package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type RoleCommandRepository interface {
	PutTrustPolicy(ctx context.Context, roleID uint64, statements []entity.RoleTrustStatement) error
	DeleteTrustPolicy(ctx context.Context, roleID uint64) error
	PutPermissionBoundary(ctx context.Context, roleID uint64, policyID uint64) error
	DeletePermissionBoundary(ctx context.Context, roleID uint64) error
}

type RoleQueryRepository interface {
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	RoleHasPermission(ctx context.Context, roleID uint64, permission string) (bool, error)
	GetTrustPolicy(ctx context.Context, roleID uint64) ([]entity.RoleTrustStatement, error)
	GetPermissionBoundary(ctx context.Context, roleID uint64) (*entity.RolePermissionBoundary, error)
	GetPermissionBoundaryStatements(ctx context.Context, roleID uint64) ([]entity.PolicyStatement, error)
}

type RoleRepository interface {
	RoleCommandRepository
	RoleQueryRepository
}

type PolicyCommandRepository interface {
	CreatePolicy(
		ctx context.Context,
		policy entity.Policy,
		statements []entity.PolicyStatement,
	) (*entity.Policy, []entity.PolicyStatement, error)
	CreatePolicyVersion(
		ctx context.Context,
		policyID uint64,
		policyName string,
		statements []entity.PolicyStatement,
		setAsDefault bool,
	) (*entity.PolicyVersion, []entity.PolicyStatement, error)
	DeletePolicyVersion(ctx context.Context, policyID uint64, version string) error
	SetDefaultPolicyVersion(ctx context.Context, policyID uint64, version string) error
	DeletePolicy(ctx context.Context, policyID uint64) error
	AttachPlatformUserPolicy(ctx context.Context, userID uint, policyID uint64) error
	DetachPlatformUserPolicy(ctx context.Context, userID uint, policyID uint64) error
	PutPlatformUserInlinePolicy(ctx context.Context, input entity.PutPlatformUserInlinePolicyInput) error
	DeletePlatformUserInlinePolicy(ctx context.Context, userID uint, name string) error
	PutPlatformUserPermissionBoundary(ctx context.Context, userID uint, policyID uint64) error
	DeletePlatformUserPermissionBoundary(ctx context.Context, userID uint) error
	AttachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyID uint64) error
	DetachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyID uint64) error
	PutTenantUserInlinePolicy(ctx context.Context, input entity.PutTenantUserInlinePolicyInput) error
	DeleteTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) error
	PutTenantUserPermissionBoundary(ctx context.Context, tenantID string, userID uint, policyID uint64) error
	DeleteTenantUserPermissionBoundary(ctx context.Context, tenantID string, userID uint) error
}

type PolicyQueryRepository interface {
	GetPolicyByName(ctx context.Context, name string) (*entity.Policy, error)
	GetPolicyStatements(ctx context.Context, policyID uint64) ([]entity.PolicyStatement, error)
	ListPolicyVersions(ctx context.Context, policyID uint64, policyName string) ([]entity.PolicyVersion, error)
	ListPolicies(ctx context.Context, scope string, query collection.Query) (collection.Page[entity.Policy], error)
	ListPolicyAttachments(ctx context.Context, policyID uint64) ([]entity.PolicyAttachment, error)
	ListRoleStatements(ctx context.Context, roleID uint64) ([]entity.PolicyStatement, error)
	ListPlatformUserStatements(ctx context.Context, userID uint) ([]entity.PolicyStatement, error)
	ListTenantUserStatements(ctx context.Context, tenantID string, userID uint) ([]entity.PolicyStatement, error)
	ListPlatformGroupStatements(ctx context.Context, userID uint) ([]entity.PolicyStatement, error)
	ListTenantGroupStatements(ctx context.Context, tenantID string, userID uint) ([]entity.PolicyStatement, error)
	ListPlatformUserPolicies(ctx context.Context, userID uint) ([]entity.Policy, error)
	GetPlatformUserInlinePolicy(ctx context.Context, userID uint, name string) (*entity.UserInlinePolicy, error)
	ListPlatformUserInlinePolicies(ctx context.Context, userID uint) ([]entity.UserInlinePolicy, error)
	GetPlatformUserPermissionBoundary(ctx context.Context, userID uint) (*entity.PermissionBoundary, error)
	GetPlatformUserPermissionBoundaryStatements(ctx context.Context, userID uint) ([]entity.PolicyStatement, error)
	ListTenantUserPolicies(ctx context.Context, tenantID string, userID uint) ([]entity.Policy, error)
	GetTenantUserInlinePolicy(
		ctx context.Context,
		tenantID string,
		userID uint,
		name string,
	) (*entity.UserInlinePolicy, error)
	ListTenantUserInlinePolicies(ctx context.Context, tenantID string, userID uint) ([]entity.UserInlinePolicy, error)
	GetTenantUserPermissionBoundary(
		ctx context.Context,
		tenantID string,
		userID uint,
	) (*entity.PermissionBoundary, error)
	GetTenantUserPermissionBoundaryStatements(
		ctx context.Context,
		tenantID string,
		userID uint,
	) ([]entity.PolicyStatement, error)
}

type PolicyRepository interface {
	PolicyCommandRepository
	PolicyQueryRepository
}
