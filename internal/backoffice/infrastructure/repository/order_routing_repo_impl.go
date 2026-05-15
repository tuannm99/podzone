package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
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
	ActivityLogJSON        string       `db:"activity_log_json"`
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

type routedOrderActivityRow struct {
	ID               int64     `db:"id"`
	OrderID          string    `db:"order_id"`
	ProductTitle     string    `db:"product_title"`
	OperatorAssignee string    `db:"operator_assignee"`
	ActivityType     string    `db:"activity_type"`
	Actor            string    `db:"actor"`
	Message          string    `db:"message"`
	DetailsJSON      string    `db:"details_json"`
	CreatedAt        time.Time `db:"created_at"`
}

func (r *OrderRoutingRepositoryImpl) List(ctx context.Context) ([]entity.RoutedOrder, error) {
	query, args, err := psql.
		Select("id", "candidate_id", "product_title", "partner", "quantity", "total", "customer_name", "status", "timeline_json", "activity_log_json", "exception_type", "exception_status", "shipment_status", "shipment_carrier", "shipment_tracking_number", "shipment_tracking_url", "shipment_notes", "operator_assignee", "shipment_sla_due_at", "issue_sla_due_at", "base_cost_snapshot", "fulfillment_cost", "shipping_cost", "issue_cost", "issue_resolution", "issue_notes", "realized_margin", "settlement_status", "settlement_notes", "shipped_at", "delivered_at", "created_at", "updated_at").
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
		Select("id", "candidate_id", "product_title", "partner", "quantity", "total", "customer_name", "status", "timeline_json", "activity_log_json", "exception_type", "exception_status", "shipment_status", "shipment_carrier", "shipment_tracking_number", "shipment_tracking_url", "shipment_notes", "operator_assignee", "shipment_sla_due_at", "issue_sla_due_at", "base_cost_snapshot", "fulfillment_cost", "shipping_cost", "issue_cost", "issue_resolution", "issue_notes", "realized_margin", "settlement_status", "settlement_notes", "shipped_at", "delivered_at", "created_at", "updated_at").
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

func (r *OrderRoutingRepositoryImpl) ListActivityFeed(ctx context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	builder := psql.
		Select("id", "order_id", "product_title", "operator_assignee", "activity_type", "actor", "message", "details_json", "created_at").
		From("routed_order_activities")
	countBuilder := psql.Select("COUNT(*)").From("routed_order_activities")
	applyFilters := func(b sq.SelectBuilder) sq.SelectBuilder {
		if query.ActivityType == "notes" {
			b = b.Where(sq.NotEq{"activity_type": entity.RoutedOrderActivityTypeSystem})
		} else if query.ActivityType != "" && query.ActivityType != "all" {
			b = b.Where(sq.Eq{"activity_type": query.ActivityType})
		}
		if !query.IncludeSystem && query.ActivityType != "system" && query.ActivityType != "notes" {
			b = b.Where(sq.NotEq{"activity_type": entity.RoutedOrderActivityTypeSystem})
		}
		if query.Since != nil {
			b = b.Where(sq.GtOrEq{"created_at": query.Since.UTC()})
		}
		if actor := strings.ToLower(strings.TrimSpace(query.ActorContains)); actor != "" {
			b = b.Where("LOWER(actor) LIKE ?", "%"+actor+"%")
		}
		return b
	}
	builder = applyFilters(builder)
	countBuilder = applyFilters(countBuilder)

	if afterID, afterCreatedAt, ok := decodeActivityCursor(query.After); ok {
		builder = builder.Where(
			sq.Or{
				sq.Lt{"created_at": afterCreatedAt},
				sq.Expr("(created_at = ? AND id < ?)", afterCreatedAt, afterID),
			},
		)
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	querySQL, args, err := builder.
		OrderBy("created_at DESC", "id DESC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []routedOrderActivityRow
	var total int
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureRoutedOrderTables(ctx, tx); err != nil {
			return err
		}
		if err := tx.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
			return err
		}
		return tx.SelectContext(ctx, &rows, querySQL, args...)
	}); err != nil {
		return nil, err
	}

	entries := make([]entity.RoutedOrderActivityFeedEntry, 0, len(rows))
	for _, row := range rows {
		entry, err := mapRoutedOrderActivityRow(row)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	var nextCursor *string
	if len(rows) == limit {
		value := encodeActivityCursor(rows[len(rows)-1].ID, rows[len(rows)-1].CreatedAt)
		nextCursor = &value
	}
	return &entity.RoutedOrderActivityFeedPage{
		Entries:    entries,
		Total:      total,
		NextCursor: nextCursor,
	}, nil
}

func (r *OrderRoutingRepositoryImpl) Create(ctx context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error) {
	timelineJSON, err := json.Marshal(order.Timeline)
	if err != nil {
		return nil, err
	}
	activityLogJSON, err := json.Marshal(order.ActivityLog)
	if err != nil {
		return nil, err
	}
	query, args, err := psql.
		Insert("routed_orders").
		Columns("id", "candidate_id", "product_title", "partner", "quantity", "total", "customer_name", "status", "timeline_json", "activity_log_json", "exception_type", "exception_status", "shipment_status", "shipment_carrier", "shipment_tracking_number", "shipment_tracking_url", "shipment_notes", "operator_assignee", "shipment_sla_due_at", "issue_sla_due_at", "base_cost_snapshot", "fulfillment_cost", "shipping_cost", "issue_cost", "issue_resolution", "issue_notes", "realized_margin", "settlement_status", "settlement_notes", "shipped_at", "delivered_at", "created_at", "updated_at").
		Values(order.ID, order.CandidateID, order.ProductTitle, order.Partner, order.Quantity, order.Total, order.CustomerName, order.Status, string(timelineJSON), string(activityLogJSON), order.ExceptionType, order.ExceptionStatus, order.ShipmentStatus, order.ShipmentCarrier, order.ShipmentTrackingNumber, order.ShipmentTrackingURL, order.ShipmentNotes, order.OperatorAssignee, order.ShipmentSlaDueAt, order.IssueSlaDueAt, order.BaseCostSnapshot, order.FulfillmentCost, order.ShippingCost, order.IssueCost, order.IssueResolution, order.IssueNotes, order.RealizedMargin, order.SettlementStatus, order.SettlementNotes, order.ShippedAt, order.DeliveredAt, order.CreatedAt, order.UpdatedAt).
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureRoutedOrderTables(ctx, tx); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
		return insertOrderActivities(ctx, tx, order.ID, order.ProductTitle, order.OperatorAssignee, order.ActivityLog)
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
	activityLogJSON, err := json.Marshal(order.ActivityLog)
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
		Set("activity_log_json", string(activityLogJSON)).
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
		var existingLogJSON string
		if err := tx.GetContext(ctx, &existingLogJSON, `SELECT activity_log_json FROM routed_orders WHERE id = $1`, order.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("routed order not found")
			}
			return err
		}
		var existingLog []entity.RoutedOrderActivity
		if existingLogJSON != "" {
			if err := json.Unmarshal([]byte(existingLogJSON), &existingLog); err != nil {
				return err
			}
		}
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("routed order not found")
		}
		if len(order.ActivityLog) > len(existingLog) {
			return insertOrderActivities(ctx, tx, order.ID, order.ProductTitle, order.OperatorAssignee, order.ActivityLog[len(existingLog):])
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
	var activityLog []entity.RoutedOrderActivity
	if row.ActivityLogJSON == "" {
		row.ActivityLogJSON = "[]"
	}
	if err := json.Unmarshal([]byte(row.ActivityLogJSON), &activityLog); err != nil {
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
		ActivityLog:            activityLog,
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

func mapRoutedOrderActivityRow(row routedOrderActivityRow) (entity.RoutedOrderActivityFeedEntry, error) {
	var details []entity.RoutedOrderActivityDetail
	if row.DetailsJSON == "" {
		row.DetailsJSON = "[]"
	}
	if err := json.Unmarshal([]byte(row.DetailsJSON), &details); err != nil {
		return entity.RoutedOrderActivityFeedEntry{}, err
	}
	return entity.RoutedOrderActivityFeedEntry{
		OrderID:          row.OrderID,
		ProductTitle:     row.ProductTitle,
		OperatorAssignee: row.OperatorAssignee,
		Activity: entity.RoutedOrderActivity{
			Type:      row.ActivityType,
			Actor:     row.Actor,
			Message:   row.Message,
			Details:   details,
			CreatedAt: row.CreatedAt,
		},
	}, nil
}

func insertOrderActivities(ctx context.Context, tx *sqlx.Tx, orderID, productTitle, operatorAssignee string, activities []entity.RoutedOrderActivity) error {
	for _, activity := range activities {
		detailsJSON, err := json.Marshal(activity.Details)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO routed_order_activities (order_id, product_title, operator_assignee, activity_type, actor, message, details_json, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			orderID,
			productTitle,
			operatorAssignee,
			activity.Type,
			activity.Actor,
			activity.Message,
			string(detailsJSON),
			activity.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func encodeActivityCursor(id int64, createdAt time.Time) string {
	return base64.StdEncoding.EncodeToString([]byte(createdAt.UTC().Format(time.RFC3339Nano) + "|" + strconv.FormatInt(id, 10)))
}

func decodeActivityCursor(cursor string) (int64, time.Time, bool) {
	if strings.TrimSpace(cursor) == "" {
		return 0, time.Time{}, false
	}
	raw, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, time.Time{}, false
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return 0, time.Time{}, false
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return 0, time.Time{}, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, time.Time{}, false
	}
	return id, createdAt, true
}
