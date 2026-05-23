package grpchandler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iamentity "github.com/tuannm99/podzone/internal/iam/entity"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

func TestCreateTenant_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		requirePlatformPermissionFunc: func(ctx context.Context, userID uint, permission string) error { return nil },
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamentity.CreateTenantCmd) (*iamentity.Tenant, error) {
			return &iamentity.Tenant{ID: "tenant-1", Name: cmd.Name, Slug: cmd.Slug}, nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamentity.Membership, error) {
			return &iamentity.Membership{TenantID: tenantID, UserID: userID, RoleName: iamentity.RoleTenantOwner, Status: iamentity.MembershipStatusActive}, nil
		},
	}))

	res, err := srv.CreateTenant(authContextForIAMUser(t, 7), &pbauthv1.CreateTenantRequest{
		OwnerUserId: 7,
		Name:        "Demo Tenant",
		Slug:        "demo-tenant",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-1", res.Tenant.Id)
	assert.Equal(t, uint64(7), res.OwnerMembership.UserId)
}

func TestCheckPermission_InactiveMembershipReturnsNotAllowed(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, iamentity.ErrInactiveMembership
		},
	}))

	res, err := srv.CheckPermission(context.Background(), &pbauthv1.CheckPermissionRequest{
		TenantId:   "tenant-1",
		UserId:     9,
		Permission: "order:update",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Allowed)
}

func TestCreatePolicy_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		requirePlatformPermissionFunc: func(ctx context.Context, userID uint, permission string) error { return nil },
		createPolicyFunc: func(ctx context.Context, input iamentity.CreatePolicyInput) (*iamentity.Policy, []iamentity.PolicyStatement, error) {
			return &iamentity.Policy{Name: input.Name, Scope: input.Scope}, input.Statements, nil
		},
	}))

	res, err := srv.CreatePolicy(authContextForIAMUser(t, 7), &pbauthv1.CreatePolicyRequest{
		Scope: "platform",
		Name:  "managed/test",
		Statements: []*pbauthv1.PolicyStatement{{
			Effect:          "allow",
			ActionPattern:   "tenant:create",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "managed/test", res.Policy.Name)
	assert.Len(t, res.Statements, 1)
}

func TestGetTenantMembership_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamentity.Membership, error) {
			return &iamentity.Membership{TenantID: tenantID, UserID: userID, RoleName: iamentity.RoleTenantAdmin, Status: iamentity.MembershipStatusActive}, nil
		},
	}))

	res, err := srv.GetTenantMembership(context.Background(), &pbauthv1.GetTenantMembershipRequest{
		TenantId: "tenant-1",
		UserId:   9,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-1", res.Membership.TenantId)
	assert.Equal(t, uint64(9), res.Membership.UserId)
}

func TestListUserTenants_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamentity.Membership, error) {
			return []iamentity.Membership{
				{TenantID: "tenant-1", UserID: userID, RoleName: iamentity.RoleTenantAdmin, Status: iamentity.MembershipStatusActive},
			}, nil
		},
	}))

	res, err := srv.ListUserTenants(authContextForIAMUser(t, 7), &pbauthv1.ListUserTenantsRequest{UserId: 7})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Memberships, 1)
	assert.Equal(t, "tenant-1", res.Memberships[0].TenantId)
}

func TestAssumeRoleRPC_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		assumeRoleFunc: func(ctx context.Context, input iamentity.AssumeRoleInput) (*iamentity.AssumedRole, error) {
			return &iamentity.AssumedRole{
				RoleID:    5,
				RoleScope: iamentity.PolicyScopeTenant,
				RoleName:  iamentity.RoleTenantAdmin,
				TenantID:  input.TenantID,
				ExpiresAt: nowPlusHour(),
			}, nil
		},
	}))

	res, err := srv.AssumeRole(context.Background(), &pbauthv1.IAMAssumeRoleRequest{
		AccessToken: rawAccessTokenForIAMUser(t, 7),
		RoleName:    iamentity.RoleTenantAdmin,
		TenantId:    "tenant-1",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(5), res.AssumedRole.RoleId)
	assert.Equal(t, "tenant-1", res.AssumedRole.TenantId)
}
