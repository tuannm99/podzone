package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type PolicyUsecase interface {
	CreatePolicy(ctx context.Context, input entity.CreatePolicyInput) (*entity.Policy, []entity.PolicyStatement, error)
	CreatePolicyVersion(ctx context.Context, input entity.CreatePolicyVersionInput) (*entity.PolicyVersion, []entity.PolicyStatement, error)
	DeletePolicyVersion(ctx context.Context, name string, version string) error
	GetPolicy(ctx context.Context, name string) (*entity.Policy, []entity.PolicyStatement, error)
	ListPolicyVersions(ctx context.Context, name string) ([]entity.PolicyVersion, error)
	SetDefaultPolicyVersion(ctx context.Context, name string, version string) error
	ListPolicies(ctx context.Context, scope string) ([]entity.Policy, error)
	ListPolicyAttachments(ctx context.Context, name string) ([]entity.PolicyAttachment, error)
	DeletePolicy(ctx context.Context, name string) error
	PutRoleTrustPolicy(ctx context.Context, input entity.PutRoleTrustPolicyInput) error
	GetRoleTrustPolicy(ctx context.Context, roleName string) ([]entity.RoleTrustStatement, error)
	DeleteRoleTrustPolicy(ctx context.Context, roleName string) error
	PutRolePermissionBoundary(ctx context.Context, roleName string, policyName string) error
	GetRolePermissionBoundary(ctx context.Context, roleName string) (*entity.RolePermissionBoundary, error)
	DeleteRolePermissionBoundary(ctx context.Context, roleName string) error
}
