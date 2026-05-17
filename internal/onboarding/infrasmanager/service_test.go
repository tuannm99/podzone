package infrasmanager

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
	coremocks "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core/mocks"
)

type connectionStoreState struct {
	upsertCalls     int
	softDeleteCalls int
	lastConn        *core.ConnectionInfo
	outbox          []core.OutboxMessage
	events          []core.ConnectionEvent
}

func newConnectionStoreMock(t *testing.T, state *connectionStoreState) *coremocks.MockConnectionStore {
	t.Helper()
	store := coremocks.NewMockConnectionStore(t)
	store.EXPECT().EnsureIndexes(mock.Anything).Return(nil).Maybe()
	store.EXPECT().Upsert(mock.Anything).RunAndReturn(func(info core.ConnectionInfo) error {
		state.upsertCalls++
		copyInfo := info
		state.lastConn = &copyInfo
		return nil
	}).Maybe()
	store.EXPECT().
		SoftDelete(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(tenantID string, infraType core.InfraType, name string) error {
			state.softDeleteCalls++
			return nil
		}).
		Maybe()
	store.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(tenantID string, infraType core.InfraType, name string) (*core.ConnectionInfo, error) {
			if state.lastConn == nil {
				return nil, nil
			}
			copyInfo := *state.lastConn
			return &copyInfo, nil
		}).
		Maybe()
	store.EXPECT().
		ListConnections(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]core.ConnectionInfo(nil), nil).
		Maybe()
	store.EXPECT().
		ListEvents(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]core.ConnectionEvent(nil), nil).
		Maybe()
	store.EXPECT().AppendEvent(mock.Anything).RunAndReturn(func(ev core.ConnectionEvent) error {
		state.events = append(state.events, ev)
		return nil
	}).Maybe()
	store.EXPECT().EnqueueOutbox(mock.Anything).RunAndReturn(func(msg core.OutboxMessage) error {
		state.outbox = append(state.outbox, msg)
		return nil
	}).Maybe()
	store.EXPECT().FindDueOutbox(mock.Anything).Return([]core.OutboxMessage(nil), nil).Maybe()
	store.EXPECT().MarkOutboxDone(mock.Anything).Return(nil).Maybe()
	store.EXPECT().MarkOutboxFailed(mock.Anything, mock.Anything).Return(nil).Maybe()
	return store
}

func TestManualUpsertConnection_RejectsSchemaModeMissingSchemaName(t *testing.T) {
	state := &connectionStoreState{}
	svc := NewService(newConnectionStoreMock(t, state))

	_, err := svc.ManualUpsertConnection("tenant-1", UpsertConnectionRequest{
		InfraType:   core.InfraPostgres,
		Endpoint:    "postgres://db",
		ClusterName: "pg-01",
		Mode:        "schema",
		DBName:      "backoffice",
	}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "schema_name is required")
	require.Zero(t, state.upsertCalls)
	require.Empty(t, state.outbox)
}

func TestManualUpsertConnection_PostgresEnqueuesPlacementPublish(t *testing.T) {
	state := &connectionStoreState{}
	svc := NewService(newConnectionStoreMock(t, state))

	resp, err := svc.ManualUpsertConnection("tenant-1", UpsertConnectionRequest{
		InfraType:   core.InfraPostgres,
		Name:        "default",
		Endpoint:    "postgres://db",
		ClusterName: "pg-01",
		Mode:        "schema",
		DBName:      "backoffice",
		SchemaName:  "t_tenant_1",
	}, map[string]string{"user": "tester"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, state.outbox, 2)

	require.Equal(t, "consul.publish", state.outbox[0].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/connections/postgres/default", state.outbox[0].Payload["key"])

	require.Equal(t, "consul.publish", state.outbox[1].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/placement", state.outbox[1].Payload["key"])
}

func TestDeleteConnection_PostgresEnqueuesPlacementDelete(t *testing.T) {
	state := &connectionStoreState{}
	svc := NewService(newConnectionStoreMock(t, state))

	corrID, err := svc.DeleteConnection("tenant-1", core.InfraPostgres, "default", map[string]string{"user": "tester"})
	require.NoError(t, err)
	require.NotEmpty(t, corrID)
	require.Equal(t, 1, state.softDeleteCalls)
	require.Len(t, state.outbox, 2)

	require.Equal(t, "consul.delete", state.outbox[0].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/placement", state.outbox[0].Payload["key"])

	require.Equal(t, "consul.delete", state.outbox[1].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/connections/postgres/default", state.outbox[1].Payload["key"])
}
