package routing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	orderentity "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	"github.com/tuannm99/podzone/pkg/ddd"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	pdtenantdbmocks "github.com/tuannm99/podzone/pkg/pdtenantdb/mocks"
	"github.com/tuannm99/podzone/pkg/testkit"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func skipIfDockerUnavailable(t *testing.T) {
	t.Helper()
	if _, ok := os.LookupEnv("XDG_RUNTIME_DIR"); !ok {
		t.Skip("docker-backed integration test requires XDG_RUNTIME_DIR")
	}
}

func TestOrderRoutingRepositoryPersistsTenantScopedOrders(t *testing.T) {
	skipIfDockerUnavailable(t)
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

	repo := New(manager)
	shipmentSLA := time.Date(2026, 5, 15, 18, 0, 0, 0, time.UTC)
	issueSLA := shipmentSLA.Add(3 * time.Hour)
	shippedAt := time.Date(2026, 5, 15, 8, 30, 0, 0, time.UTC)
	deliveredAt := shippedAt.Add(48 * time.Hour)
	createdAt := time.Date(2026, 5, 15, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)

	order := routingentity.RoutedOrder{
		ID:           "ord-repo-1",
		StoreID:      "store-repo-1",
		CandidateID:  "cand-1",
		ProductTitle: "Vintage Tee",
		Partner:      "Print Partner A",
		Quantity:     2,
		Total:        "$40.00",
		CustomerName: "Alex POD",
		Status:       routingentity.RoutedOrderStatusShipped,
		Timeline:     []string{"created", "shipment delivered"},
		ActivityLog: []routingentity.RoutedOrderActivity{
			{
				Type:    routingentity.RoutedOrderActivityTypeSystem,
				Actor:   "system",
				Message: "created",
				Details: []routingentity.RoutedOrderActivityDetail{
					{Key: "status", Value: routingentity.RoutedOrderStatusQueued},
				},
				CreatedAt: createdAt,
			},
			{
				Type:    routingentity.RoutedOrderActivityTypeShipmentNote,
				Actor:   "user:12",
				Message: "Delivered cleanly",
				Details: []routingentity.RoutedOrderActivityDetail{
					{Key: "shipment_status", Value: routingentity.RoutedOrderShipmentStatusDelivered},
				},
				CreatedAt: updatedAt,
			},
		},
		ExceptionType:          "reprint_request",
		ExceptionStatus:        routingentity.RoutedOrderExceptionStatusResolved,
		ShipmentStatus:         routingentity.RoutedOrderShipmentStatusDelivered,
		ShipmentCarrier:        "DHL",
		ShipmentTrackingNumber: "TRACK-123",
		ShipmentTrackingURL:    "https://tracking.example/TRACK-123",
		ShipmentNotes:          "Delivered cleanly",
		OperatorAssignee:       "ops.lead",
		ShipmentSlaDueAt:       &shipmentSLA,
		IssueSlaDueAt:          &issueSLA,
		RoutingBlockCode:       "",
		RoutingBlockReason:     "",
		BaseCostSnapshot:       "$16.00",
		FulfillmentCost:        "$18.00",
		ShippingCost:           "$5.50",
		IssueCost:              "$2.00",
		IssueResolution:        routingentity.RoutedOrderIssueResolutionReprint,
		IssueNotes:             "Reprint shipped",
		RealizedMargin:         "$14.50",
		SettlementStatus:       routingentity.RoutedOrderSettlementStatusPaid,
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
	require.Equal(t, ddd.Version(0), saved.AggregateVersion)

	customerOrder, err := repo.GetCustomerOrder(ctxA, order.StoreID, order.ID)
	require.NoError(t, err)
	require.Equal(t, order.ID, customerOrder.AggregateID().String())
	require.Equal(t, ddd.Version(0), customerOrder.AggregateVersion())
	require.Equal(t, orderentity.StatusShipped, customerOrder.Snapshot().Status)

	got, err := repo.GetByID(ctxA, "ord-repo-1")
	require.NoError(t, err)
	require.Equal(t, order.CandidateID, got.CandidateID)
	require.Equal(t, order.StoreID, got.StoreID)
	require.Equal(t, order.OperatorAssignee, got.OperatorAssignee)
	require.NotNil(t, got.ShipmentSlaDueAt)
	require.True(t, got.ShipmentSlaDueAt.Equal(shipmentSLA))
	require.NotNil(t, got.IssueSlaDueAt)
	require.True(t, got.IssueSlaDueAt.Equal(issueSLA))
	require.Equal(t, "", got.RoutingBlockCode)
	require.Equal(t, "", got.RoutingBlockReason)
	require.NotNil(t, got.ShippedAt)
	require.True(t, got.ShippedAt.Equal(shippedAt))
	require.NotNil(t, got.DeliveredAt)
	require.True(t, got.DeliveredAt.Equal(deliveredAt))
	require.Equal(t, []string{"created", "shipment delivered"}, got.Timeline)
	require.Len(t, got.ActivityLog, 2)
	require.Equal(t, routingentity.RoutedOrderActivityTypeShipmentNote, got.ActivityLog[1].Type)
	require.Equal(t, "user:12", got.ActivityLog[1].Actor)
	require.Equal(t, "Delivered cleanly", got.ActivityLog[1].Message)
	require.Len(t, got.ActivityLog[1].Details, 1)
	require.Equal(t, "shipment_status", got.ActivityLog[1].Details[0].Key)
	require.Equal(t, routingentity.RoutedOrderShipmentStatusDelivered, got.ActivityLog[1].Details[0].Value)
	require.Equal(t, routingentity.RoutedOrderSettlementStatusPaid, got.SettlementStatus)
	require.Equal(t, routingentity.RoutedOrderIssueResolutionReprint, got.IssueResolution)

	staleOrder := *got
	got.OperatorAssignee = "ops.next"
	got.UpdatedAt = updatedAt.Add(time.Hour)
	updated, err := repo.Update(ctxA, *got)
	require.NoError(t, err)
	require.Equal(t, ddd.Version(1), updated.AggregateVersion)

	customerOrder, err = repo.GetCustomerOrder(ctxA, order.StoreID, order.ID)
	require.NoError(t, err)
	require.Equal(t, ddd.Version(1), customerOrder.AggregateVersion())
	require.Equal(t, "ops.next", customerOrder.Snapshot().OperatorAssignee)

	staleOrder.OperatorAssignee = "ops.stale"
	_, err = repo.Update(ctxA, staleOrder)
	require.ErrorIs(t, err, ddd.ErrVersionConflict)

	list, err := repo.List(ctxA)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "ord-repo-1", list[0].ID)

	feedPage, err := repo.ListActivityFeed(ctxA, routingentity.RoutedOrderActivityFeedQuery{
		ActivityType:  "all",
		Limit:         10,
		IncludeSystem: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, feedPage.Total)
	require.Len(t, feedPage.Entries, 2)
	require.Equal(t, "ord-repo-1", feedPage.Entries[0].OrderID)
	require.Equal(t, "Print Partner A", feedPage.Entries[0].Partner)

	filteredFeed, err := repo.ListActivityFeed(ctxA, routingentity.RoutedOrderActivityFeedQuery{
		OrderID:       "ord-repo-1",
		Partner:       "partner a",
		Assignee:      "ops.lead",
		ActivityType:  routingentity.RoutedOrderActivityTypeShipmentNote,
		Limit:         10,
		IncludeSystem: true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, filteredFeed.Total)
	require.Len(t, filteredFeed.Entries, 1)
	require.Equal(t, routingentity.RoutedOrderActivityTypeShipmentNote, filteredFeed.Entries[0].Activity.Type)

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
		if !strings.Contains(strings.Join(versions, ","), "0010_drop_routed_order_activity_log_cache") {
			return fmt.Errorf("missing legacy activity log cleanup migration")
		}
		return nil
	})
	require.NoError(t, err)
}

func TestOrderRoutingRepositoryBackfillsLegacyActivityLogOnLegacyMigration(t *testing.T) {
	skipIfDockerUnavailable(t)
	info := testkit.PostgresInfo(t)
	manager := newRepositoryTestManager(t, info, map[string]pdtenantdb.Placement{
		"tenant-orders-legacy": {
			TenantID:    "tenant-orders-legacy",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      info.DBName,
			SchemaName:  "t_tenant_orders_legacy",
		},
	})
	t.Cleanup(func() { _ = manager.CloseAll() })

	ctx := toolkit.WithTenantID(context.Background(), "tenant-orders-legacy")
	err := manager.WithTenantTx(ctx, "tenant-orders-legacy", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		statements := []string{
			`CREATE TABLE IF NOT EXISTS backoffice_schema_migrations (
				version TEXT PRIMARY KEY,
				applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`CREATE TABLE IF NOT EXISTS routed_orders (
				id TEXT PRIMARY KEY,
				candidate_id TEXT NOT NULL,
				product_title TEXT NOT NULL,
				partner TEXT NOT NULL,
				quantity INTEGER NOT NULL,
				total TEXT NOT NULL,
				customer_name TEXT NOT NULL,
				status TEXT NOT NULL,
				timeline_json TEXT NOT NULL,
				exception_type TEXT NOT NULL DEFAULT '',
				exception_status TEXT NOT NULL DEFAULT '',
				created_at TIMESTAMPTZ NOT NULL,
				updated_at TIMESTAMPTZ NOT NULL
			)`,
			`CREATE TABLE IF NOT EXISTS stores (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				status TEXT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL,
				updated_at TIMESTAMPTZ NOT NULL
			)`,
			`ALTER TABLE routed_orders
				ADD COLUMN IF NOT EXISTS shipment_status TEXT NOT NULL DEFAULT 'awaiting_label',
				ADD COLUMN IF NOT EXISTS shipment_carrier TEXT NOT NULL DEFAULT '',
				ADD COLUMN IF NOT EXISTS shipment_tracking_number TEXT NOT NULL DEFAULT '',
				ADD COLUMN IF NOT EXISTS shipment_tracking_url TEXT NOT NULL DEFAULT '',
				ADD COLUMN IF NOT EXISTS shipment_notes TEXT NOT NULL DEFAULT '',
				ADD COLUMN IF NOT EXISTS shipped_at TIMESTAMPTZ NULL,
				ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ NULL`,
			`ALTER TABLE routed_orders
				ADD COLUMN IF NOT EXISTS base_cost_snapshot TEXT NOT NULL DEFAULT '$0.00',
				ADD COLUMN IF NOT EXISTS fulfillment_cost TEXT NOT NULL DEFAULT '$0.00',
				ADD COLUMN IF NOT EXISTS shipping_cost TEXT NOT NULL DEFAULT '$0.00',
				ADD COLUMN IF NOT EXISTS realized_margin TEXT NOT NULL DEFAULT '$0.00',
				ADD COLUMN IF NOT EXISTS settlement_status TEXT NOT NULL DEFAULT 'pending',
				ADD COLUMN IF NOT EXISTS settlement_notes TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE routed_orders
				ADD COLUMN IF NOT EXISTS issue_cost TEXT NOT NULL DEFAULT '$0.00',
				ADD COLUMN IF NOT EXISTS issue_resolution TEXT NOT NULL DEFAULT 'monitor',
				ADD COLUMN IF NOT EXISTS issue_notes TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE routed_orders
				ADD COLUMN IF NOT EXISTS operator_assignee TEXT NOT NULL DEFAULT 'unassigned',
				ADD COLUMN IF NOT EXISTS shipment_sla_due_at TIMESTAMPTZ NULL,
				ADD COLUMN IF NOT EXISTS issue_sla_due_at TIMESTAMPTZ NULL`,
			`ALTER TABLE routed_orders
				ADD COLUMN IF NOT EXISTS activity_log_json TEXT NOT NULL DEFAULT '[]'`,
		}
		for _, stmt := range statements {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return err
			}
		}

		versions := []string{
			"0001_create_stores",
			"0002_create_product_setup",
			"0003_create_routed_orders",
			"0004_add_manual_shipment_fields",
			"0005_add_order_settlement_fields",
			"0006_add_order_issue_cost_fields",
			"0007_add_order_queue_control_fields",
			"0008_add_order_activity_log",
		}
		for _, version := range versions {
			if _, err := tx.ExecContext(ctx, `INSERT INTO backoffice_schema_migrations (version) VALUES ($1)`, version); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO stores (id, name, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`,
			"store-legacy-1",
			"Legacy Store",
			"active",
			time.Date(2026, 5, 14, 6, 0, 0, 0, time.UTC),
			time.Date(2026, 5, 14, 6, 0, 0, 0, time.UTC),
		); err != nil {
			return err
		}

		activityLogJSON := `[{"type":"system","actor":"system","message":"Order created for Legacy Tee","details":[{"key":"status","value":"queued"}],"createdAt":"2026-05-14T07:00:00Z"},{"type":"shipment_note","actor":"user:88","message":"Label printed","details":[{"key":"shipment_status","value":"label_ready"}],"createdAt":"2026-05-14T08:00:00Z"}]`
		_, err := tx.ExecContext(ctx, `
			INSERT INTO routed_orders (
				id, candidate_id, product_title, partner, quantity, total, customer_name, status,
				timeline_json, activity_log_json, exception_type, exception_status, shipment_status,
				shipment_carrier, shipment_tracking_number, shipment_tracking_url, shipment_notes,
				operator_assignee, shipment_sla_due_at, issue_sla_due_at, base_cost_snapshot,
				fulfillment_cost, shipping_cost, issue_cost, issue_resolution, issue_notes,
				realized_margin, settlement_status, settlement_notes, shipped_at, delivered_at,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8,
				$9, $10, $11, $12, $13,
				$14, $15, $16, $17,
				$18, $19, $20, $21,
				$22, $23, $24, $25, $26,
				$27, $28, $29, $30, $31,
				$32, $33
			)
		`,
			"ord-legacy-1",
			"cand-legacy-1",
			"Legacy Tee",
			"Print Partner Legacy",
			1,
			"$24.00",
			"Legacy Customer",
			routingentity.RoutedOrderStatusQueued,
			`["created"]`,
			activityLogJSON,
			"",
			"",
			routingentity.RoutedOrderShipmentStatusLabelReady,
			"",
			"",
			"",
			"Label printed",
			"ops.legacy",
			nil,
			nil,
			"$10.00",
			"$10.00",
			"$0.00",
			"$0.00",
			routingentity.RoutedOrderIssueResolutionMonitor,
			"",
			"$14.00",
			routingentity.RoutedOrderSettlementStatusPending,
			"",
			nil,
			nil,
			time.Date(2026, 5, 14, 7, 0, 0, 0, time.UTC),
			time.Date(2026, 5, 14, 8, 0, 0, 0, time.UTC),
		)
		return err
	})
	require.NoError(t, err)

	repo := NewOrderRoutingRepository(manager)
	feedPage, err := repo.ListActivityFeed(ctx, routingentity.RoutedOrderActivityFeedQuery{
		ActivityType:  "all",
		Limit:         10,
		IncludeSystem: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, feedPage.Total)
	require.Len(t, feedPage.Entries, 2)
	require.Equal(t, "ord-legacy-1", feedPage.Entries[0].OrderID)
	require.Equal(t, "Print Partner Legacy", feedPage.Entries[0].Partner)
	require.Equal(t, routingentity.RoutedOrderActivityTypeShipmentNote, feedPage.Entries[0].Activity.Type)
	require.Equal(t, "user:88", feedPage.Entries[0].Activity.Actor)
	require.Equal(t, "ops.legacy", feedPage.Entries[0].OperatorAssignee)

	err = manager.WithTenantTx(ctx, "tenant-orders-legacy", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var versions []string
		if err := tx.SelectContext(ctx, &versions, `SELECT version FROM backoffice_schema_migrations ORDER BY version`); err != nil {
			return err
		}
		require.Contains(t, versions, "0009_create_routed_order_activities")
		require.Contains(t, versions, "0010_drop_routed_order_activity_log_cache")

		var count int
		if err := tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM routed_order_activities WHERE order_id = $1`, "ord-legacy-1"); err != nil {
			return err
		}
		require.Equal(t, 2, count)

		var hasLegacyColumn bool
		if err := tx.GetContext(ctx, &hasLegacyColumn, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = current_schema()
				  AND table_name = 'routed_orders'
				  AND column_name = 'activity_log_json'
			)
		`); err != nil {
			return err
		}
		require.False(t, hasLegacyColumn)
		return nil
	})
	require.NoError(t, err)
}

func newRepositoryTestManager(
	t *testing.T,
	info testkit.PostgresConnInfo,
	placements map[string]pdtenantdb.Placement,
) pdtenantdb.Manager {
	t.Helper()

	registry := pdtenantdbmocks.NewMockClusterRegistry(t)
	registry.EXPECT().GetCluster(mock.Anything, "pg-01").Return(pdtenantdb.ClusterConfig{
		Host:     info.Host,
		Port:     info.Port,
		User:     info.User,
		Password: info.Password,
		SSLMode:  "disable",
	}, nil).Maybe()
	resolver := pdtenantdbmocks.NewMockPlacementResolver(t)
	resolver.EXPECT().
		Resolve(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, tenantID string) (pdtenantdb.Placement, error) {
			pl, ok := placements[tenantID]
			if !ok {
				return pdtenantdb.Placement{}, fmt.Errorf("missing placement for %s", tenantID)
			}
			return pl, nil
		}).
		Maybe()
	return pdtenantdb.NewManager(&pdtenantdb.Config{
		SharedDB: info.DBName,
	}, resolver, registry)
}
