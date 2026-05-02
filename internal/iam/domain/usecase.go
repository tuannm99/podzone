package domain

import "context"

type IAMUsecase interface {
	CreateTenant(ctx context.Context, ownerUserID uint, cmd CreateTenantCmd) (*Tenant, error)
	CheckPlatformPermission(ctx context.Context, userID uint, permission string) (bool, error)
	RequirePlatformPermission(ctx context.Context, userID uint, permission string) error
	AddPlatformRole(ctx context.Context, userID uint, roleName string) error
	ListPlatformRoles(ctx context.Context, userID uint) ([]PlatformMembership, error)
	RemovePlatformRole(ctx context.Context, userID uint, roleName string) error
	AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error
	CreateInvite(ctx context.Context, tenantID, email, roleName string, invitedByUserID uint) (*TenantInvite, string, error)
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
