package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/testkit"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestOrderRoutingRepositoryPersistsTenantScopedOrders(t *testing.T) {
	info := testkit.PostgresInfo(t)
	manager := newRepositoryTestManager(t, info, map[string]pdtenantdb.Placement{
		"tenant-orders-a": {
			TenantID:    "tenant-orders-a",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      info.DBName,
			SchemaName:  "t_tenant_orders_a",
		},
		"tenant-orders-b": {
			TenantID:    "tenant-orders-b",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      info.DBName,
			SchemaName:  "t_tenant_orders_b",
		},
	})
	t.Cleanup(func() { _ = manager.CloseAll() })

	repo := NewOrderRoutingRepository(manager)
	shipmentSLA := time.Date(2026, 5, 15, 18, 0, 0, 0, time.UTC)
	issueSLA := shipmentSLA.Add(3 * time.Hour)
	shippedAt := time.Date(2026, 5, 15, 8, 30, 0, 0, time.UTC)
	deliveredAt := shippedAt.Add(48 * time.Hour)
	createdAt := time.Date(2026, 5, 15, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)

	order := entity.RoutedOrder{
		ID:           "ord-repo-1",
		CandidateID:  "cand-1",
		ProductTitle: "Vintage Tee",
		Partner:      "Print Partner A",
		Quantity:     2,
		Total:        "$40.00",
		CustomerName: "Alex POD",
		Status:       entity.RoutedOrderStatusShipped,
		Timeline:     []string{"created", "shipment delivered"},
		ActivityLog: []entity.RoutedOrderActivity{
			{
				Type:      entity.RoutedOrderActivityTypeSystem,
				Actor:     "system",
				Message:   "created",
				Details:   []entity.RoutedOrderActivityDetail{{Key: "status", Value: entity.RoutedOrderStatusQueued}},
				CreatedAt: createdAt,
			},
			{
				Type:      entity.RoutedOrderActivityTypeShipmentNote,
				Actor:     "user:12",
				Message:   "Delivered cleanly",
				Details:   []entity.RoutedOrderActivityDetail{{Key: "shipment_status", Value: entity.RoutedOrderShipmentStatusDelivered}},
				CreatedAt: updatedAt,
			},
		},
		ExceptionType:          "reprint_request",
		ExceptionStatus:        entity.RoutedOrderExceptionStatusResolved,
		ShipmentStatus:         entity.RoutedOrderShipmentStatusDelivered,
		ShipmentCarrier:        "DHL",
		ShipmentTrackingNumber: "TRACK-123",
		ShipmentTrackingURL:    "https://tracking.example/TRACK-123",
		ShipmentNotes:          "Delivered cleanly",
		OperatorAssignee:       "ops.lead",
		ShipmentSlaDueAt:       &shipmentSLA,
		IssueSlaDueAt:          &issueSLA,
		BaseCostSnapshot:       "$16.00",
		FulfillmentCost:        "$18.00",
		ShippingCost:           "$5.50",
		IssueCost:              "$2.00",
		IssueResolution:        entity.RoutedOrderIssueResolutionReprint,
		IssueNotes:             "Reprint shipped",
		RealizedMargin:         "$14.50",
		SettlementStatus:       entity.RoutedOrderSettlementStatusPaid,
		SettlementNotes:        "Settled",
		ShippedAt:              &shippedAt,
		DeliveredAt:            &deliveredAt,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
	}

	ctxA := toolkit.WithTenantID(context.Background(), "tenant-orders-a")
	saved, err := repo.Create(ctxA, order)
	require.NoError(t, err)
	require.Equal(t, "ord-repo-1", saved.ID)

	got, err := repo.GetByID(ctxA, "ord-repo-1")
	require.NoError(t, err)
	require.Equal(t, order.CandidateID, got.CandidateID)
	require.Equal(t, order.OperatorAssignee, got.OperatorAssignee)
	require.NotNil(t, got.ShipmentSlaDueAt)
	require.True(t, got.ShipmentSlaDueAt.Equal(shipmentSLA))
	require.NotNil(t, got.IssueSlaDueAt)
	require.True(t, got.IssueSlaDueAt.Equal(issueSLA))
	require.NotNil(t, got.ShippedAt)
	require.True(t, got.ShippedAt.Equal(shippedAt))
	require.NotNil(t, got.DeliveredAt)
	require.True(t, got.DeliveredAt.Equal(deliveredAt))
	require.Equal(t, []string{"created", "shipment delivered"}, got.Timeline)
	require.Len(t, got.ActivityLog, 2)
	require.Equal(t, entity.RoutedOrderActivityTypeShipmentNote, got.ActivityLog[1].Type)
	require.Equal(t, "user:12", got.ActivityLog[1].Actor)
	require.Equal(t, "Delivered cleanly", got.ActivityLog[1].Message)
	require.Len(t, got.ActivityLog[1].Details, 1)
	require.Equal(t, "shipment_status", got.ActivityLog[1].Details[0].Key)
	require.Equal(t, entity.RoutedOrderShipmentStatusDelivered, got.ActivityLog[1].Details[0].Value)
	require.Equal(t, entity.RoutedOrderSettlementStatusPaid, got.SettlementStatus)
	require.Equal(t, entity.RoutedOrderIssueResolutionReprint, got.IssueResolution)

	list, err := repo.List(ctxA)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "ord-repo-1", list[0].ID)

	feedPage, err := repo.ListActivityFeed(ctxA, inputport.ListRoutedOrderActivitiesQuery{
		ActivityType:  "all",
		Limit:         10,
		IncludeSystem: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, feedPage.Total)
	require.Len(t, feedPage.Entries, 2)
	require.Equal(t, "ord-repo-1", feedPage.Entries[0].OrderID)

	ctxB := toolkit.WithTenantID(context.Background(), "tenant-orders-b")
	_, err = repo.GetByID(ctxB, "ord-repo-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "routed order not found")

	err = manager.WithTenantTx(ctxA, "tenant-orders-a", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var versions []string
		if err := tx.SelectContext(ctxA, &versions, `SELECT version FROM backoffice_schema_migrations ORDER BY version`); err != nil {
			return err
		}
		if len(versions) == 0 {
			return fmt.Errorf("missing migrations")
		}
		if !strings.Contains(strings.Join(versions, ","), "0009_create_routed_order_activities") {
			return fmt.Errorf("missing routed order activities migration")
		}
		return nil
	})
	require.NoError(t, err)
}

type repositoryTestResolver struct {
	mu         sync.Mutex
	placements map[string]pdtenantdb.Placement
}

func (r *repositoryTestResolver) Resolve(_ context.Context, tenantID string) (pdtenantdb.Placement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pl, ok := r.placements[tenantID]
	if !ok {
		return pdtenantdb.Placement{}, fmt.Errorf("missing placement for %s", tenantID)
	}
	return pl, nil
}

type repositoryTestRegistry struct {
	cfg pdtenantdb.ClusterConfig
}

func (r *repositoryTestRegistry) GetCluster(context.Context, string) (pdtenantdb.ClusterConfig, error) {
	return r.cfg, nil
}

func newRepositoryTestManager(
	t *testing.T,
	info testkit.PostgresConnInfo,
	placements map[string]pdtenantdb.Placement,
) pdtenantdb.Manager {
	t.Helper()

	registry := &repositoryTestRegistry{cfg: pdtenantdb.ClusterConfig{
		Host:     info.Host,
		Port:     info.Port,
		User:     info.User,
		Password: info.Password,
		SSLMode:  "disable",
	}}
	resolver := &repositoryTestResolver{placements: placements}
	return pdtenantdb.NewManager(&pdtenantdb.Config{
		SharedDB: info.DBName,
	}, resolver, registry)
}
