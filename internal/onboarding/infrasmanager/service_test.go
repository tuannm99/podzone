package infrasmanager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
)

type fakeConnectionStore struct {
	upsertCalls     int
	softDeleteCalls int
	lastConn        *core.ConnectionInfo
	outbox          []core.OutboxMessage
	events          []core.ConnectionEvent
}

func (f *fakeConnectionStore) EnsureIndexes(_ context.Context) error {
	return nil
}

func (f *fakeConnectionStore) Upsert(info core.ConnectionInfo) error {
	f.upsertCalls++
	copyInfo := info
	f.lastConn = &copyInfo
	return nil
}

func (f *fakeConnectionStore) SoftDelete(tenantID string, infraType core.InfraType, name string) error {
	f.softDeleteCalls++
	return nil
}

func (f *fakeConnectionStore) Get(
	tenantID string,
	infraType core.InfraType,
	name string,
) (*core.ConnectionInfo, error) {
	if f.lastConn == nil {
		return nil, nil
	}
	copyInfo := *f.lastConn
	return &copyInfo, nil
}

func (f *fakeConnectionStore) ListConnections(
	tenantID string,
	infraType core.InfraType,
	includeDeleted bool,
	limit, offset int,
) ([]core.ConnectionInfo, error) {
	return nil, nil
}

func (f *fakeConnectionStore) ListEvents(
	tenantID string,
	infraType core.InfraType,
	name string,
	correlationID string,
	limit, offset int,
) ([]core.ConnectionEvent, error) {
	return nil, nil
}

func (f *fakeConnectionStore) AppendEvent(ev core.ConnectionEvent) error {
	f.events = append(f.events, ev)
	return nil
}

func (f *fakeConnectionStore) EnqueueOutbox(msg core.OutboxMessage) error {
	f.outbox = append(f.outbox, msg)
	return nil
}

func (f *fakeConnectionStore) FindDueOutbox(limit int) ([]core.OutboxMessage, error) {
	return nil, nil
}

func (f *fakeConnectionStore) MarkOutboxDone(eventID string) error {
	return nil
}

func (f *fakeConnectionStore) MarkOutboxFailed(eventID string, nextRetry time.Time) error {
	return nil
}

func TestManualUpsertConnection_RejectsSchemaModeMissingSchemaName(t *testing.T) {
	st := &fakeConnectionStore{}
	svc := NewService(st)

	_, err := svc.ManualUpsertConnection("tenant-1", UpsertConnectionRequest{
		InfraType:   core.InfraPostgres,
		Endpoint:    "postgres://db",
		ClusterName: "pg-01",
		Mode:        "schema",
		DBName:      "backoffice",
	}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "schema_name is required")
	require.Zero(t, st.upsertCalls)
	require.Empty(t, st.outbox)
}

func TestManualUpsertConnection_PostgresEnqueuesPlacementPublish(t *testing.T) {
	st := &fakeConnectionStore{}
	svc := NewService(st)

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
	require.Len(t, st.outbox, 2)

	require.Equal(t, "consul.publish", st.outbox[0].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/connections/postgres/default", st.outbox[0].Payload["key"])

	require.Equal(t, "consul.publish", st.outbox[1].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/placement", st.outbox[1].Payload["key"])
}

func TestDeleteConnection_PostgresEnqueuesPlacementDelete(t *testing.T) {
	st := &fakeConnectionStore{}
	svc := NewService(st)

	corrID, err := svc.DeleteConnection("tenant-1", core.InfraPostgres, "default", map[string]string{"user": "tester"})
	require.NoError(t, err)
	require.NotEmpty(t, corrID)
	require.Equal(t, 1, st.softDeleteCalls)
	require.Len(t, st.outbox, 2)

	require.Equal(t, "consul.delete", st.outbox[0].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/placement", st.outbox[0].Payload["key"])

	require.Equal(t, "consul.delete", st.outbox[1].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/connections/postgres/default", st.outbox[1].Payload["key"])
}
