package domain_test

import (
	"context"
	"fmt"
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

	out, err := svc.CreateTenant(context.Background(), 42, domain.CreateTenantCmd{Name: "Tenant One", Slug: "tenant-one"})
	require.NoError(t, err)
	require.NotNil(t, out)

	gotMembership, err := state.memberships.GetByTenantAndUser(context.Background(), out.ID, 42)
	require.NoError(t, err)
	require.Equal(t, domain.RoleTenantOwner, gotMembership.RoleName)
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
	require.ErrorIs(t, svc.RequirePermission(context.Background(), "t1", 9, "tenant:manage_members"), domain.ErrPermissionDenied)
}

func TestIAMService_RequirePlatformPermission(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.rolePermissions[7] = map[string]bool{"tenant:create": true}
	state.platformRoleIDs[11] = []uint64{7}

	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
	require.ErrorIs(t, svc.RequirePlatformPermission(context.Background(), 12, "tenant:create"), domain.ErrPermissionDenied)
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
	tenants         map[string]domain.Tenant
	roleByName      map[string]domain.Role
	rolePermissions map[uint64]map[string]bool
	platformRoleIDs map[uint][]uint64
	memberships     *membershipState
	invites         *inviteState
}

type membershipState struct {
	items map[string]domain.Membership
}

func (r *membershipState) GetByTenantAndUser(_ context.Context, tenantID string, userID uint) (*domain.Membership, error) {
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
	platformRepo := domainmocks.NewMockPlatformMembershipRepository(t)
	membershipRepo := domainmocks.NewMockMembershipRepository(t)
	inviteRepo := domainmocks.NewMockInviteRepository(t)

	state := &iamTestState{
		tenants:         map[string]domain.Tenant{},
		roleByName:      map[string]domain.Role{},
		rolePermissions: map[uint64]map[string]bool{},
		platformRoleIDs: map[uint][]uint64{},
		memberships:     &membershipState{items: map[string]domain.Membership{}},
		invites:         &inviteState{items: map[string]domain.TenantInvite{}, tokenIndex: map[string]string{}},
	}

	tenantRepo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tenant domain.Tenant) (*domain.Tenant, error) {
		copyTenant := tenant
		state.tenants[tenant.ID] = tenant
		return &copyTenant, nil
	}).Maybe()
	tenantRepo.EXPECT().GetByID(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tenantID string) (*domain.Tenant, error) {
		tenant, ok := state.tenants[tenantID]
		if !ok {
			return nil, domain.ErrTenantNotFound
		}
		copyTenant := tenant
		return &copyTenant, nil
	}).Maybe()

	roleRepo.EXPECT().GetByName(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, name string) (*domain.Role, error) {
		role, ok := state.roleByName[name]
		if !ok {
			return nil, domain.ErrRoleNotFound
		}
		copyRole := role
		return &copyRole, nil
	}).Maybe()
	roleRepo.EXPECT().RoleHasPermission(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, roleID uint64, permission string) (bool, error) {
		return state.rolePermissions[roleID][permission], nil
	}).Maybe()

	platformRepo.EXPECT().Upsert(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, userID uint, roleID uint64, status string) error {
		state.platformRoleIDs[userID] = append(state.platformRoleIDs[userID], roleID)
		return nil
	}).Maybe()
	platformRepo.EXPECT().ListRoleIDsByUser(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, userID uint) ([]uint64, error) {
		return append([]uint64(nil), state.platformRoleIDs[userID]...), nil
	}).Maybe()
	platformRepo.EXPECT().ListByUser(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, userID uint) ([]domain.PlatformMembership, error) {
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
	}).Maybe()
	platformRepo.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, userID uint, roleID uint64) error {
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
	}).Maybe()

	membershipRepo.EXPECT().Upsert(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, membership domain.Membership) error {
		state.memberships.items[membershipKey(membership.TenantID, membership.UserID)] = membership
		return nil
	}).Maybe()
	membershipRepo.EXPECT().GetByTenantAndUser(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tenantID string, userID uint) (*domain.Membership, error) {
		return state.memberships.GetByTenantAndUser(ctx, tenantID, userID)
	}).Maybe()
	membershipRepo.EXPECT().ListByTenant(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tenantID string) ([]domain.Membership, error) {
		out := make([]domain.Membership, 0)
		for _, item := range state.memberships.items {
			if item.TenantID == tenantID {
				out = append(out, item)
			}
		}
		return out, nil
	}).Maybe()
	membershipRepo.EXPECT().ListByUser(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, userID uint) ([]domain.Membership, error) {
		out := make([]domain.Membership, 0)
		for _, item := range state.memberships.items {
			if item.UserID == userID {
				out = append(out, item)
			}
		}
		return out, nil
	}).Maybe()
	membershipRepo.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tenantID string, userID uint) error {
		key := membershipKey(tenantID, userID)
		if _, ok := state.memberships.items[key]; !ok {
			return domain.ErrMembershipNotFound
		}
		delete(state.memberships.items, key)
		return nil
	}).Maybe()

	inviteRepo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, invite domain.TenantInvite) error {
		state.invites.items[invite.ID] = invite
		state.invites.tokenIndex[invite.TokenHash] = invite.ID
		return nil
	}).Maybe()
	inviteRepo.EXPECT().GetByID(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, inviteID string) (*domain.TenantInvite, error) {
		return state.invites.GetByID(ctx, inviteID)
	}).Maybe()
	inviteRepo.EXPECT().GetByTokenHash(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tokenHash string) (*domain.TenantInvite, error) {
		inviteID, ok := state.invites.tokenIndex[tokenHash]
		if !ok {
			return nil, domain.ErrInviteNotFound
		}
		return state.invites.GetByID(ctx, inviteID)
	}).Maybe()
	inviteRepo.EXPECT().ListByTenant(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tenantID string) ([]domain.TenantInvite, error) {
		out := make([]domain.TenantInvite, 0)
		for _, item := range state.invites.items {
			if item.TenantID == tenantID {
				out = append(out, item)
			}
		}
		return out, nil
	}).Maybe()
	inviteRepo.EXPECT().MarkAccepted(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, inviteID string, acceptedByUserID uint, acceptedAt time.Time) error {
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
	}).Maybe()
	inviteRepo.EXPECT().MarkRevoked(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, inviteID string, revokedAt time.Time) error {
		item, ok := state.invites.items[inviteID]
		if !ok {
			return domain.ErrInviteNotFound
		}
		item.Status = domain.InviteStatusRevoked
		item.RevokedAt = &revokedAt
		item.UpdatedAt = revokedAt
		state.invites.items[inviteID] = item
		return nil
	}).Maybe()

	return domain.NewIAMUsecase(tenantRepo, roleRepo, platformRepo, membershipRepo, inviteRepo), state
}

func membershipKey(tenantID string, userID uint) string {
	return fmt.Sprintf("%s:%d", tenantID, userID)
}
