package interactor_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
)

func TestIAMService_PlatformUserInlinePolicyAffectsPermission(t *testing.T) {
	t.Parallel()

	svc, _ := newIAMTestUsecase(t)
	require.NoError(t, svc.PutPlatformUserInlinePolicy(context.Background(), entity.PutPlatformUserInlinePolicyInput{
		UserID: 11,
		Name:   "inline-platform",
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "tenant:create", ResourcePattern: "*"},
		},
	}))
	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
}

func TestIAMService_PlatformUserPermissionBoundaryScopesDown(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.platformDirect[11] = []entity.PolicyStatement{
		{PolicyID: 20, PolicyName: "inline/platform-direct", Effect: entity.PolicyEffectAllow, ActionPattern: "tenant:create", ResourcePattern: "*"},
	}
	state.policiesByName["managed/platform-readonly-boundary"] = entity.Policy{ID: 21, Scope: entity.PolicyScopePlatform, Name: "managed/platform-readonly-boundary"}
	state.policiesByID[21] = state.policiesByName["managed/platform-readonly-boundary"]
	state.policyStatements[21] = []entity.PolicyStatement{
		{ID: 1, PolicyID: 21, PolicyName: "managed/platform-readonly-boundary", Effect: entity.PolicyEffectAllow, ActionPattern: "tenant:read", ResourcePattern: "*"},
	}

	require.NoError(t, svc.PutPlatformUserPermissionBoundary(context.Background(), 11, "managed/platform-readonly-boundary"))
	require.ErrorIs(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"), entity.ErrPermissionDenied)
}

func TestIAMService_AddAndRemovePlatformRole(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[entity.RolePlatformAdmin] = entity.Role{ID: 8, Name: entity.RolePlatformAdmin}

	require.NoError(t, svc.AddPlatformRole(context.Background(), 21, entity.RolePlatformAdmin))
	items, err := svc.ListPlatformRoles(context.Background(), 21)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint64(8), items[0].RoleID)

	require.NoError(t, svc.RemovePlatformRole(context.Background(), 21, entity.RolePlatformAdmin))
	items, err = svc.ListPlatformRoles(context.Background(), 21)
	require.NoError(t, err)
	require.Len(t, items, 0)
}

func TestIAMService_AttachAndListDirectPolicies(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["managed/platform_owner"] = entity.Policy{
		ID:    5,
		Scope: entity.PolicyScopePlatform,
		Name:  "managed/platform_owner",
	}
	state.policiesByName["tenant/custom"] = entity.Policy{ID: 6, Scope: entity.PolicyScopeTenant, Name: "tenant/custom"}
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
	require.Len(t, state.outboxRecords, 2)
	require.Equal(t, "policy.attached", state.outboxRecords[0].Envelope.Type)
	require.Equal(t, "policy.attached", state.outboxRecords[1].Envelope.Type)

	var platformPayload map[string]any
	require.NoError(t, json.Unmarshal(state.outboxRecords[0].Envelope.Payload, &platformPayload))
	require.Equal(t, "platform_user", platformPayload["attachment_type"])
	require.Equal(t, "managed/platform_owner", platformPayload["policy_name"])

	var tenantPayload map[string]any
	require.NoError(t, json.Unmarshal(state.outboxRecords[1].Envelope.Payload, &tenantPayload))
	require.Equal(t, "tenant_user", tenantPayload["attachment_type"])
	require.Equal(t, "tenant/custom", tenantPayload["policy_name"])
	require.Equal(t, "t1", tenantPayload["tenant_id"])
}

func TestIAMService_GroupPoliciesAffectPermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 4, Name: entity.RoleTenantViewer}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: entity.RoleTenantViewer,
		Status:   entity.MembershipStatusActive,
	}
	state.policiesByName["tenant/orders_editor"] = entity.Policy{
		ID:    15,
		Scope: entity.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
	}
	state.policiesByID[15] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[15] = []entity.PolicyStatement{
		{
			ID:              1,
			PolicyID:        15,
			PolicyName:      "tenant/orders_editor",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	group, err := svc.CreateGroup(context.Background(), entity.CreateGroupInput{
		Scope:    entity.PolicyScopeTenant,
		TenantID: "t1",
		Name:     "ops-team",
	})
	require.NoError(t, err)
	require.NotNil(t, group)
	require.NoError(t, svc.AddGroupMember(context.Background(), group.ID, 9))
	require.NoError(t, svc.AttachGroupPolicy(context.Background(), group.ID, "tenant/orders_editor"))
	require.Len(t, state.outboxRecords, 1)
	require.Equal(t, "policy.attached", state.outboxRecords[0].Envelope.Type)
	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"))

	var payload map[string]any
	require.NoError(t, json.Unmarshal(state.outboxRecords[0].Envelope.Payload, &payload))
	require.Equal(t, "group", payload["attachment_type"])
	require.Equal(t, "tenant/orders_editor", payload["policy_name"])
	require.Equal(t, "ops-team", payload["group_name"])
}

func TestIAMService_DeleteGroup(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	group, err := svc.CreateGroup(context.Background(), entity.CreateGroupInput{
		Scope:    entity.PolicyScopeTenant,
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
	tenant := entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 4, Name: entity.RoleTenantViewer}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: entity.RoleTenantViewer,
		Status:   entity.MembershipStatusActive,
	}

	group, err := svc.CreateGroup(context.Background(), entity.CreateGroupInput{
		Scope:    entity.PolicyScopeTenant,
		TenantID: "t1",
		Name:     "ops-inline",
	})
	require.NoError(t, err)
	require.NoError(t, svc.AddGroupMember(context.Background(), group.ID, 9))
	require.NoError(t, svc.PutGroupInlinePolicy(context.Background(), entity.PutGroupInlinePolicyInput{
		GroupID: group.ID,
		Name:    "inline-ops",
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	}))

	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"))
}

func TestIAMService_TenantUserInlinePolicyAffectsPermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 4, Name: entity.RoleTenantViewer}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: entity.RoleTenantViewer,
		Status:   entity.MembershipStatusActive,
	}

	require.NoError(t, svc.PutTenantUserInlinePolicy(context.Background(), entity.PutTenantUserInlinePolicyInput{
		TenantID: "t1",
		UserID:   9,
		Name:     "inline-tenant",
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	}))

	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"))
}

func TestIAMService_TenantUserPermissionBoundaryScopesDown(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantEditor] = entity.Role{ID: 4, Name: entity.RoleTenantEditor}
	state.memberships.items[membershipKey("t1", 9)] = entity.Membership{
		TenantID: "t1",
		UserID:   9,
		RoleID:   4,
		RoleName: entity.RoleTenantEditor,
		Status:   entity.MembershipStatusActive,
	}
	state.roleStatements[4] = []entity.PolicyStatement{
		{PolicyID: 15, PolicyName: "managed/tenant_editor", Effect: entity.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
	}
	state.policiesByName["tenant/orders-readonly-boundary"] = entity.Policy{ID: 22, Scope: entity.PolicyScopeTenant, Name: "tenant/orders-readonly-boundary"}
	state.policiesByID[22] = state.policiesByName["tenant/orders-readonly-boundary"]
	state.policyStatements[22] = []entity.PolicyStatement{
		{ID: 1, PolicyID: 22, PolicyName: "tenant/orders-readonly-boundary", Effect: entity.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
	}

	require.NoError(t, svc.PutTenantUserPermissionBoundary(context.Background(), "t1", 9, "tenant/orders-readonly-boundary"))
	require.ErrorIs(t, svc.RequirePermission(context.Background(), "t1", 9, "order:update"), entity.ErrPermissionDenied)
}
