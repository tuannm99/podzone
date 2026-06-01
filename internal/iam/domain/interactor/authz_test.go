package interactor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
)

func TestIAMService_RequirePermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantEditor] = entity.Role{ID: 2, Name: entity.RoleTenantEditor}
	state.rolePermissions[2] = map[string]bool{"store:update": true}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   2,
		RoleName: entity.RoleTenantEditor,
		Status:   entity.MembershipStatusActive,
	}

	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "store:update"))
	require.ErrorIs(
		t,
		svc.RequirePermission(context.Background(), "t1", 9, "tenant:manage_members"),
		entity.ErrPermissionDenied,
	)
}

func TestIAMService_RequirePermission_ExplicitDenyWins(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantEditor] = entity.Role{ID: 2, Name: entity.RoleTenantEditor}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   2,
		RoleName: entity.RoleTenantEditor,
		Status:   entity.MembershipStatusActive,
	}
	state.roleStatements[2] = []entity.PolicyStatement{
		{
			PolicyID:        10,
			PolicyName:      "managed/tenant_editor",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "store:update",
			ResourcePattern: "*",
		},
	}
	state.tenantDirect[membershipKey("t1", 9)] = []entity.PolicyStatement{
		{
			PolicyID:        11,
			PolicyName:      "inline/deny-store-update",
			Effect:          entity.PolicyEffectDeny,
			ActionPattern:   "store:update",
			ResourcePattern: "*",
		},
	}

	require.ErrorIs(
		t,
		svc.RequirePermission(context.Background(), "t1", 9, "store:update"),
		entity.ErrPermissionDenied,
	)
}

func TestIAMService_CheckPermissionForResource(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 3, Name: entity.RoleTenantViewer}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   3,
		RoleName: entity.RoleTenantViewer,
		Status:   entity.MembershipStatusActive,
	}
	state.tenantDirect[membershipKey("t1", 9)] = []entity.PolicyStatement{
		{
			PolicyID:        12,
			PolicyName:      "inline/store-a-read",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "store:read",
			ResourcePattern: entity.ResourceStore("t1", "store-a"),
		},
	}

	allowed, err := svc.CheckPermissionForResource(
		context.Background(),
		"t1",
		9,
		"store:read",
		entity.ResourceStore("t1", "store-a"),
	)
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = svc.CheckPermissionForResource(
		context.Background(),
		"t1",
		9,
		"store:read",
		entity.ResourceStore("t1", "store-b"),
	)
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestIAMService_RequirePlatformPermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.rolePermissions[7] = map[string]bool{"tenant:create": true}
	state.platformRoleIDs[11] = []uint64{7}

	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
	require.ErrorIs(
		t,
		svc.RequirePlatformPermission(context.Background(), 12, "tenant:create"),
		entity.ErrPermissionDenied,
	)
}

func TestIAMService_RequirePlatformPermission_DirectPolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.platformDirect[11] = []entity.PolicyStatement{
		{
			PolicyID:        20,
			PolicyName:      "inline/platform-direct",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "tenant:create",
			ResourcePattern: "*",
		},
	}

	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
}

func TestIAMService_AssumeRole_TenantTrust(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[entity.RoleTenantAdmin] = entity.Role{
		ID:    2,
		Scope: entity.PolicyScopeTenant,
		Name:  entity.RoleTenantAdmin,
	}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   1,
		RoleName: entity.RoleTenantOwner,
		Status:   entity.MembershipStatusActive,
	}
	require.NoError(t, svc.PutRoleTrustPolicy(context.Background(), entity.PutRoleTrustPolicyInput{
		RoleName: entity.RoleTenantAdmin,
		Statements: []entity.RoleTrustStatement{{
			Effect:           entity.PolicyEffectAllow,
			PrincipalType:    entity.TrustPrincipalTenantRole,
			PrincipalPattern: entity.RoleTenantOwner,
			TenantPattern:    "*",
		}},
	}))

	assumedRole, err := svc.AssumeRole(context.Background(), entity.AssumeRoleInput{
		UserID:   9,
		RoleName: entity.RoleTenantAdmin,
		TenantID: "t1",
	})
	require.NoError(t, err)
	require.NotNil(t, assumedRole)
	require.Equal(t, entity.RoleTenantAdmin, assumedRole.RoleName)
	require.Equal(t, entity.PolicyScopeTenant, assumedRole.RoleScope)
	require.Equal(t, "t1", assumedRole.TenantID)
}

func TestIAMService_CheckPermission_AssumedRoleRejectsDifferentTenant(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[entity.RoleTenantAdmin] = entity.Role{
		ID:    2,
		Scope: entity.PolicyScopeTenant,
		Name:  entity.RoleTenantAdmin,
	}
	state.roleStatements[2] = []entity.PolicyStatement{
		{
			PolicyID:        10,
			PolicyName:      "managed/tenant_admin",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	ctx := entity.WithAssumedRole(context.Background(), entity.AssumedRole{
		RoleID:    2,
		RoleScope: entity.PolicyScopeTenant,
		RoleName:  entity.RoleTenantAdmin,
		TenantID:  "t1",
	})
	allowed, err := svc.CheckPermission(ctx, "t2", 9, "order:update")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestIAMService_CheckPermission_AssumedRoleSessionPolicyScopesDown(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[entity.RoleTenantAdmin] = entity.Role{
		ID:    2,
		Scope: entity.PolicyScopeTenant,
		Name:  entity.RoleTenantAdmin,
	}
	state.roleStatements[2] = []entity.PolicyStatement{
		{
			PolicyID:        10,
			PolicyName:      "managed/tenant_admin",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	ctx := entity.WithAssumedRole(context.Background(), entity.AssumedRole{
		RoleID:    2,
		RoleScope: entity.PolicyScopeTenant,
		RoleName:  entity.RoleTenantAdmin,
		TenantID:  "t1",
	})
	ctx = entity.WithSessionPolicyStatements(ctx, []entity.PolicyStatement{
		{Effect: entity.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
	})

	allowed, err := svc.CheckPermission(ctx, "t1", 9, "order:update")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestIAMService_AssumeRole_ExternalIDPattern(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[entity.RoleTenantAdmin] = entity.Role{
		ID:    2,
		Scope: entity.PolicyScopeTenant,
		Name:  entity.RoleTenantAdmin,
	}
	state.roleTrustStatements[2] = []entity.RoleTrustStatement{{
		RoleID:            2,
		Effect:            entity.PolicyEffectAllow,
		PrincipalType:     entity.TrustPrincipalTenantRole,
		PrincipalPattern:  entity.RoleTenantOwner,
		TenantPattern:     "t1",
		ExternalIDPattern: "ext-*",
	}}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1", UserID: 9, RoleID: 1, RoleName: entity.RoleTenantOwner, Status: entity.MembershipStatusActive,
	}

	_, err := svc.AssumeRole(context.Background(), entity.AssumeRoleInput{
		UserID: 9, RoleName: entity.RoleTenantAdmin, TenantID: "t1", ExternalID: "bad",
	})
	require.ErrorIs(t, err, entity.ErrAssumeRoleDenied)

	assumedRole, err := svc.AssumeRole(context.Background(), entity.AssumeRoleInput{
		UserID: 9, RoleName: entity.RoleTenantAdmin, TenantID: "t1", ExternalID: "ext-123",
	})
	require.NoError(t, err)
	require.Equal(t, entity.RoleTenantAdmin, assumedRole.RoleName)
}

func TestIAMService_SimulateAccess_WithConditionAndBoundary(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.tenants["t1"] = entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.roleByName[entity.RoleTenantEditor] = entity.Role{
		ID:    2,
		Scope: entity.PolicyScopeTenant,
		Name:  entity.RoleTenantEditor,
	}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1", UserID: 9, RoleID: 2, RoleName: entity.RoleTenantEditor, Status: entity.MembershipStatusActive,
	}
	state.tenantDirect[membershipKey("t1", 9)] = []entity.PolicyStatement{{
		PolicyName:      "inline/orders",
		Effect:          entity.PolicyEffectAllow,
		ActionPattern:   "order:update",
		ResourcePattern: "*",
		Conditions: []entity.PolicyCondition{{
			Operator: entity.ConditionStringEquals,
			Key:      "region",
			Value:    "us",
		}},
	}}
	state.tenantBoundaryStmts[membershipKey("t1", 9)] = []entity.PolicyStatement{{
		PolicyName:      "boundary/orders",
		Effect:          entity.PolicyEffectAllow,
		ActionPattern:   "order:update",
		ResourcePattern: "*",
	}}

	result, err := svc.SimulateAccess(context.Background(), entity.SimulateAccessInput{
		Scope:      entity.PolicyScopeTenant,
		TenantID:   "t1",
		UserID:     9,
		Action:     "order:update",
		Resource:   "*",
		Attributes: map[string]string{"region": "eu"},
	})
	require.NoError(t, err)
	require.False(t, result.Allowed)

	result, err = svc.SimulateAccess(context.Background(), entity.SimulateAccessInput{
		Scope:      entity.PolicyScopeTenant,
		TenantID:   "t1",
		UserID:     9,
		Action:     "order:update",
		Resource:   "*",
		Attributes: map[string]string{"region": "us"},
	})
	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.NotEmpty(t, result.MatchedStatements)
}
