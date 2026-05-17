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
		{PolicyID: 10, PolicyName: "managed/tenant_editor", Effect: domain.PolicyEffectAllow, ActionPattern: "store:update", ResourcePattern: "*"},
	}
	state.tenantDirect[membershipKey("t1", 9)] = []domain.PolicyStatement{
		{PolicyID: 11, PolicyName: "inline/deny-store-update", Effect: domain.PolicyEffectDeny, ActionPattern: "store:update", ResourcePattern: "*"},
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
		{PolicyID: 20, PolicyName: "inline/platform-direct", Effect: domain.PolicyEffectAllow, ActionPattern: "tenant:create", ResourcePattern: "*"},
	}

	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
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
	state.policiesByName["managed/platform_owner"] = domain.Policy{ID: 5, Scope: domain.PolicyScopePlatform, Name: "managed/platform_owner"}
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
	state.policiesByName["tenant/orders_editor"] = domain.Policy{ID: 9, Scope: domain.PolicyScopeTenant, Name: "tenant/orders_editor"}
	state.policiesByID[9] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[9] = []domain.PolicyStatement{
		{ID: 1, PolicyID: 9, PolicyName: "tenant/orders_editor", Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
	}

	policy, statements, err := svc.GetPolicy(context.Background(), "tenant/orders_editor")
	require.NoError(t, err)
	require.NotNil(t, policy)
	require.Len(t, statements, 1)
	require.Equal(t, uint64(9), policy.ID)
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
		{ID: 1, PolicyID: 9, PolicyName: "tenant/orders_editor", Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
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
	state.policiesByName["tenant/orders_editor"] = domain.Policy{ID: 15, Scope: domain.PolicyScopeTenant, Name: "tenant/orders_editor"}
	state.policiesByID[15] = state.policiesByName["tenant/orders_editor"]
	state.policyStatements[15] = []domain.PolicyStatement{
		{ID: 1, PolicyID: 15, PolicyName: "tenant/orders_editor", Effect: domain.PolicyEffectAllow, ActionPattern: "order:update", ResourcePattern: "*"},
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
	tenants          map[string]domain.Tenant
	roleByName       map[string]domain.Role
	rolePermissions  map[uint64]map[string]bool
	policiesByName   map[string]domain.Policy
	policiesByID     map[uint64]domain.Policy
	policyStatements map[uint64][]domain.PolicyStatement
	nextPolicyID     uint64
	groupsByID       map[uint64]domain.Group
	nextGroupID      uint64
	roleStatements   map[uint64][]domain.PolicyStatement
	platformDirect   map[uint][]domain.PolicyStatement
	tenantDirect     map[string][]domain.PolicyStatement
	platformRoleIDs  map[uint][]uint64
	memberships      *membershipState
	invites          *inviteState
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
	platformRepo := domainmocks.NewMockPlatformMembershipRepository(t)
	membershipRepo := domainmocks.NewMockMembershipRepository(t)
	inviteRepo := domainmocks.NewMockInviteRepository(t)

	state := &iamTestState{
		tenants:          map[string]domain.Tenant{},
		roleByName:       map[string]domain.Role{},
		rolePermissions:  map[uint64]map[string]bool{},
		policiesByName:   map[string]domain.Policy{},
		policiesByID:     map[uint64]domain.Policy{},
		policyStatements: map[uint64][]domain.PolicyStatement{},
		nextPolicyID:     100,
		groupsByID:       map[uint64]domain.Group{},
		nextGroupID:      200,
		roleStatements:   map[uint64][]domain.PolicyStatement{},
		platformDirect:   map[uint][]domain.PolicyStatement{},
		tenantDirect:     map[string][]domain.PolicyStatement{},
		platformRoleIDs:  map[uint][]uint64{},
		memberships:      &membershipState{items: map[string]domain.Membership{}},
		invites:          &inviteState{items: map[string]domain.TenantInvite{}, tokenIndex: map[string]string{}},
	}

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

	policyRepo.EXPECT().
		CreatePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policy domain.Policy, statements []domain.PolicyStatement) (*domain.Policy, []domain.PolicyStatement, error) {
			policy.ID = state.nextPolicyID
			state.nextPolicyID++
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
			return &policy, outStatements, nil
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
			return append([]domain.PolicyStatement(nil), state.platformDirect[userID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListTenantUserStatements(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]domain.PolicyStatement, error) {
			return append([]domain.PolicyStatement(nil), state.tenantDirect[membershipKey(tenantID, userID)]...), nil
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
			state.tenantDirect[fmt.Sprintf("group-policy:%d", groupID)] = append(
				state.tenantDirect[fmt.Sprintf("group-policy:%d", groupID)],
				state.policyStatements[policyID]...,
			)
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

	return domain.NewIAMUsecase(tenantRepo, roleRepo, policyRepo, groupRepo, platformRepo, membershipRepo, inviteRepo), state
}

func membershipKey(tenantID string, userID uint) string {
	return fmt.Sprintf("%s:%d", tenantID, userID)
}
