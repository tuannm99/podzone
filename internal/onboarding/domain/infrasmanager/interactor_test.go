package infrasmanager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	coremocks "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport/mocks"
)

type connectionStoreState struct {
	upsertCalls     int
	softDeleteCalls int
	lastConn        *entity.ConnectionInfo
	outbox          []entity.OutboxMessage
	events          []entity.ConnectionEvent
}

func newConnectionStoreMock(t *testing.T, state *connectionStoreState) *coremocks.MockConnectionStore {
	t.Helper()
	store := coremocks.NewMockConnectionStore(t)
	store.EXPECT().EnsureIndexes(mock.Anything).Return(nil).Maybe()
	store.EXPECT().
		Upsert(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, info entity.ConnectionInfo) error {
			state.upsertCalls++
			copyInfo := info
			state.lastConn = &copyInfo
			return nil
		}).
		Maybe()
	store.EXPECT().
		SoftDelete(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, tenantID string, infraType entity.InfraType, name string) error {
			state.softDeleteCalls++
			return nil
		}).
		Maybe()
	store.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context,
			tenantID string,
			infraType entity.InfraType,
			name string,
		) (*entity.ConnectionInfo, error) {
			if state.lastConn == nil {
				return nil, nil
			}
			copyInfo := *state.lastConn
			return &copyInfo, nil
		}).
		Maybe()
	store.EXPECT().
		ListConnections(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]entity.ConnectionInfo(nil), nil).
		Maybe()
	store.EXPECT().
		ListEvents(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]entity.ConnectionEvent(nil), nil).
		Maybe()
	store.EXPECT().
		AppendEvent(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, ev entity.ConnectionEvent) error {
			state.events = append(state.events, ev)
			return nil
		}).
		Maybe()
	store.EXPECT().
		EnqueueOutbox(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, msg entity.OutboxMessage) error {
			state.outbox = append(state.outbox, msg)
			return nil
		}).
		Maybe()
	store.EXPECT().FindDueOutbox(mock.Anything, mock.Anything).Return([]entity.OutboxMessage(nil), nil).Maybe()
	store.EXPECT().MarkOutboxDone(mock.Anything, mock.Anything).Return(nil).Maybe()
	store.EXPECT().MarkOutboxFailed(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return store
}

func TestManualUpsertConnection_RejectsSchemaModeMissingSchemaName(t *testing.T) {
	state := &connectionStoreState{}
	svc := NewInteractor(newConnectionStoreMock(t, state))

	_, err := svc.ManualUpsertConnection(context.Background(), "tenant-1", inputport.UpsertConnectionRequest{
		InfraType:   entity.InfraPostgres,
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
	svc := NewInteractor(newConnectionStoreMock(t, state))

	resp, err := svc.ManualUpsertConnection(context.Background(), "tenant-1", inputport.UpsertConnectionRequest{
		InfraType:   entity.InfraPostgres,
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
	svc := NewInteractor(newConnectionStoreMock(t, state))

	corrID, err := svc.DeleteConnection(
		context.Background(),
		"tenant-1",
		entity.InfraPostgres,
		"default",
		map[string]string{"user": "tester"},
	)
	require.NoError(t, err)
	require.NotEmpty(t, corrID)
	require.Equal(t, 1, state.softDeleteCalls)
	require.Len(t, state.outbox, 2)

	require.Equal(t, "consul.delete", state.outbox[0].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/placement", state.outbox[0].Payload["key"])

	require.Equal(t, "consul.delete", state.outbox[1].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/connections/postgres/default", state.outbox[1].Payload["key"])
}

func TestEnsurePlacementRoute_RepairsMissingRouteFromAllocation(t *testing.T) {
	placements := coremocks.NewMockPlacementRepository(t)
	writer := coremocks.NewMockPlacementRouteWriter(t)
	svc := &Interactor{
		placements:  placements,
		routeWriter: writer,
	}
	allocation := entity.PlacementAllocation{
		ID:          "allocation-1",
		TenantID:    "tenant-1",
		StoreID:     "store-1",
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
		Status:      "ready",
	}

	placements.EXPECT().GetPlacementAllocation(mock.Anything, "tenant-1", "store-1").Return(&allocation, nil)
	writer.EXPECT().PublishPlacementRoute(mock.Anything, "tenant-1", allocation).Return(nil)

	ready, err := svc.EnsurePlacementRoute(context.Background(), "tenant-1", "store-1")
	require.NoError(t, err)
	require.True(t, ready)
}
