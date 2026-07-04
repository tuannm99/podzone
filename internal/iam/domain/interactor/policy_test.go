package interactor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

func TestIAMService_CreatePolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)

	policy, statements, err := svc.CreatePolicy(context.Background(), entity.CreatePolicyInput{
		Scope:       entity.PolicyScopeTenant,
		Name:        "tenant/orders_editor",
		Description: "Edit routed orders",
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, policy)
	require.Len(t, statements, 1)
	require.Equal(t, "tenant/orders_editor", policy.Name)
	require.Contains(t, state.policiesByName, policy.Name)
	require.Equal(t, "v1", policy.DefaultVersion)
}

func TestIAMService_CreateOrganizationPolicyRequiresOwner(t *testing.T) {
	t.Parallel()

	svc, _ := newIAMTestUsecase(t)
	input := entity.CreatePolicyInput{
		Scope: entity.PolicyScopeOrganization,
		Name:  "operations",
		Statements: []entity.PolicyStatement{{
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "organization:read",
			ResourcePattern: "*",
		}},
	}

	_, _, err := svc.CreatePolicy(context.Background(), input)
	require.ErrorIs(t, err, entity.ErrInvalidPolicyOwner)

	input.OrgID = "org-1"
	policy, _, err := svc.CreatePolicy(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, "org-1", policy.OrgID)
	require.Equal(t, entity.PolicyScopeOrganization, policy.Scope)
}

func TestIAMService_CreatePolicyVersionAndSetDefault(t *testing.T) {
	t.Parallel()

	svc, _ := newIAMTestUsecase(t)
	policy, _, err := svc.CreatePolicy(context.Background(), entity.CreatePolicyInput{
		Scope:       entity.PolicyScopeTenant,
		Name:        "tenant/orders_editor",
		Description: "Edit routed orders",
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)

	version, statements, err := svc.CreatePolicyVersion(context.Background(), entity.CreatePolicyVersionInput{
		Scope:      entity.PolicyScopeTenant,
		PolicyName: policy.Name,
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
		},
		SetAsDefault: true,
	})
	require.NoError(t, err)
	require.Equal(t, "v2", version.Version)
	require.True(t, version.IsDefault)
	require.Len(t, statements, 1)

	gotPolicy, gotStatements, err := svc.GetPolicy(context.Background(), tenantPolicyRef(policy.Name))
	require.NoError(t, err)
	require.Equal(t, "v2", gotPolicy.DefaultVersion)
	require.Len(t, gotStatements, 1)
	require.Equal(t, "order:read", gotStatements[0].ActionPattern)
}

func TestIAMService_DeleteNonDefaultPolicyVersion(t *testing.T) {
	t.Parallel()

	svc, _ := newIAMTestUsecase(t)
	policy, _, err := svc.CreatePolicy(context.Background(), entity.CreatePolicyInput{
		Scope: entity.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)
	_, _, err = svc.CreatePolicyVersion(context.Background(), entity.CreatePolicyVersionInput{
		Scope:      entity.PolicyScopeTenant,
		PolicyName: policy.Name,
		Statements: []entity.PolicyStatement{
			{Effect: entity.PolicyEffectAllow, ActionPattern: "order:read", ResourcePattern: "*"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeletePolicyVersion(context.Background(), tenantPolicyRef(policy.Name), "v2"))
	require.ErrorIs(
		t,
		svc.DeletePolicyVersion(context.Background(), tenantPolicyRef(policy.Name), "v1"),
		entity.ErrDefaultPolicyVersion,
	)
}

func TestIAMService_GetPolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["tenant/orders_editor"] = entity.Policy{
		ID:    9,
		Scope: entity.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
	}
	state.policiesByID[9] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[9] = []entity.PolicyStatement{
		{
			ID:              1,
			PolicyID:        9,
			PolicyName:      "tenant/orders_editor",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	policy, statements, err := svc.GetPolicy(context.Background(), tenantPolicyRef("tenant/orders_editor"))
	require.NoError(t, err)
	require.NotNil(t, policy)
	require.Len(t, statements, 1)
	require.Equal(t, uint64(9), policy.ID)
}

func TestIAMService_ListPolicyAttachments(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["tenant/orders_editor"] = entity.Policy{
		ID:    9,
		Scope: entity.PolicyScopeTenant,
		Name:  "tenant/orders_editor",
	}
	state.policiesByID[9] = state.policiesByName["tenant/orders_editor"]
	state.policyAttachments[9] = []entity.PolicyAttachment{
		{AttachmentType: "role", RoleID: 2, RoleName: "tenant_editor"},
		{AttachmentType: "group", Scope: entity.PolicyScopeTenant, TenantID: "t1", GroupID: 12, GroupName: "ops-team"},
	}

	page, err := svc.ListPolicyAttachments(
		context.Background(),
		tenantPolicyRef("tenant/orders_editor"),
		collection.Query{},
	)
	require.NoError(t, err)
	require.Len(t, page.Items, 2)
	require.Equal(t, "tenant_editor", page.Items[0].RoleName)
	require.Equal(t, uint64(12), page.Items[1].GroupID)
}

func TestIAMService_DeletePolicy(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["tenant/orders_editor"] = entity.Policy{
		ID:       9,
		Scope:    entity.PolicyScopeTenant,
		Name:     "tenant/orders_editor",
		IsSystem: false,
	}
	state.policiesByID[9] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[9] = []entity.PolicyStatement{
		{
			ID:              1,
			PolicyID:        9,
			PolicyName:      "tenant/orders_editor",
			Effect:          entity.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		},
	}

	require.NoError(t, svc.DeletePolicy(context.Background(), tenantPolicyRef("tenant/orders_editor")))
	_, ok := state.policiesByName["tenant/orders_editor"]
	require.False(t, ok)
}

func TestIAMService_DeletePolicy_BlocksSystem(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.policiesByName["managed/tenant_owner"] = entity.Policy{
		ID:       5,
		Scope:    entity.PolicyScopeTenant,
		Name:     "managed/tenant_owner",
		IsSystem: true,
	}
	state.policiesByID[5] = state.policiesByName["managed/tenant_owner"]

	err := svc.DeletePolicy(context.Background(), tenantPolicyRef("managed/tenant_owner"))
	require.ErrorIs(t, err, entity.ErrImmutablePolicy)
}

func tenantPolicyRef(name string) entity.PolicyRef {
	return entity.PolicyRef{Scope: entity.PolicyScopeTenant, Name: name}
}
