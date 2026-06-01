package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type TenantUsecase interface {
	CreateTenant(ctx context.Context, ownerUserID uint, cmd entity.CreateTenantCmd) (*entity.Tenant, error)
	AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error
	PutTenantUserInlinePolicy(ctx context.Context, input entity.PutTenantUserInlinePolicyInput) error
	GetTenantUserInlinePolicy(
		ctx context.Context,
		tenantID string,
		userID uint,
		name string,
	) (*entity.UserInlinePolicy, error)
	ListTenantUserInlinePolicies(ctx context.Context, tenantID string, userID uint) ([]entity.UserInlinePolicy, error)
	DeleteTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) error
	AttachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyName string) error
	DetachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyName string) error
	ListTenantUserPolicies(ctx context.Context, tenantID string, userID uint) ([]entity.Policy, error)
	PutTenantUserPermissionBoundary(ctx context.Context, tenantID string, userID uint, policyName string) error
	GetTenantUserPermissionBoundary(ctx context.Context, tenantID string, userID uint) (*entity.PermissionBoundary, error)
	DeleteTenantUserPermissionBoundary(ctx context.Context, tenantID string, userID uint) error
	CreateInvite(
		ctx context.Context,
		tenantID, email, roleName string,
		invitedByUserID uint,
	) (*entity.TenantInvite, string, error)
	GetInvite(ctx context.Context, inviteID string) (*entity.TenantInvite, error)
	ListTenantInvites(ctx context.Context, tenantID string) ([]entity.TenantInvite, error)
	RevokeInvite(ctx context.Context, inviteID string) error
	AcceptInvite(ctx context.Context, inviteToken string, userID uint, email string) (*entity.Membership, error)
	GetMembership(ctx context.Context, tenantID string, userID uint) (*entity.Membership, error)
	ListUserTenants(ctx context.Context, userID uint) ([]entity.Membership, error)
	ListTenantMembers(ctx context.Context, tenantID string) ([]entity.Membership, error)
	RemoveMember(ctx context.Context, tenantID string, userID uint) error
}
