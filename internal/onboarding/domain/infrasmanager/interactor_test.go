package infrasmanager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	coremocks "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/collection"
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
		ListConnections(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(collection.Page[entity.ConnectionInfo]{}, nil).
		Maybe()
	store.EXPECT().
		ListEvents(mock.Anything, mock.Anything, mock.Anything).
		Return(collection.Page[entity.ConnectionEvent]{}, nil).
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

func TestUpsertDatabaseCluster_ValidatesAndPersistsInventory(t *testing.T) {
	inventory := coremocks.NewMockResourceInventoryRepository(t)
	svc := &Interactor{inventory: inventory}
	inventory.EXPECT().
		UpsertDatabaseCluster(mock.Anything, entity.DatabaseCluster{
			Name:           "pg-primary",
			Engine:         "postgres",
			Region:         "local",
			PlacementDB:    "podzone_tenants",
			MaxTenants:     100,
			MaxSchemas:     100,
			MaxConnections: 500,
			Status:         "active",
			Healthy:        true,
		}).
		Return(nil).
		Once()

	err := svc.UpsertDatabaseCluster(context.Background(), inputport.DatabaseClusterResource{
		Name:           " pg-primary ",
		Engine:         "postgres",
		Region:         "local",
		PlacementDB:    "podzone_tenants",
		MaxTenants:     100,
		MaxSchemas:     100,
		MaxConnections: 500,
		Healthy:        true,
	})

	require.NoError(t, err)
}

func TestUpsertRuntimePool_RejectsNegativeCapacity(t *testing.T) {
	svc := &Interactor{}

	err := svc.UpsertRuntimePool(context.Background(), inputport.RuntimePoolResource{
		Name:       "docker-local",
		Kind:       "docker",
		MaxTenants: -1,
	})

	require.ErrorIs(t, err, entity.ErrInvalidInput)
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

	require.Equal(t, "kv_store.publish", state.outbox[0].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/connections/postgres/default", state.outbox[0].Payload["key"])

	require.Equal(t, "kv_store.publish", state.outbox[1].Topic)
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

	require.Equal(t, "kv_store.delete", state.outbox[0].Topic)
	require.Equal(t, "podzone/tenants/tenant-1/placement", state.outbox[0].Payload["key"])

	require.Equal(t, "kv_store.delete", state.outbox[1].Topic)
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

	placements.EXPECT().GetTenantPlacementAllocation(mock.Anything, "tenant-1").Return(&allocation, nil)
	writer.EXPECT().PublishPlacementRoute(mock.Anything, "tenant-1", allocation).Return(nil)

	ready, err := svc.EnsurePlacementRoute(context.Background(), "tenant-1")
	require.NoError(t, err)
	require.True(t, ready)
}

func TestGetTenantPlacementStatusDetectsMissingRoute(t *testing.T) {
	placements := coremocks.NewMockPlacementRepository(t)
	reader := coremocks.NewMockPlacementRouteReader(t)
	svc := &Interactor{
		placements:  placements,
		routeReader: reader,
	}
	allocation := entity.PlacementAllocation{
		ID:          "allocation-1",
		TenantID:    "tenant-1",
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
		Status:      "ready",
	}

	placements.EXPECT().GetTenantPlacementAllocation(mock.Anything, "tenant-1").Return(&allocation, nil)
	reader.EXPECT().GetPlacementRoute(mock.Anything, "tenant-1").Return(nil, nil)

	status, err := svc.GetTenantPlacementStatus(context.Background(), "tenant-1")
	require.NoError(t, err)
	require.True(t, status.AllocationReady)
	require.False(t, status.RouteReady)
	require.False(t, status.InSync)
	require.True(t, status.NeedsRepair)
	require.Equal(t, "placement route is missing", status.Reason)
}

func TestReconcileTenantPlacementRepublishesDriftedRoute(t *testing.T) {
	placements := coremocks.NewMockPlacementRepository(t)
	reader := coremocks.NewMockPlacementRouteReader(t)
	writer := coremocks.NewMockPlacementRouteWriter(t)
	svc := &Interactor{
		placements:  placements,
		routeReader: reader,
		routeWriter: writer,
	}
	allocation := entity.PlacementAllocation{
		ID:          "allocation-1",
		TenantID:    "tenant-1",
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
		Status:      "ready",
	}
	driftedRoute := &entity.PlacementRoute{
		ClusterName: "pg-old",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
	}

	placements.EXPECT().GetTenantPlacementAllocation(mock.Anything, "tenant-1").Return(&allocation, nil).Twice()
	reader.EXPECT().GetPlacementRoute(mock.Anything, "tenant-1").Return(driftedRoute, nil)
	writer.EXPECT().PublishPlacementRoute(mock.Anything, "tenant-1", allocation).Return(nil)

	resp, err := svc.ReconcileTenantPlacement(
		context.Background(),
		"tenant-1",
		map[string]string{"user": "7"},
	)
	require.NoError(t, err)
	require.True(t, resp.Repaired)
	require.True(t, resp.Status.RouteReady)
	require.True(t, resp.Status.InSync)
	require.False(t, resp.Status.NeedsRepair)
	require.Equal(t, "podzone/tenants/tenant-1/placement", resp.KVStoreKey)
	require.NotNil(t, resp.PublishedAt)
}

func TestProvisionStorePlacement_ReusesTenantAllocationAcrossStores(t *testing.T) {
	placements := coremocks.NewMockPlacementRepository(t)
	plans := coremocks.NewMockPlacementPlanRepository(t)
	planner := coremocks.NewMockPlacementPlanner(t)
	provisioner := coremocks.NewMockStorageProvisioner(t)
	svc := &Interactor{
		placements:  placements,
		plans:       plans,
		planner:     planner,
		provisioner: provisioner,
	}
	allocation := entity.PlacementAllocation{
		ID:          "allocation-1",
		RequestID:   "request-1",
		TenantID:    "tenant-1",
		StoreID:     "store-1",
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
		Status:      "ready",
	}
	placements.EXPECT().
		GetTenantPlacementAllocation(mock.Anything, "tenant-1").
		Return(&allocation, nil)

	response, err := svc.ProvisionStorePlacement(
		context.Background(),
		inputport.ProvisionStorePlacementRequest{
			RequestID: "request-2",
			TenantID:  "tenant-1",
			StoreID:   "store-2",
		},
		nil,
	)

	require.NoError(t, err)
	require.Equal(t, "allocation-1", response.AllocationID)
	plans.AssertNotCalled(t, "GetPlacementPlanByRequestID", mock.Anything, mock.Anything)
	planner.AssertNotCalled(t, "PlanStorePlacement", mock.Anything, mock.Anything)
	provisioner.AssertNotCalled(t, "ProvisionStorePlacement", mock.Anything, mock.Anything, mock.Anything)
}

func TestProvisionStorePlacement_SavesPlanBeforeProvisioning(t *testing.T) {
	state := &connectionStoreState{}
	store := newConnectionStoreMock(t, state)
	placements := coremocks.NewMockPlacementRepository(t)
	plans := coremocks.NewMockPlacementPlanRepository(t)
	planner := coremocks.NewMockPlacementPlanner(t)
	provisioner := coremocks.NewMockStorageProvisioner(t)
	svc := &Interactor{
		st:          store,
		placements:  placements,
		plans:       plans,
		planner:     planner,
		provisioner: provisioner,
	}
	req := inputport.ProvisionStorePlacementRequest{
		RequestID:   "request-1",
		TenantID:    "tenant-1",
		StoreID:     "store-1",
		Subdomain:   "tenant-one",
		RequestedBy: "user-1",
	}
	placementReq := entity.StorePlacementRequest{
		RequestID:   req.RequestID,
		TenantID:    req.TenantID,
		StoreID:     req.StoreID,
		Subdomain:   req.Subdomain,
		RequestedBy: req.RequestedBy,
	}
	plan := entity.PlacementPlan{
		RequestID:   req.RequestID,
		TenantID:    req.TenantID,
		StoreID:     req.StoreID,
		Runtime:     entity.PlacementRuntimeLocalDocker,
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
	}
	allocation := entity.PlacementAllocation{
		ID:          "allocation-1",
		RequestID:   req.RequestID,
		TenantID:    req.TenantID,
		StoreID:     req.StoreID,
		Runtime:     entity.PlacementRuntimeLocalDocker,
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
		Endpoint:    "postgres://postgres:***@pgbouncer:6432/podzone_tenants",
		SecretRef:   "docker/postgres/default",
		Status:      "ready",
	}

	placements.EXPECT().GetTenantPlacementAllocation(mock.Anything, req.TenantID).Return(nil, nil)
	plans.EXPECT().GetPlacementPlanByRequestID(mock.Anything, req.RequestID).Return(nil, nil)
	planner.EXPECT().PlanStorePlacement(mock.Anything, placementReq).Return(plan, nil)
	plans.EXPECT().SavePlacementPlan(mock.Anything, plan).Return(nil)
	provisioner.EXPECT().ProvisionStorePlacement(mock.Anything, placementReq, plan).Return(allocation, nil)
	placements.EXPECT().SavePlacementAllocation(mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.ProvisionStorePlacement(context.Background(), req, map[string]string{"service": "test"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "allocation-1", resp.AllocationID)
	require.Len(t, state.outbox, 2)
}

func TestCheckDatabaseClusterHealthUpdatesInventory(t *testing.T) {
	inventory := coremocks.NewMockResourceInventoryRepository(t)
	placements := coremocks.NewMockPlacementRepository(t)
	checker := coremocks.NewMockResourceHealthChecker(t)
	svc := &Interactor{
		inventory:  inventory,
		placements: placements,
		health:     checker,
	}
	cluster := &entity.DatabaseCluster{
		Name:        "pg-default",
		Engine:      "postgres",
		PlacementDB: "podzone_tenants",
	}
	checkedAt := time.Now().UTC()
	health := entity.DatabaseClusterHealth{
		Healthy:            true,
		CurrentSchemas:     3,
		CurrentConnections: 4,
		Message:            "ok",
		CheckedAt:          checkedAt,
	}

	inventory.EXPECT().GetDatabaseCluster(mock.Anything, "pg-default").Return(cluster, nil)
	checker.EXPECT().CheckDatabaseClusterHealth(mock.Anything, *cluster).Return(health, nil)
	placements.EXPECT().CountReadyPlacementAllocationsByCluster(mock.Anything, "pg-default").Return(2, nil)
	inventory.EXPECT().
		UpdateDatabaseClusterHealth(
			mock.Anything,
			"pg-default",
			mock.MatchedBy(func(next entity.DatabaseClusterHealth) bool {
				return next.Healthy &&
					next.CurrentTenants == 2 &&
					next.CurrentSchemas == 3 &&
					next.CurrentConnections == 4
			}),
		).
		Return(nil)

	resp, err := svc.CheckDatabaseClusterHealth(context.Background(), "pg-default")
	require.NoError(t, err)
	require.Equal(t, "pg-default", resp.Name)
	require.True(t, resp.Healthy)
	require.Equal(t, 2, resp.CurrentTenants)
	require.Equal(t, 3, resp.CurrentSchemas)
	require.Equal(t, 4, resp.CurrentConnections)
	require.Equal(t, checkedAt, resp.CheckedAt)
}

func TestProvisionStorePlacement_ReusesPersistedPlanOnRetry(t *testing.T) {
	state := &connectionStoreState{}
	store := newConnectionStoreMock(t, state)
	placements := coremocks.NewMockPlacementRepository(t)
	plans := coremocks.NewMockPlacementPlanRepository(t)
	planner := coremocks.NewMockPlacementPlanner(t)
	provisioner := coremocks.NewMockStorageProvisioner(t)
	svc := &Interactor{
		st:          store,
		placements:  placements,
		plans:       plans,
		planner:     planner,
		provisioner: provisioner,
	}
	req := inputport.ProvisionStorePlacementRequest{
		RequestID:   "request-1",
		TenantID:    "tenant-1",
		StoreID:     "store-1",
		Subdomain:   "tenant-one",
		RequestedBy: "user-1",
	}
	placementReq := entity.StorePlacementRequest{
		RequestID:   req.RequestID,
		TenantID:    req.TenantID,
		StoreID:     req.StoreID,
		Subdomain:   req.Subdomain,
		RequestedBy: req.RequestedBy,
	}
	plan := entity.PlacementPlan{
		RequestID:   req.RequestID,
		TenantID:    req.TenantID,
		StoreID:     req.StoreID,
		Runtime:     entity.PlacementRuntimeLocalDocker,
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
	}
	allocation := entity.PlacementAllocation{
		ID:          "allocation-1",
		RequestID:   req.RequestID,
		TenantID:    req.TenantID,
		StoreID:     req.StoreID,
		Runtime:     entity.PlacementRuntimeLocalDocker,
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
		Endpoint:    "postgres://postgres:***@pgbouncer:6432/podzone_tenants",
		SecretRef:   "docker/postgres/default",
		Status:      "ready",
	}

	placements.EXPECT().GetTenantPlacementAllocation(mock.Anything, req.TenantID).Return(nil, nil)
	plans.EXPECT().GetPlacementPlanByRequestID(mock.Anything, req.RequestID).Return(&plan, nil)
	provisioner.EXPECT().ProvisionStorePlacement(mock.Anything, placementReq, plan).Return(allocation, nil)
	placements.EXPECT().SavePlacementAllocation(mock.Anything, mock.Anything).Return(nil)
	planner.EXPECT().PlanStorePlacement(mock.Anything, mock.Anything).Maybe()

	resp, err := svc.ProvisionStorePlacement(context.Background(), req, map[string]string{"service": "test"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "allocation-1", resp.AllocationID)
	planner.AssertNotCalled(t, "PlanStorePlacement", mock.Anything, mock.Anything)
}
