package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	"github.com/tuannm99/podzone/internal/backoffice/migrations"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type OrderRoutingRepositoryImpl struct {
	mgr pdtenantdb.Manager
}

func NewOrderRoutingRepository(mgr pdtenantdb.Manager) outputport.OrderRoutingRepository {
	return &OrderRoutingRepositoryImpl{mgr: mgr}
}

type routedOrderRow struct {
	ID                     string       `db:"id"`
	CandidateID            string       `db:"candidate_id"`
	ProductTitle           string       `db:"product_title"`
	Partner                string       `db:"partner"`
	Quantity               int          `db:"quantity"`
	Total                  string       `db:"total"`
	CustomerName           string       `db:"customer_name"`
	Status                 string       `db:"status"`
	TimelineJSON           string       `db:"timeline_json"`
	ExceptionType          string       `db:"exception_type"`
	ExceptionStatus        string       `db:"exception_status"`
	ShipmentStatus         string       `db:"shipment_status"`
	ShipmentCarrier        string       `db:"shipment_carrier"`
	ShipmentTrackingNumber string       `db:"shipment_tracking_number"`
	ShipmentTrackingURL    string       `db:"shipment_tracking_url"`
	ShipmentNotes          string       `db:"shipment_notes"`
	OperatorAssignee       string       `db:"operator_assignee"`
	ShipmentSlaDueAt       sql.NullTime `db:"shipment_sla_due_at"`
	IssueSlaDueAt          sql.NullTime `db:"issue_sla_due_at"`
	BaseCostSnapshot       string       `db:"base_cost_snapshot"`
	FulfillmentCost        string       `db:"fulfillment_cost"`
	ShippingCost           string       `db:"shipping_cost"`
	IssueCost              string       `db:"issue_cost"`
	IssueResolution        string       `db:"issue_resolution"`
	IssueNotes             string       `db:"issue_notes"`
	RealizedMargin         string       `db:"realized_margin"`
	SettlementStatus       string       `db:"settlement_status"`
	SettlementNotes        string       `db:"settlement_notes"`
	ShippedAt              sql.NullTime `db:"shipped_at"`
	DeliveredAt            sql.NullTime `db:"delivered_at"`
	CreatedAt              time.Time    `db:"created_at"`
	UpdatedAt              time.Time    `db:"updated_at"`
}

func (r *OrderRoutingRepositoryImpl) List(ctx context.Context) ([]entity.RoutedOrder, error) {
	query, args, err := psql.
		Select("id", "candidate_id", "product_title", "partner", "quantity", "total", "customer_name", "status", "timeline_json", "exception_type", "exception_status", "shipment_status", "shipment_carrier", "shipment_tracking_number", "shipment_tracking_url", "shipment_notes", "operator_assignee", "shipment_sla_due_at", "issue_sla_due_at", "base_cost_snapshot", "fulfillment_cost", "shipping_cost", "issue_cost", "issue_resolution", "issue_notes", "realized_margin", "settlement_status", "settlement_notes", "shipped_at", "delivered_at", "created_at", "updated_at").
		From("routed_orders").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []routedOrderRow
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureRoutedOrderTables(ctx, tx); err != nil {
			return err
		}
		return tx.SelectContext(ctx, &rows, query, args...)
	}); err != nil {
		return nil, err
	}

	out := make([]entity.RoutedOrder, 0, len(rows))
	for _, row := range rows {
		order, err := mapRoutedOrderRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, order)
	}
	return out, nil
}

func (r *OrderRoutingRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.RoutedOrder, error) {
	query, args, err := psql.
		Select("id", "candidate_id", "product_title", "partner", "quantity", "total", "customer_name", "status", "timeline_json", "exception_type", "exception_status", "shipment_status", "shipment_carrier", "shipment_tracking_number", "shipment_tracking_url", "shipment_notes", "operator_assignee", "shipment_sla_due_at", "issue_sla_due_at", "base_cost_snapshot", "fulfillment_cost", "shipping_cost", "issue_cost", "issue_resolution", "issue_notes", "realized_margin", "settlement_status", "settlement_notes", "shipped_at", "delivered_at", "created_at", "updated_at").
		From("routed_orders").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row routedOrderRow
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureRoutedOrderTables(ctx, tx); err != nil {
			return err
		}
		if err := tx.GetContext(ctx, &row, query, args...); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("routed order not found")
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	order, err := mapRoutedOrderRow(row)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRoutingRepositoryImpl) Create(ctx context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error) {
	timelineJSON, err := json.Marshal(order.Timeline)
	if err != nil {
		return nil, err
	}
	query, args, err := psql.
		Insert("routed_orders").
		Columns("id", "candidate_id", "product_title", "partner", "quantity", "total", "customer_name", "status", "timeline_json", "exception_type", "exception_status", "shipment_status", "shipment_carrier", "shipment_tracking_number", "shipment_tracking_url", "shipment_notes", "operator_assignee", "shipment_sla_due_at", "issue_sla_due_at", "base_cost_snapshot", "fulfillment_cost", "shipping_cost", "issue_cost", "issue_resolution", "issue_notes", "realized_margin", "settlement_status", "settlement_notes", "shipped_at", "delivered_at", "created_at", "updated_at").
		Values(order.ID, order.CandidateID, order.ProductTitle, order.Partner, order.Quantity, order.Total, order.CustomerName, order.Status, string(timelineJSON), order.ExceptionType, order.ExceptionStatus, order.ShipmentStatus, order.ShipmentCarrier, order.ShipmentTrackingNumber, order.ShipmentTrackingURL, order.ShipmentNotes, order.OperatorAssignee, order.ShipmentSlaDueAt, order.IssueSlaDueAt, order.BaseCostSnapshot, order.FulfillmentCost, order.ShippingCost, order.IssueCost, order.IssueResolution, order.IssueNotes, order.RealizedMargin, order.SettlementStatus, order.SettlementNotes, order.ShippedAt, order.DeliveredAt, order.CreatedAt, order.UpdatedAt).
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureRoutedOrderTables(ctx, tx); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, query, args...)
		return err
	}); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRoutingRepositoryImpl) Update(ctx context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error) {
	timelineJSON, err := json.Marshal(order.Timeline)
	if err != nil {
		return nil, err
	}
	query, args, err := psql.
		Update("routed_orders").
		Set("candidate_id", order.CandidateID).
		Set("product_title", order.ProductTitle).
		Set("partner", order.Partner).
		Set("quantity", order.Quantity).
		Set("total", order.Total).
		Set("customer_name", order.CustomerName).
		Set("status", order.Status).
		Set("timeline_json", string(timelineJSON)).
		Set("exception_type", order.ExceptionType).
		Set("exception_status", order.ExceptionStatus).
		Set("shipment_status", order.ShipmentStatus).
		Set("shipment_carrier", order.ShipmentCarrier).
		Set("shipment_tracking_number", order.ShipmentTrackingNumber).
		Set("shipment_tracking_url", order.ShipmentTrackingURL).
		Set("shipment_notes", order.ShipmentNotes).
		Set("operator_assignee", order.OperatorAssignee).
		Set("shipment_sla_due_at", order.ShipmentSlaDueAt).
		Set("issue_sla_due_at", order.IssueSlaDueAt).
		Set("base_cost_snapshot", order.BaseCostSnapshot).
		Set("fulfillment_cost", order.FulfillmentCost).
		Set("shipping_cost", order.ShippingCost).
		Set("issue_cost", order.IssueCost).
		Set("issue_resolution", order.IssueResolution).
		Set("issue_notes", order.IssueNotes).
		Set("realized_margin", order.RealizedMargin).
		Set("settlement_status", order.SettlementStatus).
		Set("settlement_notes", order.SettlementNotes).
		Set("shipped_at", order.ShippedAt).
		Set("delivered_at", order.DeliveredAt).
		Set("updated_at", order.UpdatedAt).
		Where(sq.Eq{"id": order.ID}).
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureRoutedOrderTables(ctx, tx); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("routed order not found")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRoutingRepositoryImpl) withTenantTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return err
	}
	return r.mgr.WithTenantTx(ctx, tenantID, nil, fn)
}

func ensureRoutedOrderTables(ctx context.Context, tx *sqlx.Tx) error {
	return migrations.ApplyTx(ctx, tx)
}

func mapRoutedOrderRow(row routedOrderRow) (entity.RoutedOrder, error) {
	var timeline []string
	if err := json.Unmarshal([]byte(row.TimelineJSON), &timeline); err != nil {
		return entity.RoutedOrder{}, err
	}
	var shippedAt *time.Time
	if row.ShippedAt.Valid {
		shippedAt = &row.ShippedAt.Time
	}
	var deliveredAt *time.Time
	if row.DeliveredAt.Valid {
		deliveredAt = &row.DeliveredAt.Time
	}
	var shipmentSlaDueAt *time.Time
	if row.ShipmentSlaDueAt.Valid {
		shipmentSlaDueAt = &row.ShipmentSlaDueAt.Time
	}
	var issueSlaDueAt *time.Time
	if row.IssueSlaDueAt.Valid {
		issueSlaDueAt = &row.IssueSlaDueAt.Time
	}
	return entity.RoutedOrder{
		ID:                     row.ID,
		CandidateID:            row.CandidateID,
		ProductTitle:           row.ProductTitle,
		Partner:                row.Partner,
		Quantity:               row.Quantity,
		Total:                  row.Total,
		CustomerName:           row.CustomerName,
		Status:                 row.Status,
		Timeline:               timeline,
		ExceptionType:          row.ExceptionType,
		ExceptionStatus:        row.ExceptionStatus,
		ShipmentStatus:         row.ShipmentStatus,
		ShipmentCarrier:        row.ShipmentCarrier,
		ShipmentTrackingNumber: row.ShipmentTrackingNumber,
		ShipmentTrackingURL:    row.ShipmentTrackingURL,
		ShipmentNotes:          row.ShipmentNotes,
		OperatorAssignee:       row.OperatorAssignee,
		ShipmentSlaDueAt:       shipmentSlaDueAt,
		IssueSlaDueAt:          issueSlaDueAt,
		BaseCostSnapshot:       row.BaseCostSnapshot,
		FulfillmentCost:        row.FulfillmentCost,
		ShippingCost:           row.ShippingCost,
		IssueCost:              row.IssueCost,
		IssueResolution:        row.IssueResolution,
		IssueNotes:             row.IssueNotes,
		RealizedMargin:         row.RealizedMargin,
		SettlementStatus:       row.SettlementStatus,
		SettlementNotes:        row.SettlementNotes,
		ShippedAt:              shippedAt,
		DeliveredAt:            deliveredAt,
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}, nil
}
