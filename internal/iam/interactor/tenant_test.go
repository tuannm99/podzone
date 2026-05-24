package interactor_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	entity "github.com/tuannm99/podzone/internal/iam/entity"
)

func TestIAMService_CreateTenant_AssignsOwnerRole(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	state.roleByName[entity.RoleTenantOwner] = entity.Role{ID: 1, Name: entity.RoleTenantOwner}

	out, err := svc.CreateTenant(
		context.Background(),
		42,
		entity.CreateTenantCmd{Name: "Tenant One", Slug: "tenant-one"},
	)
	require.NoError(t, err)
	require.NotNil(t, out)

	gotMembership, err := state.memberships.GetByTenantAndUser(context.Background(), out.ID, 42)
	require.NoError(t, err)
	require.Equal(t, entity.RoleTenantOwner, gotMembership.RoleName)
	require.Len(t, state.outboxRecords, 1)
	require.Equal(t, "podzone.iam.events", state.outboxRecords[0].Topic)
	require.Equal(t, "tenant.created", state.outboxRecords[0].Envelope.Type)
	require.Equal(t, out.ID, state.outboxRecords[0].Envelope.TenantID)
}

func TestIAMService_CreateAndAcceptInvite(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 3, Name: entity.RoleTenantViewer}

	invite, rawToken, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", entity.RoleTenantViewer, 7)
	require.NoError(t, err)
	require.NotEmpty(t, rawToken)
	require.Equal(t, entity.InviteStatusPending, invite.Status)
	require.NotEmpty(t, invite.TokenHash)

	membership, err := svc.AcceptInvite(context.Background(), rawToken, 11, "neo@mx.io")
	require.NoError(t, err)
	require.Equal(t, tenant.ID, membership.TenantID)
	require.Equal(t, uint(11), membership.UserID)
	require.Equal(t, entity.RoleTenantViewer, membership.RoleName)

	storedInvite, err := state.invites.GetByID(context.Background(), invite.ID)
	require.NoError(t, err)
	require.Equal(t, entity.InviteStatusAccepted, storedInvite.Status)
	require.NotNil(t, storedInvite.AcceptedByUserID)
	require.Equal(t, uint(11), *storedInvite.AcceptedByUserID)
}

func TestIAMService_AddMember_AppendsOutboxEvent(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 3, Name: entity.RoleTenantViewer}

	require.NoError(t, svc.AddMember(context.Background(), tenant.ID, 11, entity.RoleTenantViewer))
	require.Len(t, state.outboxRecords, 1)
	require.Equal(t, "tenant.member.added", state.outboxRecords[0].Envelope.Type)
	require.Equal(t, tenant.ID, state.outboxRecords[0].Envelope.TenantID)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(state.outboxRecords[0].Envelope.Payload, &payload))
	require.Equal(t, tenant.ID, payload["tenant_id"])
	require.Equal(t, float64(11), payload["user_id"])
	require.Equal(t, entity.RoleTenantViewer, payload["role_name"])
}

func TestIAMService_AcceptInvite_EmailMismatch(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 3, Name: entity.RoleTenantViewer}

	_, rawToken, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", entity.RoleTenantViewer, 7)
	require.NoError(t, err)

	_, err = svc.AcceptInvite(context.Background(), rawToken, 11, "trinity@mx.io")
	require.ErrorIs(t, err, entity.ErrInviteEmailMismatch)
}

func TestIAMService_RevokeInvite(t *testing.T) {
	t.Parallel()

	svc, state := newIAMTestUsecase(t)
	tenant := entity.Tenant{ID: "tenant-1", Name: "Tenant One", Slug: "tenant-one"}
	state.tenants[tenant.ID] = tenant
	state.roleByName[entity.RoleTenantViewer] = entity.Role{ID: 3, Name: entity.RoleTenantViewer}

	invite, _, err := svc.CreateInvite(context.Background(), tenant.ID, "neo@mx.io", entity.RoleTenantViewer, 7)
	require.NoError(t, err)

	require.NoError(t, svc.RevokeInvite(context.Background(), invite.ID))

	storedInvite, err := state.invites.GetByID(context.Background(), invite.ID)
	require.NoError(t, err)
	require.Equal(t, entity.InviteStatusRevoked, storedInvite.Status)
	require.NotNil(t, storedInvite.RevokedAt)
}
