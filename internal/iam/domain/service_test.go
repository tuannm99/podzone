package domain_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domain "github.com/tuannm99/podzone/internal/iam/domain"
	domainmocks "github.com/tuannm99/podzone/internal/iam/domain/mocks"
)

func TestIAMService_CreateTenant_AssignsOwnerRole(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[domain.RoleTenantOwner] = domain.Role{ID: 1, Name: domain.RoleTenantOwner}

	out, err := svc.CreateTenant(
		context.Background(),
		42,
		domain.CreateTenantCmd{Name: "Tenant One", Slug: "tenant-one"},
	)
	require.NoError(t, err)
	require.NotNil(t, out)

	gotMembership, err := state.memberships.GetByTenantAndUser(context.Background(), out.ID, 42)
	require.NoError(t, err)
	require.Equal(t, domain.RoleTenantOwner, gotMembership.RoleName)
}

func TestIAMService_CreatePolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)

	policy, statements, err := svc.CreatePolicy(context.Background(), domain.CreatePolicyInput{
		Scope:       domain.PolicyScopeTenant,
		Name:        "tenant/orders_editor",
		Description: "Edit routed orders",
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, policy)
	require.Len(t, statements, 1)
	require.Equal(t, "tenant/orders_editor", policy.Name)
	require.Contains(t, state.policiesByName, policy.Name)
	require.Equal(t, "v1", policy.DefaultVersion)
}

func TestIAMService_CreatePolicyVersionAndSetDefault(t *testing.T) {
	t.Parallel()

	svc, _ := newIAMTestUsecase(t)
	policy, _, err := svc.CreatePolicy(context.Background(), domain.CreatePolicyInput{
		Scope:       domain.PolicyScopeTenant,
		Name:        "tenant/orders_editor",
		Description: "Edit routed orders",
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)

	version, statements, err := svc.CreatePolicyVersion(context.Background(), domain.CreatePolicyVersionInput{
		PolicyName: policy.Name,
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
		},
		SetAsDefault: true,
	})
	require.NoError(t, err)
	require.Equal(t, "v2", version.Version)
	require.True(t, version.IsDefault)
	require.Len(t, statements, 1)

	gotPolicy, gotStatements, err := svc.GetPolicy(context.Background(), policy.Name)
	require.NoError(t, err)
	require.Equal(t, "v2", gotPolicy.DefaultVersion)
	require.Len(t, gotStatements, 1)
	require.Equal(t, "order:read", gotStatements[0].ActionPattern)
}

func TestIAMService_DeleteNonDefaultPolicyVersion(t *testing.T) {
	t.Parallel()

	svc, _ := newIAMTestUsecase(t)
	policy, _, err := svc.CreatePolicy(context.Background(), domain.CreatePolicyInput{
		Scope: domain.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)
	_, _, err = svc.CreatePolicyVersion(context.Background(), domain.CreatePolicyVersionInput{
		PolicyName: policy.Name,
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeletePolicyVersion(context.Background(), policy.Name, "v2"))
	require.ErrorIs(t, svc.DeletePolicyVersion(context.Background(), policy.Name, "v1"), domain.ErrDefaultPolicyVersion)
}

func TestIAMService_RequirePermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantEditor] = domain.Role{ID: 2, Name: domain.RoleTenantEditor}
	state.rolePermissions[2] = map[string]bool{"store:update": true}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   2,
		RoleName: domain.RoleTenantEditor,
		Status:   domain.MembershipStatusActive,
	}

	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "store:update"))
	require.ErrorIs(
		t,
		svc.RequirePermission(context.Background(), "t1", 9, "tenant:manage_members"),
		domain.ErrPermissionDenied,
	)
}

func TestIAMService_RequirePermission_ExplicitDenyWins(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantEditor] = domain.Role{ID: 2, Name: domain.RoleTenantEditor}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   2,
		RoleName: domain.RoleTenantEditor,
		Status:   domain.MembershipStatusActive,
	}
	state.roleStatements[2] = []domain.PolicyStatement{
		{
			PolicyID:        10,
			PolicyName:      "managed/tenant_editor",
			Effect:          domain.PolicyEffectAllow,
			ActionPattern:   "store:update",
			ResourcePattern: "*",
		},
	}
	state.tenantDirect[membershipKey("t1", 9)] = []domain.PolicyStatement{
		{
			PolicyID:        11,
			PolicyName:      "inline/deny-store-update",
			Effect:          domain.PolicyEffectDeny,
			ActionPattern:   "store:update",
			ResourcePattern: "*",
		},
	}

	require.ErrorIs(
		t,
		svc.RequirePermission(context.Background(), "t1", 9, "store:update"),
		domain.ErrPermissionDenied,
	)
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
		domain.ErrPermissionDenied,
	)
}

func TestIAMService_RequirePlatformPermission_DirectPolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.platformDirect[11] = []domain.PolicyStatement{
		{
			PolicyID:        20,
			PolicyName:      "inline/platform-direct",
			Effect:          domain.PolicyEffectAllow,
			ActionPattern:   "tenant:create",
			ResourcePattern: "*",
		},
	}

	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
}

func TestIAMService_AssumeRole_TenantTrust(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[domain.RoleTenantAdmin] = domain.Role{
		ID:    2,
		Scope: domain.PolicyScopeTenant,
		Name:  domain.RoleTenantAdmin,
	}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   1,
		RoleName: domain.RoleTenantOwner,
		Status:   domain.MembershipStatusActive,
	}
	require.NoError(t, svc.PutRoleTrustPolicy(context.Background(), domain.PutRoleTrustPolicyInput{
		RoleName: domain.RoleTenantAdmin,
		Statements: []domain.RoleTrustStatement{{
			Effect:           domain.PolicyEffectAllow,
			PrincipalType:    domain.TrustPrincipalTenantRole,
			PrincipalPattern: domain.RoleTenantOwner,
			TenantPattern:    "*",
		}},
	}))

	assumedRole, err := svc.AssumeRole(context.Background(), domain.AssumeRoleInput{
		UserID:   9,
		RoleName: domain.RoleTenantAdmin,
		TenantID: "t1",
	})
	require.NoError(t, err)
	require.NotNil(t, assumedRole)
	require.Equal(t, domain.RoleTenantAdmin, assumedRole.RoleName)
	require.Equal(t, domain.PolicyScopeTenant, assumedRole.RoleScope)
	require.Equal(t, "t1", assumedRole.TenantID)
}

func TestIAMService_CheckPermission_AssumedRoleRejectsDifferentTenant(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[domain.RoleTenantAdmin] = domain.Role{
		ID:    2,
		Scope: domain.PolicyScopeTenant,
		Name:  domain.RoleTenantAdmin,
	}
	state.roleStatements[2] = []domain.PolicyStatement{
		{
			PolicyID:        10,
			PolicyName:      "managed/tenant_admin",
			Effect:          domain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	ctx := domain.WithAssumedRole(context.Background(), domain.AssumedRole{
		RoleID:    2,
		RoleScope: domain.PolicyScopeTenant,
		RoleName:  domain.RoleTenantAdmin,
		TenantID:  "t1",
	})
	allowed, err := svc.CheckPermission(ctx, "t2", 9, "order:update")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestIAMService_CheckPermission_AssumedRoleSessionPolicyScopesDown(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[domain.RoleTenantAdmin] = domain.Role{
		ID:    2,
		Scope: domain.PolicyScopeTenant,
		Name:  domain.RoleTenantAdmin,
	}
	state.roleStatements[2] = []domain.PolicyStatement{
		{
			PolicyID:        10,
			PolicyName:      "managed/tenant_admin",
			Effect:          domain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	ctx := domain.WithAssumedRole(context.Background(), domain.AssumedRole{
		RoleID:    2,
		RoleScope: domain.PolicyScopeTenant,
		RoleName:  domain.RoleTenantAdmin,
		TenantID:  "t1",
	})
	ctx = domain.WithSessionPolicyStatements(ctx, []domain.PolicyStatement{
		{Effect: domain.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
	})

	allowed, err := svc.CheckPermission(ctx, "t1", 9, "order:update")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestIAMService_AssumeRole_ExternalIDPattern(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[domain.RoleTenantAdmin] = domain.Role{ID: 2, Scope: domain.PolicyScopeTenant, Name: domain.RoleTenantAdmin}
	state.roleTrustStatements[2] = []domain.RoleTrustStatement{{
		RoleID:            2,
		Effect:            domain.PolicyEffectAllow,
		PrincipalType:     domain.TrustPrincipalTenantRole,
		PrincipalPattern:  domain.RoleTenantOwner,
		TenantPattern:     "t1",
		ExternalIDPattern: "ext-*",
	}}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1", UserID: 9, RoleID: 1, RoleName: domain.RoleTenantOwner, Status: domain.MembershipStatusActive,
	}

	_, err := svc.AssumeRole(context.Background(), domain.AssumeRoleInput{
		UserID: 9, RoleName: domain.RoleTenantAdmin, TenantID: "t1", ExternalID: "bad",
	})
	require.ErrorIs(t, err, domain.ErrAssumeRoleDenied)

	assumedRole, err := svc.AssumeRole(context.Background(), domain.AssumeRoleInput{
		UserID: 9, RoleName: domain.RoleTenantAdmin, TenantID: "t1", ExternalID: "ext-123",
	})
	require.NoError(t, err)
	require.Equal(t, domain.RoleTenantAdmin, assumedRole.RoleName)
}

func TestIAMService_SimulateAccess_WithConditionAndBoundary(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.tenants["t1"] = domain.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.roleByName[domain.RoleTenantEditor] = domain.Role{ID: 2, Scope: domain.PolicyScopeTenant, Name: domain.RoleTenantEditor}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1", UserID: 9, RoleID: 2, RoleName: domain.RoleTenantEditor, Status: domain.MembershipStatusActive,
	}
	state.tenantDirect[membershipKey("t1", 9)] = []domain.PolicyStatement{{
		PolicyName:      "inline/orders",
		Effect:          domain.PolicyEffectAllow,
		ActionPattern:   "order:update",
		ResourcePattern: "*",
		Conditions: []domain.PolicyCondition{{
			Operator: domain.ConditionStringEquals,
			Key:      "region",
			Value:    "us",
		}},
	}}
	state.tenantBoundaryStmts[membershipKey("t1", 9)] = []domain.PolicyStatement{{
		PolicyName:      "boundary/orders",
		Effect:          domain.PolicyEffectAllow,
		ActionPattern:   "order:update",
		ResourcePattern: "*",
	}}

	result, err := svc.SimulateAccess(context.Background(), domain.SimulateAccessInput{
		Scope:      domain.PolicyScopeTenant,
		TenantID:   "t1",
		UserID:     9,
		Action:     "order:update",
		Resource:   "*",
		Attributes: map[string]string{"region": "eu"},
	})
	require.NoError(t, err)
	require.False(t, result.Allowed)

	result, err = svc.SimulateAccess(context.Background(), domain.SimulateAccessInput{
		Scope:      domain.PolicyScopeTenant,
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

func TestIAMService_PlatformUserInlinePolicyAffectsPermission(t *testing.T) {
	t.Parallel()

	svc, _ := newIAMTestUsecase(t)
	require.NoError(t, svc.PutPlatformUserInlinePolicy(context.Background(), domain.PutPlatformUserInlinePolicyInput{
		UserID: 11,
		Name:   "inline-platform",
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "tenant:create", ResourcePattern: "*"},
		},
	}))
	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
}

func TestIAMService_PlatformUserPermissionBoundaryScopesDown(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.platformDirect[11] = []domain.PolicyStatement{
		{PolicyID: 20, PolicyName: "inline/platform-direct", Effect: domain.PolicyEffectAllow, ActionPattern: "tenant:create", ResourcePattern: "*"},
	}
	state.policiesByName["managed/platform-readonly-boundary"] = domain.Policy{ID: 21, Scope: domain.PolicyScopePlatform, Name: "managed/platform-readonly-boundary"}
	state.policiesByID[21] = state.policiesByName["managed/platform-readonly-boundary"]
	state.policyStatements[21] = []domain.PolicyStatement{
		{ID: 1, PolicyID: 21, PolicyName: "managed/platform-readonly-boundary", Effect: domain.PolicyEffectAllow, ActionPattern: "tenant:read", ResourcePattern: "*"},
	}

	require.NoError(t, svc.PutPlatformUserPermissionBoundary(context.Background(), 11, "managed/platform-readonly-boundary"))
	require.ErrorIs(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"), domain.ErrPermissionDenied)
}

func TestIAMService_AddAndRemovePlatformRole(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[domain.RolePlatformAdmin] = domain.Role{ID: 8, Name: domain.RolePlatformAdmin}

	require.NoError(t, svc.AddPlatformRole(context.Background(), 21, domain.RolePlatformAdmin))
	items, err := svc.ListPlatformRoles(context.Background(), 21)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint64(8), items[0].RoleID)

	require.NoError(t, svc.RemovePlatformRole(context.Background(), 21, domain.RolePlatformAdmin))
	items, err = svc.ListPlatformRoles(context.Background(), 21)
	require.NoError(t, err)
	require.Len(t, items, 0)
}

func TestIAMService_AttachAndListDirectPolicies(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["managed/platform_owner"] = domain.Policy{
		ID:    5,
		Scope: domain.PolicyScopePlatform,
		Name:  "managed/platform_owner",
	}
	state.policiesByName["tenant/custom"] = domain.Policy{ID: 6, Scope: domain.PolicyScopeTenant, Name: "tenant/custom"}
	state.policiesByID[5] = state.policiesByName["managed/platform_owner"]
	state.policiesByID[6] = state.policiesByName["tenant/custom"]

	require.NoError(t, svc.AttachPlatformUserPolicy(context.Background(), 21, "managed/platform_owner"))
	platformPolicies, err := svc.ListPlatformUserPolicies(context.Background(), 21)
	require.NoError(t, err)
	require.Len(t, platformPolicies, 1)
	require.Equal(t, "managed/platform_owner", platformPolicies[0].Name)

	require.NoError(t, svc.AttachTenantUserPolicy(context.Background(), "t1", 21, "tenant/custom"))
	tenantPolicies, err := svc.ListTenantUserPolicies(context.Background(), "t1", 21)
	require.NoError(t, err)
	require.Len(t, tenantPolicies, 1)
	require.Equal(t, "tenant/custom", tenantPolicies[0].Name)
}

func TestIAMService_GetPolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["tenant/orders_editor"] = domain.Policy{
		ID:    9,
		Scope: domain.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
	}
	state.policiesByID[9] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[9] = []domain.PolicyStatement{
		{
			ID:              1,
			PolicyID:        9,
			PolicyName:      "tenant/orders_editor",
			Effect:          domain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	policy, statements, err := svc.GetPolicy(context.Background(), "tenant/orders_editor")
	require.NoError(t, err)
	require.NotNil(t, policy)
	require.Len(t, statements, 1)
	require.Equal(t, uint64(9), policy.ID)
}

func TestIAMService_ListPolicyAttachments(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["tenant/orders_editor"] = domain.Policy{
		ID:    9,
		Scope: domain.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
	}
	state.policiesByID[9] = state.policiesByName["tenant/orders_editor"]
	state.policyAttachments[9] = []domain.PolicyAttachment{
		{AttachmentType: "role", RoleID: 2, RoleName: "tenant_editor"},
		{AttachmentType: "group", Scope: domain.PolicyScopeTenant, TenantID: "t1", GroupID: 12, GroupName: "ops-team"},
	}

	items, err := svc.ListPolicyAttachments(context.Background(), "tenant/orders_editor")
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "tenant_editor", items[0].RoleName)
	require.Equal(t, uint64(12), items[1].GroupID)
}

func TestIAMService_DeletePolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["tenant/orders_editor"] = domain.Policy{
		ID:       9,
		Scope:    domain.PolicyScopeTenant,
		Name:     "tenant/orders_editor",
		IsSystem: false,
	}
	state.policiesByID[9] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[9] = []domain.PolicyStatement{
		{
			ID:              1,
			PolicyID:        9,
			PolicyName:      "tenant/orders_editor",
			Effect:          domain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	require.NoError(t, svc.DeletePolicy(context.Background(), "tenant/orders_editor"))
	_, ok := state.policiesByName["tenant/orders_editor"]
	require.False(t, ok)
}

func TestIAMService_DeletePolicy_BlocksSystem(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["managed/tenant_owner"] = domain.Policy{
		ID:       5,
		Scope:    domain.PolicyScopeTenant,
		Name:     "managed/tenant_owner",
		IsSystem: true,
	}
	state.policiesByID[5] = state.policiesByName["managed/tenant_owner"]

	err := svc.DeletePolicy(context.Background(), "managed/tenant_owner")
	require.ErrorIs(t, err, domain.ErrImmutablePolicy)
}

func TestIAMService_GroupPoliciesAffectPermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantViewer] = domain.Role{ID: 4, Name: domain.RoleTenantViewer}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: domain.RoleTenantViewer,
		Status:   domain.MembershipStatusActive,
	}
	state.policiesByName["tenant/orders_editor"] = domain.Policy{
		ID:    15,
		Scope: domain.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
	}
	state.policiesByID[15] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[15] = []domain.PolicyStatement{
		{
			ID:              1,
			PolicyID:        15,
			PolicyName:      "tenant/orders_editor",
			Effect:          domain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	group, err := svc.CreateGroup(context.Background(), domain.CreateGroupInput{
		Scope:    domain.PolicyScopeTenant,
		TenantID: "t1",
		Name:     "ops-team",
	})
	require.NoError(t, err)
	require.NotNil(t, group)
	require.NoError(t, svc.AddGroupMember(context.Background(), group.ID, 9))
	require.NoError(t, svc.AttachGroupPolicy(context.Background(), group.ID, "tenant/orders_editor"))
	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"))
}

func TestIAMService_DeleteGroup(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	group, err := svc.CreateGroup(context.Background(), domain.CreateGroupInput{
		Scope:    domain.PolicyScopeTenant,
		TenantID: "t1",
		Name:     "ops-team",
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteGroup(context.Background(), group.ID))
	_, ok := state.groupsByID[group.ID]
	require.False(t, ok)
}

func TestIAMService_GroupInlinePolicyAffectsPermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantViewer] = domain.Role{ID: 4, Name: domain.RoleTenantViewer}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: domain.RoleTenantViewer,
		Status:   domain.MembershipStatusActive,
	}

	group, err := svc.CreateGroup(context.Background(), domain.CreateGroupInput{
		Scope:    domain.PolicyScopeTenant,
		TenantID: "t1",
		Name:     "ops-inline",
	})
	require.NoError(t, err)
	require.NoError(t, svc.AddGroupMember(context.Background(), group.ID, 9))
	require.NoError(t, svc.PutGroupInlinePolicy(context.Background(), domain.PutGroupInlinePolicyInput{
		GroupID: group.ID,
		Name:    "inline-ops",
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	}))

	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"))
}

func TestIAMService_TenantUserInlinePolicyAffectsPermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantViewer] = domain.Role{ID: 4, Name: domain.RoleTenantViewer}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: domain.RoleTenantViewer,
		Status:   domain.MembershipStatusActive,
	}

	require.NoError(t, svc.PutTenantUserInlinePolicy(context.Background(), domain.PutTenantUserInlinePolicyInput{
		TenantID: "t1",
		UserID:   9,
		Name:     "inline-tenant",
		Statements: []domain.PolicyStatement{
			{Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	}))

	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"))
}

func TestIAMService_TenantUserPermissionBoundaryScopesDown(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantEditor] = domain.Role{ID: 4, Name: domain.RoleTenantEditor}
	state.memberships.items[membershipKey("t1", 9)] = domain.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: domain.RoleTenantEditor,
		Status:   domain.MembershipStatusActive,
	}
	state.roleStatements[4] = []domain.PolicyStatement{
		{PolicyID: 15, PolicyName: "managed/tenant_editor", Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
	}
	state.policiesByName["tenant/orders-readonly-boundary"] = domain.Policy{ID: 22, Scope: domain.PolicyScopeTenant, Name: "tenant/orders-readonly-boundary"}
	state.policiesByID[22] = state.policiesByName["tenant/orders-readonly-boundary"]
	state.policyStatements[22] = []domain.PolicyStatement{
		{ID: 1, PolicyID: 22, PolicyName: "tenant/orders-readonly-boundary", Effect: domain.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
	}

	require.NoError(t, svc.PutTenantUserPermissionBoundary(context.Background(), "t1", 9, "tenant/orders-readonly-boundary"))
	require.ErrorIs(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"), domain.ErrPermissionDenied)
}

func TestIAMService_CreateAndAcceptInvite(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantViewer] = domain.Role{ID: 3, Name: domain.RoleTenantViewer}

	invite, rawToken, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", domain.RoleTenantViewer, 7)
	require.NoError(t, err)
	require.NotEmpty(t, rawToken)
	require.Equal(t, domain.InviteStatusPending, invite.Status)
	require.NotEmpty(t, invite.TokenHash)

	membership, err := svc.AcceptInvite(context.Background(), rawToken, 11, "neo@mx.io")
	require.NoError(t, err)
	require.Equal(t, tenant.ID, membership.TenantID)
	require.Equal(t, uint(11), membership.UserID)
	require.Equal(t, domain.RoleTenantViewer, membership.RoleName)

	storedInvite, err := state.invites.GetByID(context.Background(), invite.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InviteStatusAccepted, storedInvite.Status)
	require.NotNil(t, storedInvite.AcceptedByUserID)
	require.Equal(t, uint(11), *storedInvite.AcceptedByUserID)
}

func TestIAMService_AcceptInvite_EmailMismatch(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantViewer] = domain.Role{ID: 3, Name: domain.RoleTenantViewer}

	_, rawToken, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", domain.RoleTenantViewer, 7)
	require.NoError(t, err)

	_, err = svc.AcceptInvite(context.Background(), rawToken, 11, "trinity@mx.io")
	require.ErrorIs(t, err, domain.ErrInviteEmailMismatch)
}

func TestIAMService_RevokeInvite(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := domain.Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[domain.RoleTenantViewer] = domain.Role{ID: 3, Name: domain.RoleTenantViewer}

	invite, _, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", domain.RoleTenantViewer, 7)
	require.NoError(t, err)

	require.NoError(t, svc.RevokeInvite(context.Background(), invite.ID))

	storedInvite, err := state.invites.GetByID(context.Background(), invite.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InviteStatusRevoked, storedInvite.Status)
	require.NotNil(t, storedInvite.RevokedAt)
}

type iamTestState struct {
	tenants                 map[string]domain.Tenant
	roleByName              map[string]domain.Role
	roleTrustStatements     map[uint64][]domain.RoleTrustStatement
	rolePermissions         map[uint64]map[string]bool
	policiesByName          map[string]domain.Policy
	policiesByID            map[uint64]domain.Policy
	policyStatements        map[uint64][]domain.PolicyStatement
	policyVersions          map[uint64][]domain.PolicyVersion
	policyVersionStatements map[string][]domain.PolicyStatement
	policyAttachments       map[uint64][]domain.PolicyAttachment
	nextPolicyID            uint64
	groupsByID              map[uint64]domain.Group
	groupInlinePolicies     map[uint64]map[string]domain.GroupInlinePolicy
	nextGroupID             uint64
	roleStatements          map[uint64][]domain.PolicyStatement
	roleBoundary            map[uint64]*domain.RolePermissionBoundary
	roleBoundaryStmts       map[uint64][]domain.PolicyStatement
	platformDirect          map[uint][]domain.PolicyStatement
	tenantDirect            map[string][]domain.PolicyStatement
	platformBoundary        map[uint]*domain.PermissionBoundary
	platformBoundaryStmts   map[uint][]domain.PolicyStatement
	platformInlinePolicies  map[uint]map[string]domain.UserInlinePolicy
	tenantBoundary          map[string]*domain.PermissionBoundary
	tenantBoundaryStmts     map[string][]domain.PolicyStatement
	tenantInlinePolicies    map[string]map[string]domain.UserInlinePolicy
	platformRoleIDs         map[uint][]uint64
	memberships             *membershipState
	invites                 *inviteState
}

type membershipState struct {
	items map[string]domain.Membership
}

func (r *membershipState) GetByTenantAndUser(
	_ context.Context,
	tenantID string,
	userID uint,
) (*domain.Membership, error) {
	item, ok := r.items[membershipKey(tenantID, userID)]
	if !ok {
		return nil, domain.ErrMembershipNotFound
	}
	copyItem := item
	return &copyItem, nil
}

type inviteState struct {
	items      map[string]domain.TenantInvite
	tokenIndex map[string]string
}

func (s *iamTestState) roleByID(roleID uint64) domain.Role {
	for _, role := range s.roleByName {
		if role.ID == roleID {
			return role
		}
	}
	return domain.Role{ID: roleID}
}

func (r *inviteState) GetByID(_ context.Context, inviteID string) (*domain.TenantInvite, error) {
	item, ok := r.items[inviteID]
	if !ok {
		return nil, domain.ErrInviteNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func newIAMTestUsecase(t *testing.T) (domain.IAMUsecase, *iamTestState) {
	t.Helper()

	tenantRepo := domainmocks.NewMockTenantRepository(t)
	roleRepo := domainmocks.NewMockRoleRepository(t)
	policyRepo := domainmocks.NewMockPolicyRepository(t)
	groupRepo := domainmocks.NewMockGroupRepository(t)
	orgRepo := domainmocks.NewMockOrganizationRepository(t)
	platformRepo := domainmocks.NewMockPlatformMembershipRepository(t)
	membershipRepo := domainmocks.NewMockMembershipRepository(t)
	inviteRepo := domainmocks.NewMockInviteRepository(t)

	state := &iamTestState{
		tenants:                 map[string]domain.Tenant{},
		roleByName:              map[string]domain.Role{},
		roleTrustStatements:     map[uint64][]domain.RoleTrustStatement{},
		rolePermissions:         map[uint64]map[string]bool{},
		policiesByName:          map[string]domain.Policy{},
		policiesByID:            map[uint64]domain.Policy{},
		policyStatements:        map[uint64][]domain.PolicyStatement{},
		policyVersions:          map[uint64][]domain.PolicyVersion{},
		policyVersionStatements: map[string][]domain.PolicyStatement{},
		policyAttachments:       map[uint64][]domain.PolicyAttachment{},
		nextPolicyID:            100,
		groupsByID:              map[uint64]domain.Group{},
		groupInlinePolicies:     map[uint64]map[string]domain.GroupInlinePolicy{},
		nextGroupID:             200,
		roleStatements:          map[uint64][]domain.PolicyStatement{},
		roleBoundary:            map[uint64]*domain.RolePermissionBoundary{},
		roleBoundaryStmts:       map[uint64][]domain.PolicyStatement{},
		platformDirect:          map[uint][]domain.PolicyStatement{},
		tenantDirect:            map[string][]domain.PolicyStatement{},
		platformBoundary:        map[uint]*domain.PermissionBoundary{},
		platformBoundaryStmts:   map[uint][]domain.PolicyStatement{},
		platformInlinePolicies:  map[uint]map[string]domain.UserInlinePolicy{},
		tenantBoundary:          map[string]*domain.PermissionBoundary{},
		tenantBoundaryStmts:     map[string][]domain.PolicyStatement{},
		tenantInlinePolicies:    map[string]map[string]domain.UserInlinePolicy{},
		platformRoleIDs:         map[uint][]uint64{},
		memberships:             &membershipState{items: map[string]domain.Membership{}},
		invites:                 &inviteState{items: map[string]domain.TenantInvite{}, tokenIndex: map[string]string{}},
	}

	orgRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, org domain.Organization) (*domain.Organization, error) {
			copyOrg := org
			return &copyOrg, nil
		}).Maybe()
	orgRepo.EXPECT().
		List(mock.Anything).
		RunAndReturn(func(ctx context.Context) ([]domain.Organization, error) {
			return nil, nil
		}).Maybe()
	orgRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, orgID string) (*domain.Organization, error) {
			if orgID == "" {
				return nil, domain.ErrOrganizationNotFound
			}
			return &domain.Organization{ID: orgID, Slug: orgID, Name: orgID}, nil
		}).Maybe()
	orgRepo.EXPECT().
		AttachServiceControlPolicy(mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
	orgRepo.EXPECT().
		DetachServiceControlPolicy(mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
	orgRepo.EXPECT().
		ListServiceControlPolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, orgID string) ([]domain.Policy, error) { return nil, nil }).Maybe()
	orgRepo.EXPECT().
		ListServiceControlPolicyStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, orgID string) ([]domain.PolicyStatement, error) { return nil, nil }).Maybe()

	tenantRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenant domain.Tenant) (*domain.Tenant, error) {
			copyTenant := tenant
			state.tenants[tenant.ID] = tenant
			return &copyTenant, nil
		}).
		Maybe()
	tenantRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) (*domain.Tenant, error) {
			tenant, ok := state.tenants[tenantID]
			if !ok {
				return nil, domain.ErrTenantNotFound
			}
			copyTenant := tenant
			return &copyTenant, nil
		}).
		Maybe()
	tenantRepo.EXPECT().
		AttachOrganization(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, orgID string) error {
			tenant := state.tenants[tenantID]
			tenant.OrgID = orgID
			state.tenants[tenantID] = tenant
			return nil
		}).Maybe()
	tenantRepo.EXPECT().
		DetachOrganization(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) error {
			tenant := state.tenants[tenantID]
			tenant.OrgID = ""
			state.tenants[tenantID] = tenant
			return nil
		}).Maybe()

	roleRepo.EXPECT().
		GetByName(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, name string) (*domain.Role, error) {
			role, ok := state.roleByName[name]
			if !ok {
				return nil, domain.ErrRoleNotFound
			}
			copyRole := role
			return &copyRole, nil
		}).
		Maybe()
	roleRepo.EXPECT().
		RoleHasPermission(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64, permission string) (bool, error) {
			return state.rolePermissions[roleID][permission], nil
		}).
		Maybe()
	roleRepo.EXPECT().
		PutTrustPolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64, statements []domain.RoleTrustStatement) error {
			state.roleTrustStatements[roleID] = append([]domain.RoleTrustStatement(nil), statements...)
			return nil
		}).
		Maybe()
	roleRepo.EXPECT().
		GetTrustPolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) ([]domain.RoleTrustStatement, error) {
			return append([]domain.RoleTrustStatement(nil), state.roleTrustStatements[roleID]...), nil
		}).
		Maybe()
	roleRepo.EXPECT().
		DeleteTrustPolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) error {
			delete(state.roleTrustStatements, roleID)
			return nil
		}).
		Maybe()
	roleRepo.EXPECT().
		PutPermissionBoundary(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64, policyID uint64) error {
			policy := state.policiesByID[policyID]
			state.roleBoundary[roleID] = &domain.RolePermissionBoundary{
				RoleID:     roleID,
				RoleName:   state.roleByID(roleID).Name,
				PolicyID:   policyID,
				PolicyName: policy.Name,
				CreatedAt:  time.Now().UTC(),
			}
			state.roleBoundaryStmts[roleID] = append([]domain.PolicyStatement(nil), state.policyStatements[policyID]...)
			return nil
		}).
		Maybe()
	roleRepo.EXPECT().
		GetPermissionBoundary(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) (*domain.RolePermissionBoundary, error) {
			item := state.roleBoundary[roleID]
			if item == nil {
				return nil, nil
			}
			copyItem := *item
			return &copyItem, nil
		}).
		Maybe()
	roleRepo.EXPECT().
		GetPermissionBoundaryStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) ([]domain.PolicyStatement, error) {
			return append([]domain.PolicyStatement(nil), state.roleBoundaryStmts[roleID]...), nil
		}).
		Maybe()
	roleRepo.EXPECT().
		DeletePermissionBoundary(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) error {
			delete(state.roleBoundary, roleID)
			delete(state.roleBoundaryStmts, roleID)
			return nil
		}).
		Maybe()

	policyRepo.EXPECT().
		CreatePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policy domain.Policy, statements []domain.PolicyStatement) (*domain.Policy, []domain.PolicyStatement, error) {
			policy.ID = state.nextPolicyID
			state.nextPolicyID++
			policy.DefaultVersion = "v1"
			state.policiesByName[policy.Name] = policy
			state.policiesByID[policy.ID] = policy
			outStatements := make([]domain.PolicyStatement, 0, len(statements))
			for i, statement := range statements {
				statement.ID = uint64(i + 1)
				statement.PolicyID = policy.ID
				statement.PolicyName = policy.Name
				outStatements = append(outStatements, statement)
			}
			state.policyStatements[policy.ID] = append([]domain.PolicyStatement(nil), outStatements...)
			state.policyVersions[policy.ID] = []domain.PolicyVersion{{
				ID:         1,
				PolicyID:   policy.ID,
				PolicyName: policy.Name,
				Version:    "v1",
				IsDefault:  true,
				CreatedAt:  policy.CreatedAt,
			}}
			state.policyVersionStatements[fmt.Sprintf("%d:%s", policy.ID, "v1")] = append([]domain.PolicyStatement(nil), outStatements...)
			return &policy, outStatements, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		CreatePolicyVersion(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, policyName string, statements []domain.PolicyStatement, setAsDefault bool) (*domain.PolicyVersion, []domain.PolicyStatement, error) {
			versions := state.policyVersions[policyID]
			versionLabel := fmt.Sprintf("v%d", len(versions)+1)
			version := domain.PolicyVersion{
				ID:         uint64(len(versions) + 1),
				PolicyID:   policyID,
				PolicyName: policyName,
				Version:    versionLabel,
				IsDefault:  setAsDefault,
				CreatedAt:  time.Now().UTC(),
			}
			if setAsDefault {
				for i := range versions {
					versions[i].IsDefault = false
				}
				policy := state.policiesByID[policyID]
				policy.DefaultVersion = versionLabel
				state.policiesByID[policyID] = policy
				state.policiesByName[policy.Name] = policy
				state.policyStatements[policyID] = append([]domain.PolicyStatement(nil), statements...)
			}
			state.policyVersions[policyID] = append(versions, version)
			state.policyVersionStatements[fmt.Sprintf("%d:%s", policyID, versionLabel)] = append([]domain.PolicyStatement(nil), statements...)
			return &version, append([]domain.PolicyStatement(nil), statements...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePolicyVersion(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, version string) error {
			versions := state.policyVersions[policyID]
			nextVersions := make([]domain.PolicyVersion, 0, len(versions))
			found := false
			for _, item := range versions {
				if item.Version != version {
					nextVersions = append(nextVersions, item)
					continue
				}
				if item.IsDefault {
					return domain.ErrDefaultPolicyVersion
				}
				found = true
			}
			if !found {
				return domain.ErrPolicyVersionNotFound
			}
			state.policyVersions[policyID] = nextVersions
			delete(state.policyVersionStatements, fmt.Sprintf("%d:%s", policyID, version))
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPolicyByName(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, name string) (*domain.Policy, error) {
			policy, ok := state.policiesByName[name]
			if !ok {
				return nil, domain.ErrRoleNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPolicyStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64) ([]domain.PolicyStatement, error) {
			return append([]domain.PolicyStatement(nil), state.policyStatements[policyID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPolicyVersions(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, policyName string) ([]domain.PolicyVersion, error) {
			return append([]domain.PolicyVersion(nil), state.policyVersions[policyID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		SetDefaultPolicyVersion(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, version string) error {
			versions := state.policyVersions[policyID]
			for i := range versions {
				versions[i].IsDefault = versions[i].Version == version
			}
			state.policyVersions[policyID] = versions
			policy := state.policiesByID[policyID]
			policy.DefaultVersion = version
			state.policiesByID[policyID] = policy
			state.policiesByName[policy.Name] = policy
			state.policyStatements[policyID] = append([]domain.PolicyStatement(nil), state.policyVersionStatements[fmt.Sprintf("%d:%s", policyID, version)]...)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, scope string) ([]domain.Policy, error) {
			out := make([]domain.Policy, 0)
			for _, policy := range state.policiesByName {
				if scope == "" || policy.Scope == scope {
					out = append(out, policy)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPolicyAttachments(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64) ([]domain.PolicyAttachment, error) {
			return append([]domain.PolicyAttachment(nil), state.policyAttachments[policyID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64) error {
			policy, ok := state.policiesByID[policyID]
			if !ok {
				return domain.ErrPolicyNotFound
			}
			if policy.IsSystem {
				return domain.ErrImmutablePolicy
			}
			for _, statements := range state.tenantDirect {
				for _, statement := range statements {
					if statement.PolicyID == policyID {
						return domain.ErrPolicyInUse
					}
				}
			}
			for _, statements := range state.platformDirect {
				for _, statement := range statements {
					if statement.PolicyID == policyID {
						return domain.ErrPolicyInUse
					}
				}
			}
			for _, statements := range state.roleStatements {
				for _, statement := range statements {
					if statement.PolicyID == policyID {
						return domain.ErrPolicyInUse
					}
				}
			}
			delete(state.policiesByName, policy.Name)
			delete(state.policiesByID, policyID)
			delete(state.policyStatements, policyID)
			delete(state.policyAttachments, policyID)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListRoleStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) ([]domain.PolicyStatement, error) {
			return append([]domain.PolicyStatement(nil), state.roleStatements[roleID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPlatformUserStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]domain.PolicyStatement, error) {
			out := append([]domain.PolicyStatement(nil), state.platformDirect[userID]...)
			for _, inline := range state.platformInlinePolicies[userID] {
				out = append(out, inline.Statements...)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListTenantUserStatements(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]domain.PolicyStatement, error) {
			key := membershipKey(tenantID, userID)
			out := append([]domain.PolicyStatement(nil), state.tenantDirect[key]...)
			for _, inline := range state.tenantInlinePolicies[key] {
				out = append(out, inline.Statements...)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPlatformGroupStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]domain.PolicyStatement, error) {
			out := make([]domain.PolicyStatement, 0)
			for _, group := range state.groupsByID {
				if group.Scope != domain.PolicyScopePlatform {
					continue
				}
				if _, ok := state.memberships.items[membershipKey(fmt.Sprintf("group:%d", group.ID), userID)]; !ok {
					continue
				}
				out = append(out, state.tenantDirect[fmt.Sprintf("group-policy:%d", group.ID)]...)
				for _, inline := range state.groupInlinePolicies[group.ID] {
					out = append(out, inline.Statements...)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListTenantGroupStatements(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]domain.PolicyStatement, error) {
			out := make([]domain.PolicyStatement, 0)
			for _, group := range state.groupsByID {
				if group.Scope != domain.PolicyScopeTenant || group.TenantID != tenantID {
					continue
				}
				if _, ok := state.memberships.items[membershipKey(fmt.Sprintf("group:%d", group.ID), userID)]; !ok {
					continue
				}
				out = append(out, state.tenantDirect[fmt.Sprintf("group-policy:%d", group.ID)]...)
				for _, inline := range state.groupInlinePolicies[group.ID] {
					out = append(out, inline.Statements...)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		AttachPlatformUserPolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			state.platformDirect[userID] = append(state.platformDirect[userID], domain.PolicyStatement{
				PolicyID:        policy.ID,
				PolicyName:      policy.Name,
				Effect:          domain.PolicyEffectAllow,
				ActionPattern:   "*",
				ResourcePattern: "*",
			})
			state.policyAttachments[policyID] = append(state.policyAttachments[policyID], domain.PolicyAttachment{
				AttachmentType: "platform_user",
				Scope:          domain.PolicyScopePlatform,
				UserID:         userID,
			})
			return nil
		}).
		Maybe()

	groupRepo.EXPECT().
		CreateGroup(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, group domain.Group) (*domain.Group, error) {
			group.ID = state.nextGroupID
			state.nextGroupID++
			state.groupsByID[group.ID] = group
			copyGroup := group
			return &copyGroup, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		ListGroups(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, scope string, tenantID string) ([]domain.Group, error) {
			out := make([]domain.Group, 0)
			for _, group := range state.groupsByID {
				if scope != "" && group.Scope != scope {
					continue
				}
				if tenantID != "" && group.TenantID != tenantID {
					continue
				}
				out = append(out, group)
			}
			return out, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		DeleteGroup(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64) error {
			group, ok := state.groupsByID[groupID]
			if !ok {
				return domain.ErrGroupNotFound
			}
			if group.IsSystem {
				return domain.ErrImmutableGroup
			}
			delete(state.groupsByID, groupID)
			delete(state.groupInlinePolicies, groupID)
			for key := range state.memberships.items {
				if strings.HasPrefix(key, fmt.Sprintf("group:%d:", groupID)) {
					delete(state.memberships.items, key)
				}
			}
			delete(state.tenantDirect, fmt.Sprintf("group-policy:%d", groupID))
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		PutInlinePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, input domain.PutGroupInlinePolicyInput) error {
			if state.groupInlinePolicies[input.GroupID] == nil {
				state.groupInlinePolicies[input.GroupID] = map[string]domain.GroupInlinePolicy{}
			}
			state.groupInlinePolicies[input.GroupID][input.Name] = domain.GroupInlinePolicy{
				GroupID:     input.GroupID,
				Name:        input.Name,
				Description: input.Description,
				Statements:  append([]domain.PolicyStatement(nil), input.Statements...),
			}
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		GetInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, name string) (*domain.GroupInlinePolicy, error) {
			policies := state.groupInlinePolicies[groupID]
			policy, ok := policies[name]
			if !ok {
				return nil, domain.ErrPolicyNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		ListInlinePolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64) ([]domain.GroupInlinePolicy, error) {
			policies := state.groupInlinePolicies[groupID]
			out := make([]domain.GroupInlinePolicy, 0, len(policies))
			for _, policy := range policies {
				out = append(out, policy)
			}
			return out, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		DeleteInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, name string) error {
			if state.groupInlinePolicies[groupID] == nil {
				return domain.ErrPolicyNotFound
			}
			delete(state.groupInlinePolicies[groupID], name)
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		AddMember(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, userID uint) error {
			state.memberships.items[membershipKey(fmt.Sprintf("group:%d", groupID), userID)] = domain.Membership{
				TenantID: fmt.Sprintf("group:%d", groupID),
				UserID:   userID,
			}
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		AttachPolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, policyID uint64) error {
			group := state.groupsByID[groupID]
			state.tenantDirect[fmt.Sprintf("group-policy:%d", groupID)] = append(
				state.tenantDirect[fmt.Sprintf("group-policy:%d", groupID)],
				state.policyStatements[policyID]...,
			)
			state.policyAttachments[policyID] = append(state.policyAttachments[policyID], domain.PolicyAttachment{
				AttachmentType: "group",
				Scope:          group.Scope,
				TenantID:       group.TenantID,
				GroupID:        group.ID,
				GroupName:      group.Name,
			})
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DetachPlatformUserPolicy(mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Maybe()
	policyRepo.EXPECT().
		ListPlatformUserPolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]domain.Policy, error) {
			out := make([]domain.Policy, 0, len(state.platformDirect[userID]))
			for _, statement := range state.platformDirect[userID] {
				if policy, ok := state.policiesByID[statement.PolicyID]; ok {
					out = append(out, policy)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutPlatformUserInlinePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, input domain.PutPlatformUserInlinePolicyInput) error {
			if state.platformInlinePolicies[input.UserID] == nil {
				state.platformInlinePolicies[input.UserID] = map[string]domain.UserInlinePolicy{}
			}
			state.platformInlinePolicies[input.UserID][input.Name] = domain.UserInlinePolicy{
				Scope:       domain.PolicyScopePlatform,
				UserID:      input.UserID,
				Name:        input.Name,
				Description: input.Description,
				Statements:  append([]domain.PolicyStatement(nil), input.Statements...),
			}
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPlatformUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, name string) (*domain.UserInlinePolicy, error) {
			policies := state.platformInlinePolicies[userID]
			policy, ok := policies[name]
			if !ok {
				return nil, domain.ErrPolicyNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPlatformUserInlinePolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]domain.UserInlinePolicy, error) {
			policies := state.platformInlinePolicies[userID]
			out := make([]domain.UserInlinePolicy, 0, len(policies))
			for _, policy := range policies {
				out = append(out, policy)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePlatformUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, name string) error {
			if state.platformInlinePolicies[userID] == nil {
				return domain.ErrPolicyNotFound
			}
			delete(state.platformInlinePolicies[userID], name)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutPlatformUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			boundary := &domain.PermissionBoundary{
				Scope:      domain.PolicyScopePlatform,
				UserID:     userID,
				PolicyID:   policyID,
				PolicyName: policy.Name,
				CreatedAt:  time.Now().UTC(),
			}
			state.platformBoundary[userID] = boundary
			state.platformBoundaryStmts[userID] = append([]domain.PolicyStatement(nil), state.policyStatements[policyID]...)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPlatformUserPermissionBoundary(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) (*domain.PermissionBoundary, error) {
			item := state.platformBoundary[userID]
			if item == nil {
				return nil, nil
			}
			copyItem := *item
			return &copyItem, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPlatformUserPermissionBoundaryStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]domain.PolicyStatement, error) {
			return append([]domain.PolicyStatement(nil), state.platformBoundaryStmts[userID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePlatformUserPermissionBoundary(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) error {
			delete(state.platformBoundary, userID)
			delete(state.platformBoundaryStmts, userID)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		AttachTenantUserPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			key := membershipKey(tenantID, userID)
			state.tenantDirect[key] = append(state.tenantDirect[key], domain.PolicyStatement{
				PolicyID:        policy.ID,
				PolicyName:      policy.Name,
				Effect:          domain.PolicyEffectAllow,
				ActionPattern:   "*",
				ResourcePattern: "*",
			})
			state.policyAttachments[policyID] = append(state.policyAttachments[policyID], domain.PolicyAttachment{
				AttachmentType: "tenant_user",
				Scope:          domain.PolicyScopeTenant,
				TenantID:       tenantID,
				UserID:         userID,
			})
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DetachTenantUserPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Maybe()
	policyRepo.EXPECT().
		ListTenantUserPolicies(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]domain.Policy, error) {
			key := membershipKey(tenantID, userID)
			out := make([]domain.Policy, 0, len(state.tenantDirect[key]))
			for _, statement := range state.tenantDirect[key] {
				if policy, ok := state.policiesByID[statement.PolicyID]; ok {
					out = append(out, policy)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutTenantUserInlinePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, input domain.PutTenantUserInlinePolicyInput) error {
			key := membershipKey(input.TenantID, input.UserID)
			if state.tenantInlinePolicies[key] == nil {
				state.tenantInlinePolicies[key] = map[string]domain.UserInlinePolicy{}
			}
			state.tenantInlinePolicies[key][input.Name] = domain.UserInlinePolicy{
				Scope:       domain.PolicyScopeTenant,
				TenantID:    input.TenantID,
				UserID:      input.UserID,
				Name:        input.Name,
				Description: input.Description,
				Statements:  append([]domain.PolicyStatement(nil), input.Statements...),
			}
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetTenantUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, name string) (*domain.UserInlinePolicy, error) {
			policies := state.tenantInlinePolicies[membershipKey(tenantID, userID)]
			policy, ok := policies[name]
			if !ok {
				return nil, domain.ErrPolicyNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListTenantUserInlinePolicies(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]domain.UserInlinePolicy, error) {
			policies := state.tenantInlinePolicies[membershipKey(tenantID, userID)]
			out := make([]domain.UserInlinePolicy, 0, len(policies))
			for _, policy := range policies {
				out = append(out, policy)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeleteTenantUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, name string) error {
			key := membershipKey(tenantID, userID)
			if state.tenantInlinePolicies[key] == nil {
				return domain.ErrPolicyNotFound
			}
			delete(state.tenantInlinePolicies[key], name)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutTenantUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			key := membershipKey(tenantID, userID)
			boundary := &domain.PermissionBoundary{
				Scope:      domain.PolicyScopeTenant,
				TenantID:   tenantID,
				UserID:     userID,
				PolicyID:   policyID,
				PolicyName: policy.Name,
				CreatedAt:  time.Now().UTC(),
			}
			state.tenantBoundary[key] = boundary
			state.tenantBoundaryStmts[key] = append([]domain.PolicyStatement(nil), state.policyStatements[policyID]...)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetTenantUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) (*domain.PermissionBoundary, error) {
			item := state.tenantBoundary[membershipKey(tenantID, userID)]
			if item == nil {
				return nil, nil
			}
			copyItem := *item
			return &copyItem, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetTenantUserPermissionBoundaryStatements(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]domain.PolicyStatement, error) {
			return append([]domain.PolicyStatement(nil), state.tenantBoundaryStmts[membershipKey(tenantID, userID)]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeleteTenantUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) error {
			key := membershipKey(tenantID, userID)
			delete(state.tenantBoundary, key)
			delete(state.tenantBoundaryStmts, key)
			return nil
		}).
		Maybe()

	platformRepo.EXPECT().
		Upsert(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, roleID uint64, status string) error {
			state.platformRoleIDs[userID] = append(state.platformRoleIDs[userID], roleID)
			return nil
		}).
		Maybe()
	platformRepo.EXPECT().
		ListRoleIDsByUser(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]uint64, error) {
			return append([]uint64(nil), state.platformRoleIDs[userID]...), nil
		}).
		Maybe()
	platformRepo.EXPECT().
		ListByUser(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]domain.PlatformMembership, error) {
			roleIDs := state.platformRoleIDs[userID]
			out := make([]domain.PlatformMembership, 0, len(roleIDs))
			for _, roleID := range roleIDs {
				out = append(out, domain.PlatformMembership{
					UserID:   userID,
					RoleID:   roleID,
					RoleName: fmt.Sprintf("role-%d", roleID),
					Status:   domain.MembershipStatusActive,
				})
			}
			return out, nil
		}).
		Maybe()
	platformRepo.EXPECT().
		Delete(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, roleID uint64) error {
			roleIDs := state.platformRoleIDs[userID]
			next := make([]uint64, 0, len(roleIDs))
			found := false
			for _, id := range roleIDs {
				if id == roleID {
					found = true
					continue
				}
				next = append(next, id)
			}
			if !found {
				return domain.ErrMembershipNotFound
			}
			state.platformRoleIDs[userID] = next
			return nil
		}).
		Maybe()

	membershipRepo.EXPECT().
		Upsert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, membership domain.Membership) error {
			state.memberships.items[membershipKey(membership.TenantID, membership.UserID)] = membership
			return nil
		}).
		Maybe()
	membershipRepo.EXPECT().
		GetByTenantAndUser(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) (*domain.Membership, error) {
			return state.memberships.GetByTenantAndUser(ctx, tenantID, userID)
		}).
		Maybe()
	membershipRepo.EXPECT().
		ListByTenant(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) ([]domain.Membership, error) {
			out := make([]domain.Membership, 0)
			for _, item := range state.memberships.items {
				if item.TenantID == tenantID {
					out = append(out, item)
				}
			}
			return out, nil
		}).
		Maybe()
	membershipRepo.EXPECT().
		ListByUser(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]domain.Membership, error) {
			out := make([]domain.Membership, 0)
			for _, item := range state.memberships.items {
				if item.UserID == userID {
					out = append(out, item)
				}
			}
			return out, nil
		}).
		Maybe()
	membershipRepo.EXPECT().
		Delete(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) error {
			key := membershipKey(tenantID, userID)
			if _, ok := state.memberships.items[key]; !ok {
				return domain.ErrMembershipNotFound
			}
			delete(state.memberships.items, key)
			return nil
		}).
		Maybe()

	inviteRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, invite domain.TenantInvite) error {
			state.invites.items[invite.ID] = invite
			state.invites.tokenIndex[invite.TokenHash] = invite.ID
			return nil
		}).
		Maybe()
	inviteRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, inviteID string) (*domain.TenantInvite, error) {
			return state.invites.GetByID(ctx, inviteID)
		}).
		Maybe()
	inviteRepo.EXPECT().
		GetByTokenHash(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tokenHash string) (*domain.TenantInvite, error) {
			inviteID, ok := state.invites.tokenIndex[tokenHash]
			if !ok {
				return nil, domain.ErrInviteNotFound
			}
			return state.invites.GetByID(ctx, inviteID)
		}).
		Maybe()
	inviteRepo.EXPECT().
		ListByTenant(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) ([]domain.TenantInvite, error) {
			out := make([]domain.TenantInvite, 0)
			for _, item := range state.invites.items {
				if item.TenantID == tenantID {
					out = append(out, item)
				}
			}
			return out, nil
		}).
		Maybe()
	inviteRepo.EXPECT().
		MarkAccepted(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, inviteID string, acceptedByUserID uint, acceptedAt time.Time) error {
			item, ok := state.invites.items[inviteID]
			if !ok {
				return domain.ErrInviteNotFound
			}
			item.Status = domain.InviteStatusAccepted
			item.AcceptedByUserID = &acceptedByUserID
			item.AcceptedAt = &acceptedAt
			item.UpdatedAt = acceptedAt
			state.invites.items[inviteID] = item
			return nil
		}).
		Maybe()
	inviteRepo.EXPECT().
		MarkRevoked(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, inviteID string, revokedAt time.Time) error {
			item, ok := state.invites.items[inviteID]
			if !ok {
				return domain.ErrInviteNotFound
			}
			item.Status = domain.InviteStatusRevoked
			item.RevokedAt = &revokedAt
			item.UpdatedAt = revokedAt
			state.invites.items[inviteID] = item
			return nil
		}).
		Maybe()

	return domain.NewIAMUsecase(
		tenantRepo,
		roleRepo,
		policyRepo,
		groupRepo,
		orgRepo,
		platformRepo,
		membershipRepo,
		inviteRepo,
	), state
}

func membershipKey(tenantID string, userID uint) string {
	return fmt.Sprintf("%s:%d", tenantID, userID)
}
