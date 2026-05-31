package interactor_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/inputport"
	iaminteractor "github.com/tuannm99/podzone/internal/iam/domain/interactor"
	outputportmocks "github.com/tuannm99/podzone/internal/iam/domain/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type iamTestState struct {
	tenants                 map[string]entity.Tenant
	roleByName              map[string]entity.Role
	roleTrustStatements     map[uint64][]entity.RoleTrustStatement
	rolePermissions         map[uint64]map[string]bool
	policiesByName          map[string]entity.Policy
	policiesByID            map[uint64]entity.Policy
	policyStatements        map[uint64][]entity.PolicyStatement
	policyVersions          map[uint64][]entity.PolicyVersion
	policyVersionStatements map[string][]entity.PolicyStatement
	policyAttachments       map[uint64][]entity.PolicyAttachment
	nextPolicyID            uint64
	groupsByID              map[uint64]entity.Group
	groupInlinePolicies     map[uint64]map[string]entity.GroupInlinePolicy
	nextGroupID             uint64
	roleStatements          map[uint64][]entity.PolicyStatement
	roleBoundary            map[uint64]*entity.RolePermissionBoundary
	roleBoundaryStmts       map[uint64][]entity.PolicyStatement
	platformDirect          map[uint][]entity.PolicyStatement
	tenantDirect            map[string][]entity.PolicyStatement
	platformBoundary        map[uint]*entity.PermissionBoundary
	platformBoundaryStmts   map[uint][]entity.PolicyStatement
	platformInlinePolicies  map[uint]map[string]entity.UserInlinePolicy
	tenantBoundary          map[string]*entity.PermissionBoundary
	tenantBoundaryStmts     map[string][]entity.PolicyStatement
	tenantInlinePolicies    map[string]map[string]entity.UserInlinePolicy
	platformRoleIDs         map[uint][]uint64
	memberships             *membershipState
	invites                 *inviteState
	outboxRecords           []messaging.OutboxRecord
}

type membershipState struct {
	items map[string]entity.Membership
}

func (r *membershipState) GetByTenantAndUser(
	_ context.Context,
	tenantID string,
	userID uint,
) (*entity.Membership, error) {
	item, ok := r.items[membershipKey(tenantID, userID)]
	if !ok {
		return nil, entity.ErrMembershipNotFound
	}
	copyItem := item
	return &copyItem, nil
}

type inviteState struct {
	items      map[string]entity.TenantInvite
	tokenIndex map[string]string
}

func (s *iamTestState) roleByID(roleID uint64) entity.Role {
	for _, role := range s.roleByName {
		if role.ID == roleID {
			return role
		}
	}
	return entity.Role{ID: roleID}
}

func (r *inviteState) GetByID(_ context.Context, inviteID string) (*entity.TenantInvite, error) {
	item, ok := r.items[inviteID]
	if !ok {
		return nil, entity.ErrInviteNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func newIAMTestUsecase(t *testing.T) (inputport.IAMUsecase, *iamTestState) {
	t.Helper()

	tenantRepo := outputportmocks.NewMockTenantRepository(t)
	roleRepo := outputportmocks.NewMockRoleRepository(t)
	policyRepo := outputportmocks.NewMockPolicyRepository(t)
	groupRepo := outputportmocks.NewMockGroupRepository(t)
	orgRepo := outputportmocks.NewMockOrganizationRepository(t)
	platformRepo := outputportmocks.NewMockPlatformMembershipRepository(t)
	membershipRepo := outputportmocks.NewMockMembershipRepository(t)
	inviteRepo := outputportmocks.NewMockInviteRepository(t)
	outboxRepo := outputportmocks.NewMockOutboxRepository(t)

	state := &iamTestState{
		tenants:                 map[string]entity.Tenant{},
		roleByName:              map[string]entity.Role{},
		roleTrustStatements:     map[uint64][]entity.RoleTrustStatement{},
		rolePermissions:         map[uint64]map[string]bool{},
		policiesByName:          map[string]entity.Policy{},
		policiesByID:            map[uint64]entity.Policy{},
		policyStatements:        map[uint64][]entity.PolicyStatement{},
		policyVersions:          map[uint64][]entity.PolicyVersion{},
		policyVersionStatements: map[string][]entity.PolicyStatement{},
		policyAttachments:       map[uint64][]entity.PolicyAttachment{},
		nextPolicyID:            100,
		groupsByID:              map[uint64]entity.Group{},
		groupInlinePolicies:     map[uint64]map[string]entity.GroupInlinePolicy{},
		nextGroupID:             200,
		roleStatements:          map[uint64][]entity.PolicyStatement{},
		roleBoundary:            map[uint64]*entity.RolePermissionBoundary{},
		roleBoundaryStmts:       map[uint64][]entity.PolicyStatement{},
		platformDirect:          map[uint][]entity.PolicyStatement{},
		tenantDirect:            map[string][]entity.PolicyStatement{},
		platformBoundary:        map[uint]*entity.PermissionBoundary{},
		platformBoundaryStmts:   map[uint][]entity.PolicyStatement{},
		platformInlinePolicies:  map[uint]map[string]entity.UserInlinePolicy{},
		tenantBoundary:          map[string]*entity.PermissionBoundary{},
		tenantBoundaryStmts:     map[string][]entity.PolicyStatement{},
		tenantInlinePolicies:    map[string]map[string]entity.UserInlinePolicy{},
		platformRoleIDs:         map[uint][]uint64{},
		memberships:             &membershipState{items: map[string]entity.Membership{}},
		invites:                 &inviteState{items: map[string]entity.TenantInvite{}, tokenIndex: map[string]string{}},
	}

	outboxRepo.EXPECT().
		Append(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error {
			state.outboxRecords = append(state.outboxRecords, record)
			return nil
		}).
		Maybe()

	orgRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, org entity.Organization) (*entity.Organization, error) {
			copyOrg := org
			return &copyOrg, nil
		}).
		Maybe()
	orgRepo.EXPECT().
		List(mock.Anything).
		RunAndReturn(func(ctx context.Context) ([]entity.Organization, error) { return nil, nil }).
		Maybe()
	orgRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, orgID string) (*entity.Organization, error) {
			if orgID == "" {
				return nil, entity.ErrOrganizationNotFound
			}
			return &entity.Organization{ID: orgID, Slug: orgID, Name: orgID}, nil
		}).
		Maybe()
	orgRepo.EXPECT().AttachServiceControlPolicy(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	orgRepo.EXPECT().DetachServiceControlPolicy(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	orgRepo.EXPECT().
		ListServiceControlPolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, orgID string) ([]entity.Policy, error) { return nil, nil }).
		Maybe()
	orgRepo.EXPECT().
		ListServiceControlPolicyStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, orgID string) ([]entity.PolicyStatement, error) { return nil, nil }).
		Maybe()

	tenantRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenant entity.Tenant) (*entity.Tenant, error) {
			copyTenant := tenant
			state.tenants[tenant.ID] = tenant
			return &copyTenant, nil
		}).
		Maybe()
	tenantRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) (*entity.Tenant, error) {
			tenant, ok := state.tenants[tenantID]
			if !ok {
				return nil, entity.ErrTenantNotFound
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
		}).
		Maybe()
	tenantRepo.EXPECT().
		DetachOrganization(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) error {
			tenant := state.tenants[tenantID]
			tenant.OrgID = ""
			state.tenants[tenantID] = tenant
			return nil
		}).
		Maybe()

	roleRepo.EXPECT().
		GetByName(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, name string) (*entity.Role, error) {
			role, ok := state.roleByName[name]
			if !ok {
				return nil, entity.ErrRoleNotFound
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
		RunAndReturn(func(ctx context.Context, roleID uint64, statements []entity.RoleTrustStatement) error {
			state.roleTrustStatements[roleID] = append([]entity.RoleTrustStatement(nil), statements...)
			return nil
		}).
		Maybe()
	roleRepo.EXPECT().
		GetTrustPolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) ([]entity.RoleTrustStatement, error) {
			return append([]entity.RoleTrustStatement(nil), state.roleTrustStatements[roleID]...), nil
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
			state.roleBoundary[roleID] = &entity.RolePermissionBoundary{
				RoleID:     roleID,
				RoleName:   state.roleByID(roleID).Name,
				PolicyID:   policyID,
				PolicyName: policy.Name,
				CreatedAt:  time.Now().UTC(),
			}
			state.roleBoundaryStmts[roleID] = append([]entity.PolicyStatement(nil), state.policyStatements[policyID]...)
			return nil
		}).
		Maybe()
	roleRepo.EXPECT().
		GetPermissionBoundary(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) (*entity.RolePermissionBoundary, error) {
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
		RunAndReturn(func(ctx context.Context, roleID uint64) ([]entity.PolicyStatement, error) {
			return append([]entity.PolicyStatement(nil), state.roleBoundaryStmts[roleID]...), nil
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

	// Policy, group, principal, membership, and invite mocks are kept stateful here so
	// slice-specific tests can stay focused on business rules rather than harness setup.
	configurePolicyRepoMocks(policyRepo, state)
	configureGroupRepoMocks(groupRepo, state)
	configurePlatformRepoMocks(platformRepo, state)
	configureMembershipRepoMocks(membershipRepo, state)
	configureInviteRepoMocks(inviteRepo, state)

	return iaminteractor.NewIAMUsecase(
		tenantRepo,
		roleRepo,
		policyRepo,
		groupRepo,
		orgRepo,
		platformRepo,
		membershipRepo,
		inviteRepo,
		outboxRepo,
	), state
}

func membershipKey(tenantID string, userID uint) string {
	return fmt.Sprintf("%s:%d", tenantID, userID)
}
