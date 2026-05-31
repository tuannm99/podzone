package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type GroupUsecase interface {
	CreateGroup(ctx context.Context, input entity.CreateGroupInput) (*entity.Group, error)
	ListGroups(ctx context.Context, scope string, tenantID string) ([]entity.Group, error)
	DeleteGroup(ctx context.Context, groupID uint64) error
	PutGroupInlinePolicy(ctx context.Context, input entity.PutGroupInlinePolicyInput) error
	GetGroupInlinePolicy(ctx context.Context, groupID uint64, name string) (*entity.GroupInlinePolicy, error)
	ListGroupInlinePolicies(ctx context.Context, groupID uint64) ([]entity.GroupInlinePolicy, error)
	DeleteGroupInlinePolicy(ctx context.Context, groupID uint64, name string) error
	AddGroupMember(ctx context.Context, groupID uint64, userID uint) error
	RemoveGroupMember(ctx context.Context, groupID uint64, userID uint) error
	ListGroupMembers(ctx context.Context, groupID uint64) ([]uint, error)
	AttachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error
	DetachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error
	ListGroupPolicies(ctx context.Context, groupID uint64) ([]entity.Policy, error)
}

type PlatformPrincipalUsecase interface {
	CheckPlatformPermission(ctx context.Context, userID uint, permission string) (bool, error)
	RequirePlatformPermission(ctx context.Context, userID uint, permission string) error
	AddPlatformRole(ctx context.Context, userID uint, roleName string) error
	ListPlatformRoles(ctx context.Context, userID uint) ([]entity.PlatformMembership, error)
	RemovePlatformRole(ctx context.Context, userID uint, roleName string) error
	PutPlatformUserInlinePolicy(ctx context.Context, input entity.PutPlatformUserInlinePolicyInput) error
	GetPlatformUserInlinePolicy(ctx context.Context, userID uint, name string) (*entity.UserInlinePolicy, error)
	ListPlatformUserInlinePolicies(ctx context.Context, userID uint) ([]entity.UserInlinePolicy, error)
	DeletePlatformUserInlinePolicy(ctx context.Context, userID uint, name string) error
	AttachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error
	DetachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error
	ListPlatformUserPolicies(ctx context.Context, userID uint) ([]entity.Policy, error)
	PutPlatformUserPermissionBoundary(ctx context.Context, userID uint, policyName string) error
	GetPlatformUserPermissionBoundary(ctx context.Context, userID uint) (*entity.PermissionBoundary, error)
	DeletePlatformUserPermissionBoundary(ctx context.Context, userID uint) error
}
