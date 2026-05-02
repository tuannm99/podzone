package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type tenantRepoFake struct {
	created *Tenant
	tenants map[string]Tenant
	err     error
}

func (r *tenantRepoFake) Create(ctx context.Context, tenant Tenant) (*Tenant, error) {
	if r.err != nil {
		return nil, r.err
	}
	copyTenant := tenant
	r.created = &copyTenant
	if r.tenants == nil {
		r.tenants = map[string]Tenant{}
	}
	r.tenants[tenant.ID] = tenant
	return &copyTenant, nil
}

func (r *tenantRepoFake) GetByID(ctx context.Context, tenantID string) (*Tenant, error) {
	t, ok := r.tenants[tenantID]
	if !ok {
		return nil, ErrTenantNotFound
	}
	copyTenant := t
	return &copyTenant, nil
}

type roleRepoFake struct {
	roles       map[string]Role
	permissions map[uint64]map[string]bool
}

func (r *roleRepoFake) GetByName(ctx context.Context, name string) (*Role, error) {
	role, ok := r.roles[name]
	if !ok {
		return nil, ErrRoleNotFound
	}
	copyRole := role
	return &copyRole, nil
}

func (r *roleRepoFake) RoleHasPermission(ctx context.Context, roleID uint64, permission string) (bool, error) {
	return r.permissions[roleID][permission], nil
}

type membershipRepoFake struct {
	items map[string]Membership
}

type inviteRepoFake struct {
	items      map[string]TenantInvite
	tokenIndex map[string]string
}

type platformMembershipRepoFake struct {
	roleIDs map[uint][]uint64
}

func (r *platformMembershipRepoFake) Upsert(ctx context.Context, userID uint, roleID uint64, status string) error {
	if r.roleIDs == nil {
		r.roleIDs = map[uint][]uint64{}
	}
	r.roleIDs[userID] = append(r.roleIDs[userID], roleID)
	return nil
}

func (r *platformMembershipRepoFake) ListRoleIDsByUser(ctx context.Context, userID uint) ([]uint64, error) {
	return r.roleIDs[userID], nil
}

func (r *platformMembershipRepoFake) ListByUser(ctx context.Context, userID uint) ([]PlatformMembership, error) {
	roleIDs := r.roleIDs[userID]
	out := make([]PlatformMembership, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		out = append(out, PlatformMembership{
			UserID:   userID,
			RoleID:   roleID,
			RoleName: fmt.Sprintf("role-%d", roleID),
			Status:   MembershipStatusActive,
		})
	}
	return out, nil
}

func (r *platformMembershipRepoFake) Delete(ctx context.Context, userID uint, roleID uint64) error {
	roleIDs := r.roleIDs[userID]
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
		return ErrMembershipNotFound
	}
	r.roleIDs[userID] = next
	return nil
}

func membershipKey(tenantID string, userID uint) string {
	return fmt.Sprintf("%s:%d", tenantID, userID)
}

func (r *membershipRepoFake) Upsert(ctx context.Context, membership Membership) error {
	if r.items == nil {
		r.items = map[string]Membership{}
	}
	r.items[membershipKey(membership.TenantID, membership.UserID)] = membership
	return nil
}

func (r *membershipRepoFake) GetByTenantAndUser(ctx context.Context, tenantID string, userID uint) (*Membership, error) {
	item, ok := r.items[membershipKey(tenantID, userID)]
	if !ok {
		return nil, ErrMembershipNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (r *membershipRepoFake) ListByTenant(ctx context.Context, tenantID string) ([]Membership, error) {
	out := make([]Membership, 0)
	for _, item := range r.items {
		if item.TenantID == tenantID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *membershipRepoFake) ListByUser(ctx context.Context, userID uint) ([]Membership, error) {
	out := make([]Membership, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *membershipRepoFake) Delete(ctx context.Context, tenantID string, userID uint) error {
	key := membershipKey(tenantID, userID)
	if _, ok := r.items[key]; !ok {
		return ErrMembershipNotFound
	}
	delete(r.items, key)
	return nil
}

func (r *inviteRepoFake) Create(ctx context.Context, invite TenantInvite) error {
	if r.items == nil {
		r.items = map[string]TenantInvite{}
	}
	if r.tokenIndex == nil {
		r.tokenIndex = map[string]string{}
	}
	r.items[invite.ID] = invite
	r.tokenIndex[invite.TokenHash] = invite.ID
	return nil
}

func (r *inviteRepoFake) GetByID(ctx context.Context, inviteID string) (*TenantInvite, error) {
	item, ok := r.items[inviteID]
	if !ok {
		return nil, ErrInviteNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (r *inviteRepoFake) GetByTokenHash(ctx context.Context, tokenHash string) (*TenantInvite, error) {
	inviteID, ok := r.tokenIndex[tokenHash]
	if !ok {
		return nil, ErrInviteNotFound
	}
	return r.GetByID(ctx, inviteID)
}

func (r *inviteRepoFake) ListByTenant(ctx context.Context, tenantID string) ([]TenantInvite, error) {
	out := make([]TenantInvite, 0)
	for _, item := range r.items {
		if item.TenantID == tenantID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *inviteRepoFake) MarkAccepted(ctx context.Context, inviteID string, acceptedByUserID uint, acceptedAt time.Time) error {
	item, ok := r.items[inviteID]
	if !ok {
		return ErrInviteNotFound
	}
	item.Status = InviteStatusAccepted
	item.AcceptedByUserID = &acceptedByUserID
	item.AcceptedAt = &acceptedAt
	item.UpdatedAt = acceptedAt
	r.items[inviteID] = item
	return nil
}

func (r *inviteRepoFake) MarkRevoked(ctx context.Context, inviteID string, revokedAt time.Time) error {
	item, ok := r.items[inviteID]
	if !ok {
		return ErrInviteNotFound
	}
	item.Status = InviteStatusRevoked
	item.RevokedAt = &revokedAt
	item.UpdatedAt = revokedAt
	r.items[inviteID] = item
	return nil
}

func TestIAMService_CreateTenant_AssignsOwnerRole(t *testing.T) {
	tenants := &tenantRepoFake{}
	roles := &roleRepoFake{
		roles: map[string]Role{
			RoleTenantOwner: {ID: 1, Name: RoleTenantOwner},
		},
	}
	platformRoles := &platformMembershipRepoFake{}
	memberships := &membershipRepoFake{}
	invites := &inviteRepoFake{}
	svc := NewIAMUsecase(tenants, roles, platformRoles, memberships, invites)

	out, err := svc.CreateTenant(context.Background(), 42, CreateTenantCmd{Name: "Tenant One", Slug: "tenant-one"})
	require.NoError(t, err)
	require.NotNil(t, out)
	gotMembership, err := memberships.GetByTenantAndUser(context.Background(), out.ID, 42)
	require.NoError(t, err)
	require.Equal(t, RoleTenantOwner, gotMembership.RoleName)
}

func TestIAMService_RequirePermission(t *testing.T) {
	tenant := Tenant{ID: "t1", Name: "Tenant", Slug: "tenant"}
	tenants := &tenantRepoFake{tenants: map[string]Tenant{tenant.ID: tenant}}
	roles := &roleRepoFake{
		roles: map[string]Role{
			RoleTenantEditor: {ID: 2, Name: RoleTenantEditor},
		},
		permissions: map[uint64]map[string]bool{
			2: {"store:update": true},
		},
	}
	platformRoles := &platformMembershipRepoFake{}
	memberships := &membershipRepoFake{
		items: map[string]Membership{
			membershipKey("t1", 9): {
				TenantID: "t1",
				UserID:   9,
				RoleID:   2,
				RoleName: RoleTenantEditor,
				Status:   MembershipStatusActive,
			},
		},
	}
	invites := &inviteRepoFake{}
	svc := NewIAMUsecase(tenants, roles, platformRoles, memberships, invites)

	require.NoError(t, svc.RequirePermission(context.Background(), "t1", 9, "store:update"))
	require.ErrorIs(t, svc.RequirePermission(context.Background(), "t1", 9, "tenant:manage_members"), ErrPermissionDenied)
}

func TestIAMService_RequirePlatformPermission(t *testing.T) {
	tenants := &tenantRepoFake{}
	roles := &roleRepoFake{
		permissions: map[uint64]map[string]bool{
			7: {"tenant:create": true},
		},
	}
	platformRoles := &platformMembershipRepoFake{
		roleIDs: map[uint][]uint64{
			11: {7},
		},
	}
	memberships := &membershipRepoFake{}
	invites := &inviteRepoFake{}
	svc := NewIAMUsecase(tenants, roles, platformRoles, memberships, invites)

	require.NoError(t, svc.RequirePlatformPermission(context.Background(), 11, "tenant:create"))
	require.ErrorIs(t, svc.RequirePlatformPermission(context.Background(), 12, "tenant:create"), ErrPermissionDenied)
}

func TestIAMService_AddAndRemovePlatformRole(t *testing.T) {
	tenants := &tenantRepoFake{}
	roles := &roleRepoFake{
		roles: map[string]Role{
			RolePlatformAdmin: {ID: 8, Name: RolePlatformAdmin},
		},
	}
	platformRoles := &platformMembershipRepoFake{}
	memberships := &membershipRepoFake{}
	invites := &inviteRepoFake{}
	svc := NewIAMUsecase(tenants, roles, platformRoles, memberships, invites)

	require.NoError(t, svc.AddPlatformRole(context.Background(), 21, RolePlatformAdmin))
	items, err := svc.ListPlatformRoles(context.Background(), 21)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint64(8), items[0].RoleID)

	require.NoError(t, svc.RemovePlatformRole(context.Background(), 21, RolePlatformAdmin))
	items, err = svc.ListPlatformRoles(context.Background(), 21)
	require.NoError(t, err)
	require.Len(t, items, 0)
}

func TestIAMService_CreateAndAcceptInvite(t *testing.T) {
	tenant := Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	tenants := &tenantRepoFake{tenants: map[string]Tenant{tenant.ID: tenant}}
	roles := &roleRepoFake{
		roles: map[string]Role{
			RoleTenantViewer: {ID: 3, Name: RoleTenantViewer},
		},
	}
	platformRoles := &platformMembershipRepoFake{}
	memberships := &membershipRepoFake{}
	invites := &inviteRepoFake{}
	svc := NewIAMUsecase(tenants, roles, platformRoles, memberships, invites)

	invite, rawToken, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", RoleTenantViewer, 7)
	require.NoError(t, err)
	require.NotEmpty(t, rawToken)
	require.Equal(t, InviteStatusPending, invite.Status)
	require.NotEmpty(t, invite.TokenHash)

	membership, err := svc.AcceptInvite(context.Background(), rawToken, 11, "neo@mx.io")
	require.NoError(t, err)
	require.Equal(t, tenant.ID, membership.TenantID)
	require.Equal(t, uint(11), membership.UserID)
	require.Equal(t, RoleTenantViewer, membership.RoleName)

	storedInvite, err := invites.GetByID(context.Background(), invite.ID)
	require.NoError(t, err)
	require.Equal(t, InviteStatusAccepted, storedInvite.Status)
	require.NotNil(t, storedInvite.AcceptedByUserID)
	require.Equal(t, uint(11), *storedInvite.AcceptedByUserID)
}

func TestIAMService_AcceptInvite_EmailMismatch(t *testing.T) {
	tenant := Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	tenants := &tenantRepoFake{tenants: map[string]Tenant{tenant.ID: tenant}}
	roles := &roleRepoFake{
		roles: map[string]Role{
			RoleTenantViewer: {ID: 3, Name: RoleTenantViewer},
		},
	}
	platformRoles := &platformMembershipRepoFake{}
	memberships := &membershipRepoFake{}
	invites := &inviteRepoFake{}
	svc := NewIAMUsecase(tenants, roles, platformRoles, memberships, invites)

	_, rawToken, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", RoleTenantViewer, 7)
	require.NoError(t, err)

	_, err = svc.AcceptInvite(context.Background(), rawToken, 11, "trinity@mx.io")
	require.ErrorIs(t, err, ErrInviteEmailMismatch)
}

func TestIAMService_RevokeInvite(t *testing.T) {
	tenant := Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	tenants := &tenantRepoFake{tenants: map[string]Tenant{tenant.ID: tenant}}
	roles := &roleRepoFake{
		roles: map[string]Role{
			RoleTenantViewer: {ID: 3, Name: RoleTenantViewer},
		},
	}
	platformRoles := &platformMembershipRepoFake{}
	memberships := &membershipRepoFake{}
	invites := &inviteRepoFake{}
	svc := NewIAMUsecase(tenants, roles, platformRoles, memberships, invites)

	invite, _, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", RoleTenantViewer, 7)
	require.NoError(t, err)

	require.NoError(t, svc.RevokeInvite(context.Background(), invite.ID))

	storedInvite, err := invites.GetByID(context.Background(), invite.ID)
	require.NoError(t, err)
	require.Equal(t, InviteStatusRevoked, storedInvite.Status)
	require.NotNil(t, storedInvite.RevokedAt)
}
