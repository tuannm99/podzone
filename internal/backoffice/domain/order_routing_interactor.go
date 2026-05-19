package interactor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type OrderRoutingInteractor struct {
	orders   outputport.OrderRoutingRepository
	products outputport.ProductSetupRepository
	partners outputport.PartnerDirectory
}

func NewOrderRoutingInteractor(
	orders outputport.OrderRoutingRepository,
	products outputport.ProductSetupRepository,
	partners outputport.PartnerDirectory,
) inputport.OrderRoutingUsecase {
	return &OrderRoutingInteractor{orders: orders, products: products, partners: partners}
}

func (i *OrderRoutingInteractor) ListRoutedOrders(ctx context.Context) ([]entity.RoutedOrder, error) {
	return i.orders.List(ctx)
}

func (i *OrderRoutingInteractor) ListRoutedOrderActivities(
	ctx context.Context,
	query inputport.ListRoutedOrderActivitiesQuery,
) (*entity.RoutedOrderActivityFeedPage, error) {
	return i.orders.ListActivityFeed(ctx, query)
}

func (i *OrderRoutingInteractor) RecommendRoutedOrderPartner(
	ctx context.Context,
	query inputport.RecommendRoutedOrderPartnerQuery,
) (*entity.RoutedOrderRecommendation, error) {
	candidateID := strings.TrimSpace(query.CandidateID)
	if candidateID == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	candidate, err := i.products.GetCandidateByID(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, fmt.Errorf("product candidate not found")
	}
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	partners, err := i.partners.ListActivePartners(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return buildRoutingRecommendation(
		candidate,
		partners,
		normalizeRoutingLabel(query.ProductType),
		normalizeRoutingLabel(query.ShipRegion),
		strings.TrimSpace(query.PreferredPartner),
	), nil
}

func (i *OrderRoutingInteractor) CreateRoutedOrder(
	ctx context.Context,
	cmd inputport.CreateRoutedOrderCmd,
) (*entity.RoutedOrder, error) {
	candidateID := strings.TrimSpace(cmd.CandidateID)
	if candidateID == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	candidate, err := i.products.GetCandidateByID(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if candidate == nil || candidate.Status != entity.ProductSetupCandidateStatusPublishedMock {
		return nil, fmt.Errorf("published mock product candidate is required")
	}
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	partners, err := i.partners.ListActivePartners(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	recommendation := buildRoutingRecommendation(
		candidate,
		partners,
		normalizeRoutingLabel(cmd.ProductType),
		normalizeRoutingLabel(cmd.ShipRegion),
		strings.TrimSpace(cmd.PreferredPartner),
	)
	if recommendation.SelectedPartner == "" {
		return nil, fmt.Errorf(
			"no eligible active partner found for %s in %s",
			fallbackRoutingLabel(recommendation.ProductType, "product type"),
			fallbackRoutingLabel(recommendation.ShipRegion, "target region"),
		)
	}

	qty := cmd.Quantity
	if qty < 1 {
		qty = 1
	}
	customerName := strings.TrimSpace(cmd.CustomerName)
	if customerName == "" {
		customerName = "Sample customer"
	}
	now := time.Now().UTC()
	actor := activityActorFromContext(ctx)
	order := entity.RoutedOrder{
		ID:               fmt.Sprintf("ORD-%s", strings.ToUpper(uuid.NewString()[:8])),
		CandidateID:      candidate.ID,
		ProductTitle:     candidate.Title,
		Partner:          recommendation.SelectedPartner,
		Quantity:         qty,
		Total:            multiplyMoney(candidate.RetailPrice, qty),
		CustomerName:     customerName,
		Status:           entity.RoutedOrderStatusQueued,
		ShipmentStatus:   entity.RoutedOrderShipmentStatusAwaitingLabel,
		OperatorAssignee: "unassigned",
		BaseCostSnapshot: multiplyMoney(candidate.BaseCost, qty),
		FulfillmentCost:  multiplyMoney(candidate.BaseCost, qty),
		ShippingCost:     "$0.00",
		IssueCost:        "$0.00",
		IssueResolution:  entity.RoutedOrderIssueResolutionMonitor,
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
		Timeline: []string{
			fmt.Sprintf("Order created for %s", candidate.Title),
			routingTimelineEntry(entity.RoutedOrderStatusQueued, candidate.Partner),
		},
		ActivityLog: []entity.RoutedOrderActivity{
			newActivity(
				entity.RoutedOrderActivityTypeSystem,
				actor,
				fmt.Sprintf("Order created for %s", candidate.Title),
				now,
				activityDetails(
					"candidate_id", candidate.ID,
					"quantity", fmt.Sprintf("%d", qty),
					"status", entity.RoutedOrderStatusQueued,
					"product_type", recommendation.ProductType,
					"ship_region", recommendation.ShipRegion,
				),
			),
			newActivity(
				entity.RoutedOrderActivityTypeSystem,
				actor,
				routingTimelineEntry(entity.RoutedOrderStatusQueued, recommendation.SelectedPartner),
				now,
				activityDetails(
					"status", entity.RoutedOrderStatusQueued,
					"partner", recommendation.SelectedPartner,
					"routing_summary", recommendation.Summary,
					"candidate_partner", recommendation.CandidatePartner,
				),
			),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	order.RealizedMargin = calculateMargin(order.Total, order.FulfillmentCost, order.ShippingCost, order.IssueCost)
	return i.orders.Create(ctx, order)
}

func (i *OrderRoutingInteractor) AdvanceRoutedOrder(ctx context.Context, orderID string) (*entity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}
	if order.ExceptionStatus == entity.RoutedOrderExceptionStatusOpen ||
		order.ExceptionStatus == entity.RoutedOrderExceptionStatusEscalated {
		return nil, fmt.Errorf("resolve the active exception before advancing the routed order")
	}

	nextStatus := order.Status
	switch order.Status {
	case entity.RoutedOrderStatusQueued:
		nextStatus = entity.RoutedOrderStatusInProduction
	case entity.RoutedOrderStatusInProduction:
		nextStatus = entity.RoutedOrderStatusShipped
	case entity.RoutedOrderStatusShipped:
		nextStatus = entity.RoutedOrderStatusShipped
	default:
		return nil, fmt.Errorf("invalid routed order status")
	}
	if nextStatus == order.Status && order.Status == entity.RoutedOrderStatusShipped {
		return order, nil
	}

	order.Status = nextStatus
	entry := routingTimelineEntry(nextStatus, order.Partner)
	now := time.Now().UTC()
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		entity.RoutedOrderActivityTypeSystem,
		activityActorFromContext(ctx),
		entry,
		now,
		activityDetails("status", nextStatus, "partner", order.Partner),
	))
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) OpenOrderException(
	ctx context.Context,
	cmd inputport.OpenOrderExceptionCmd,
) (*entity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}
	exceptionType := normalizeExceptionType(cmd.ExceptionType)
	if exceptionType == "" {
		return nil, fmt.Errorf("invalid exception type")
	}
	if order.ExceptionStatus == entity.RoutedOrderExceptionStatusOpen {
		return order, nil
	}

	order.ExceptionType = exceptionType
	order.ExceptionStatus = entity.RoutedOrderExceptionStatusOpen
	entry := fmt.Sprintf("Exception opened: %s", strings.ReplaceAll(exceptionType, "_", " "))
	now := time.Now().UTC()
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		entity.RoutedOrderActivityTypeSystem,
		activityActorFromContext(ctx),
		entry,
		now,
		activityDetails("exception_type", exceptionType, "exception_status", entity.RoutedOrderExceptionStatusOpen),
	))
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderExceptionStatus(
	ctx context.Context,
	cmd inputport.UpdateOrderExceptionStatusCmd,
) (*entity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}
	if order.ExceptionType == "" {
		return nil, fmt.Errorf("routed order has no active exception type")
	}

	status := normalizeExceptionStatus(cmd.Status)
	if status == "" {
		return nil, fmt.Errorf("invalid exception status")
	}
	order.ExceptionStatus = status
	entry := fmt.Sprintf("Exception %s: %s", status, strings.ReplaceAll(order.ExceptionType, "_", " "))
	now := time.Now().UTC()
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		entity.RoutedOrderActivityTypeSystem,
		activityActorFromContext(ctx),
		entry,
		now,
		activityDetails("exception_type", order.ExceptionType, "exception_status", status),
	))
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderShipment(
	ctx context.Context,
	cmd inputport.UpdateOrderShipmentCmd,
) (*entity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}

	shipmentStatus := normalizeShipmentStatus(cmd.ShipmentStatus)
	if shipmentStatus == "" {
		return nil, fmt.Errorf("invalid shipment status")
	}

	now := time.Now().UTC()
	actor := activityActorFromContext(ctx)
	order.ShipmentStatus = shipmentStatus
	order.ShipmentCarrier = strings.TrimSpace(cmd.Carrier)
	order.ShipmentTrackingNumber = strings.TrimSpace(cmd.TrackingNumber)
	order.ShipmentTrackingURL = strings.TrimSpace(cmd.TrackingURL)
	order.ShipmentNotes = strings.TrimSpace(cmd.Notes)
	switch shipmentStatus {
	case entity.RoutedOrderShipmentStatusInTransit:
		if order.Status != entity.RoutedOrderStatusShipped {
			order.Status = entity.RoutedOrderStatusShipped
			entry := routingTimelineEntry(entity.RoutedOrderStatusShipped, order.Partner)
			order.Timeline = append(order.Timeline, entry)
			order.ActivityLog = append(order.ActivityLog, newActivity(
				entity.RoutedOrderActivityTypeSystem,
				actor,
				entry,
				now,
				activityDetails("status", entity.RoutedOrderStatusShipped, "partner", order.Partner),
			))
		}
		if order.ShippedAt == nil {
			order.ShippedAt = &now
		}
		order.DeliveredAt = nil
	case entity.RoutedOrderShipmentStatusDelivered:
		if order.Status != entity.RoutedOrderStatusShipped {
			order.Status = entity.RoutedOrderStatusShipped
			entry := routingTimelineEntry(entity.RoutedOrderStatusShipped, order.Partner)
			order.Timeline = append(order.Timeline, entry)
			order.ActivityLog = append(order.ActivityLog, newActivity(
				entity.RoutedOrderActivityTypeSystem,
				actor,
				entry,
				now,
				activityDetails("status", entity.RoutedOrderStatusShipped, "partner", order.Partner),
			))
		}
		if order.ShippedAt == nil {
			order.ShippedAt = &now
		}
		order.DeliveredAt = &now
	case entity.RoutedOrderShipmentStatusAwaitingLabel,
		entity.RoutedOrderShipmentStatusLabelReady,
		entity.RoutedOrderShipmentStatusDeliveryIssue:
		if shipmentStatus != entity.RoutedOrderShipmentStatusDelivered {
			order.DeliveredAt = nil
		}
	}

	shipmentEntry := shipmentTimelineEntry(order)
	order.Timeline = append(order.Timeline, shipmentEntry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		entity.RoutedOrderActivityTypeSystem,
		actor,
		shipmentEntry,
		now,
		activityDetails(
			"shipment_status", order.ShipmentStatus,
			"carrier", fallbackShipmentCarrier(order.ShipmentCarrier),
			"tracking_number", order.ShipmentTrackingNumber,
		),
	))
	if order.ShipmentNotes != "" {
		order.ActivityLog = append(order.ActivityLog, newActivity(
			entity.RoutedOrderActivityTypeShipmentNote,
			actor,
			order.ShipmentNotes,
			now,
			activityDetails(
				"shipment_status", order.ShipmentStatus,
				"carrier", fallbackShipmentCarrier(order.ShipmentCarrier),
			),
		))
	}
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderQueueControl(
	ctx context.Context,
	cmd inputport.UpdateOrderQueueControlCmd,
) (*entity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}

	assignee := strings.TrimSpace(cmd.OperatorAssignee)
	if assignee == "" {
		assignee = "unassigned"
	}
	order.OperatorAssignee = assignee
	order.ShipmentSlaDueAt = cmd.ShipmentSlaDueAt
	order.IssueSlaDueAt = cmd.IssueSlaDueAt
	entry := queueControlTimelineEntry(order)
	now := time.Now().UTC()
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		entity.RoutedOrderActivityTypeSystem,
		activityActorFromContext(ctx),
		entry,
		now,
		activityDetails(
			"operator_assignee", order.OperatorAssignee,
			"shipment_sla_due_at", formatOptionalTime(order.ShipmentSlaDueAt),
			"issue_sla_due_at", formatOptionalTime(order.IssueSlaDueAt),
		),
	))
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) BulkUpdateRoutedOrders(
	ctx context.Context,
	cmd inputport.BulkUpdateRoutedOrdersCmd,
) ([]entity.RoutedOrder, error) {
	if len(cmd.OrderIDs) == 0 {
		return nil, fmt.Errorf("at least one routed order id is required")
	}
	if cmd.OperatorAssignee == nil && cmd.ShipmentSlaDueAt == nil && cmd.SettlementStatus == nil {
		return nil, fmt.Errorf("at least one bulk update field is required")
	}

	var normalizedSettlementStatus string
	if cmd.SettlementStatus != nil {
		normalizedSettlementStatus = normalizeSettlementStatus(strings.TrimSpace(*cmd.SettlementStatus))
		if normalizedSettlementStatus == "" {
			return nil, fmt.Errorf("invalid settlement status")
		}
	}

	updated := make([]entity.RoutedOrder, 0, len(cmd.OrderIDs))
	for _, rawID := range cmd.OrderIDs {
		orderID := strings.TrimSpace(rawID)
		if orderID == "" {
			continue
		}
		order, err := i.orders.GetByID(ctx, orderID)
		if err != nil {
			return nil, err
		}
		if order == nil {
			return nil, fmt.Errorf("routed order not found")
		}

		if cmd.OperatorAssignee != nil {
			assignee := strings.TrimSpace(*cmd.OperatorAssignee)
			if assignee == "" {
				assignee = "unassigned"
			}
			order.OperatorAssignee = assignee
		}
		if cmd.ShipmentSlaDueAt != nil {
			order.ShipmentSlaDueAt = cmd.ShipmentSlaDueAt
		}
		if cmd.SettlementStatus != nil {
			order.SettlementStatus = normalizedSettlementStatus
		}

		entry := bulkUpdateTimelineEntry(order, cmd)
		now := time.Now().UTC()
		order.Timeline = append(order.Timeline, entry)
		order.ActivityLog = append(order.ActivityLog, newActivity(
			entity.RoutedOrderActivityTypeSystem,
			activityActorFromContext(ctx),
			entry,
			now,
			activityDetails(
				"operator_assignee", order.OperatorAssignee,
				"shipment_sla_due_at", formatOptionalTime(order.ShipmentSlaDueAt),
				"settlement_status", order.SettlementStatus,
			),
		))
		order.UpdatedAt = now
		saved, err := i.orders.Update(ctx, *order)
		if err != nil {
			return nil, err
		}
		updated = append(updated, *saved)
	}
	return updated, nil
}

func (i *OrderRoutingInteractor) UpdateOrderSettlement(
	ctx context.Context,
	cmd inputport.UpdateOrderSettlementCmd,
) (*entity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}

	fulfillmentCost, err := normalizeMoney(cmd.FulfillmentCost)
	if err != nil {
		return nil, fmt.Errorf("invalid fulfillment cost")
	}
	shippingCost, err := normalizeMoney(cmd.ShippingCost)
	if err != nil {
		return nil, fmt.Errorf("invalid shipping cost")
	}
	settlementStatus := normalizeSettlementStatus(cmd.SettlementStatus)
	if settlementStatus == "" {
		return nil, fmt.Errorf("invalid settlement status")
	}

	order.FulfillmentCost = fulfillmentCost
	order.ShippingCost = shippingCost
	order.RealizedMargin = calculateMargin(order.Total, fulfillmentCost, shippingCost, order.IssueCost)
	order.SettlementStatus = settlementStatus
	order.SettlementNotes = strings.TrimSpace(cmd.Notes)
	entry := settlementTimelineEntry(order)
	now := time.Now().UTC()
	actor := activityActorFromContext(ctx)
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		entity.RoutedOrderActivityTypeSystem,
		actor,
		entry,
		now,
		activityDetails(
			"settlement_status", order.SettlementStatus,
			"fulfillment_cost", order.FulfillmentCost,
			"shipping_cost", order.ShippingCost,
			"realized_margin", order.RealizedMargin,
		),
	))
	if order.SettlementNotes != "" {
		order.ActivityLog = append(order.ActivityLog, newActivity(
			entity.RoutedOrderActivityTypeSettlementNote,
			actor,
			order.SettlementNotes,
			now,
			activityDetails("settlement_status", order.SettlementStatus, "realized_margin", order.RealizedMargin),
		))
	}
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderIssueHandling(
	ctx context.Context,
	cmd inputport.UpdateOrderIssueHandlingCmd,
) (*entity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}
	if order.ExceptionType == "" && order.ShipmentStatus != entity.RoutedOrderShipmentStatusDeliveryIssue {
		return nil, fmt.Errorf("issue cost handling requires an active exception or delivery issue")
	}

	issueCost, err := normalizeMoney(cmd.IssueCost)
	if err != nil {
		return nil, fmt.Errorf("invalid issue cost")
	}
	issueResolution := normalizeIssueResolution(cmd.IssueResolution)
	if issueResolution == "" {
		return nil, fmt.Errorf("invalid issue resolution")
	}

	order.IssueCost = issueCost
	order.IssueResolution = issueResolution
	order.IssueNotes = strings.TrimSpace(cmd.Notes)
	order.RealizedMargin = calculateMargin(order.Total, order.FulfillmentCost, order.ShippingCost, order.IssueCost)
	entry := issueHandlingTimelineEntry(order)
	now := time.Now().UTC()
	actor := activityActorFromContext(ctx)
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		entity.RoutedOrderActivityTypeSystem,
		actor,
		entry,
		now,
		activityDetails(
			"issue_resolution", order.IssueResolution,
			"issue_cost", order.IssueCost,
			"realized_margin", order.RealizedMargin,
		),
	))
	if order.IssueNotes != "" {
		order.ActivityLog = append(order.ActivityLog, newActivity(
			entity.RoutedOrderActivityTypeIssueNote,
			actor,
			order.IssueNotes,
			now,
			activityDetails("issue_resolution", order.IssueResolution, "issue_cost", order.IssueCost),
		))
	}
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func newActivity(
	activityType,
	actor,
	message string,
	createdAt time.Time,
	details []entity.RoutedOrderActivityDetail,
) entity.RoutedOrderActivity {
	return entity.RoutedOrderActivity{
		Type:      activityType,
		Actor:     strings.TrimSpace(actor),
		Message:   strings.TrimSpace(message),
		Details:   details,
		CreatedAt: createdAt,
	}
}

func activityActorFromContext(ctx context.Context) string {
	userID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return "system"
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "system"
	}
	return "user:" + userID
}

func activityDetails(pairs ...string) []entity.RoutedOrderActivityDetail {
	details := make([]entity.RoutedOrderActivityDetail, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key := strings.TrimSpace(pairs[i])
		value := strings.TrimSpace(pairs[i+1])
		if key == "" || value == "" {
			continue
		}
		details = append(details, entity.RoutedOrderActivityDetail{
			Key:   key,
			Value: value,
		})
	}
	return details
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func routingTimelineEntry(status, partner string) string {
	switch status {
	case entity.RoutedOrderStatusInProduction:
		return fmt.Sprintf("Sent to %s for POD production", partner)
	case entity.RoutedOrderStatusShipped:
		return fmt.Sprintf("Marked as shipped from %s", partner)
	default:
		return fmt.Sprintf("Queued for %s", partner)
	}
}

func normalizeExceptionType(raw string) string {
	switch strings.TrimSpace(raw) {
	case "artwork_issue", "partner_delay", "address_hold", "reprint_request":
		return raw
	default:
		return ""
	}
}

func normalizeExceptionStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case entity.RoutedOrderExceptionStatusOpen,
		entity.RoutedOrderExceptionStatusEscalated,
		entity.RoutedOrderExceptionStatusResolved:
		return raw
	default:
		return ""
	}
}

func normalizeShipmentStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case entity.RoutedOrderShipmentStatusAwaitingLabel,
		entity.RoutedOrderShipmentStatusLabelReady,
		entity.RoutedOrderShipmentStatusInTransit,
		entity.RoutedOrderShipmentStatusDelivered,
		entity.RoutedOrderShipmentStatusDeliveryIssue:
		return raw
	default:
		return ""
	}
}

func normalizeSettlementStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case entity.RoutedOrderSettlementStatusPending,
		entity.RoutedOrderSettlementStatusReconciled,
		entity.RoutedOrderSettlementStatusPaid,
		entity.RoutedOrderSettlementStatusDisputed:
		return raw
	default:
		return ""
	}
}

func normalizeIssueResolution(raw string) string {
	switch strings.TrimSpace(raw) {
	case entity.RoutedOrderIssueResolutionMonitor,
		entity.RoutedOrderIssueResolutionReprint,
		entity.RoutedOrderIssueResolutionRefund,
		entity.RoutedOrderIssueResolutionCarrierClaim,
		entity.RoutedOrderIssueResolutionAddressRetry:
		return raw
	default:
		return ""
	}
}

func shipmentTimelineEntry(order *entity.RoutedOrder) string {
	switch order.ShipmentStatus {
	case entity.RoutedOrderShipmentStatusLabelReady:
		return fmt.Sprintf("Shipment prepared with %s", fallbackShipmentCarrier(order.ShipmentCarrier))
	case entity.RoutedOrderShipmentStatusInTransit:
		return fmt.Sprintf(
			"Shipment is in transit via %s%s",
			fallbackShipmentCarrier(order.ShipmentCarrier),
			fallbackTrackingSuffix(order.ShipmentTrackingNumber),
		)
	case entity.RoutedOrderShipmentStatusDelivered:
		return "Shipment marked delivered by store operator"
	case entity.RoutedOrderShipmentStatusDeliveryIssue:
		return "Shipment issue flagged for manual follow-up"
	default:
		return "Shipment is awaiting label assignment"
	}
}

func settlementTimelineEntry(order *entity.RoutedOrder) string {
	switch order.SettlementStatus {
	case entity.RoutedOrderSettlementStatusPaid:
		return fmt.Sprintf("Settlement marked paid with realized margin %s", order.RealizedMargin)
	case entity.RoutedOrderSettlementStatusDisputed:
		return "Settlement flagged for manual dispute follow-up"
	case entity.RoutedOrderSettlementStatusReconciled:
		return fmt.Sprintf("Settlement reconciled with realized margin %s", order.RealizedMargin)
	default:
		return fmt.Sprintf("Settlement remains pending with current margin %s", order.RealizedMargin)
	}
}

func issueHandlingTimelineEntry(order *entity.RoutedOrder) string {
	return fmt.Sprintf(
		"Issue handling updated: %s with impact %s",
		strings.ReplaceAll(order.IssueResolution, "_", " "),
		order.IssueCost,
	)
}

func queueControlTimelineEntry(order *entity.RoutedOrder) string {
	shipmentDue := "none"
	if order.ShipmentSlaDueAt != nil {
		shipmentDue = order.ShipmentSlaDueAt.UTC().Format(time.RFC3339)
	}
	issueDue := "none"
	if order.IssueSlaDueAt != nil {
		issueDue = order.IssueSlaDueAt.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf(
		"Queue ownership updated: %s · shipment SLA %s · issue SLA %s",
		order.OperatorAssignee,
		shipmentDue,
		issueDue,
	)
}

func bulkUpdateTimelineEntry(order *entity.RoutedOrder, cmd inputport.BulkUpdateRoutedOrdersCmd) string {
	parts := make([]string, 0, 3)
	if cmd.OperatorAssignee != nil {
		parts = append(parts, fmt.Sprintf("owner %s", order.OperatorAssignee))
	}
	if cmd.ShipmentSlaDueAt != nil {
		parts = append(parts, fmt.Sprintf("shipment SLA %s", order.ShipmentSlaDueAt.UTC().Format(time.RFC3339)))
	}
	if cmd.SettlementStatus != nil {
		parts = append(parts, fmt.Sprintf("settlement %s", order.SettlementStatus))
	}
	return fmt.Sprintf("Bulk queue update applied: %s", strings.Join(parts, " · "))
}

func fallbackShipmentCarrier(carrier string) string {
	carrier = strings.TrimSpace(carrier)
	if carrier == "" {
		return "manual carrier"
	}
	return carrier
}

func fallbackTrackingSuffix(trackingNumber string) string {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return ""
	}
	return fmt.Sprintf(" (%s)", trackingNumber)
}

func buildRoutingRecommendation(
	candidate *entity.ProductSetupCandidate,
	partners []entity.PartnerRoutingProfile,
	productType string,
	shipRegion string,
	preferredPartner string,
) *entity.RoutedOrderRecommendation {
	recommendation := &entity.RoutedOrderRecommendation{
		CandidateID:      candidate.ID,
		ProductTitle:     candidate.Title,
		CandidatePartner: strings.TrimSpace(candidate.Partner),
		ProductType:      productType,
		ShipRegion:       shipRegion,
		Options:          make([]entity.RoutingPartnerOption, 0, len(partners)),
	}

	for _, partner := range partners {
		eligible, reason := evaluatePartnerEligibility(partner, productType, shipRegion)
		recommendation.Options = append(recommendation.Options, entity.RoutingPartnerOption{
			Partner:  partner,
			Eligible: eligible,
			Reason:   reason,
		})
	}

	sort.SliceStable(recommendation.Options, func(i, j int) bool {
		left := recommendation.Options[i]
		right := recommendation.Options[j]
		if left.Eligible != right.Eligible {
			return left.Eligible
		}
		if left.Partner.RoutingPriority != right.Partner.RoutingPriority {
			return left.Partner.RoutingPriority > right.Partner.RoutingPriority
		}
		if left.Partner.SLADays != right.Partner.SLADays {
			return left.Partner.SLADays < right.Partner.SLADays
		}
		return strings.ToLower(left.Partner.Name) < strings.ToLower(right.Partner.Name)
	})

	preferredPartner = strings.TrimSpace(preferredPartner)
	if preferredPartner != "" {
		for _, option := range recommendation.Options {
			if !option.Eligible {
				continue
			}
			if strings.EqualFold(option.Partner.Name, preferredPartner) ||
				strings.EqualFold(option.Partner.Code, preferredPartner) {
				recommendation.SelectedPartner = option.Partner.Name
				recommendation.Summary = fmt.Sprintf(
					"Preferred partner %s is eligible for %s in %s.",
					option.Partner.Name,
					fallbackRoutingLabel(productType, "any product"),
					fallbackRoutingLabel(shipRegion, "any region"),
				)
				return recommendation
			}
		}
		recommendation.Summary = fmt.Sprintf(
			"Preferred partner %s is not eligible for %s in %s.",
			preferredPartner,
			fallbackRoutingLabel(productType, "any product"),
			fallbackRoutingLabel(shipRegion, "any region"),
		)
	}

	candidatePartner := recommendation.CandidatePartner
	if candidatePartner != "" {
		for _, option := range recommendation.Options {
			if option.Eligible && strings.EqualFold(option.Partner.Name, candidatePartner) {
				recommendation.SelectedPartner = option.Partner.Name
				recommendation.Summary = fmt.Sprintf(
					"Candidate default partner %s remains eligible for %s in %s.",
					option.Partner.Name,
					fallbackRoutingLabel(productType, "any product"),
					fallbackRoutingLabel(shipRegion, "any region"),
				)
				return recommendation
			}
		}
	}

	for _, option := range recommendation.Options {
		if option.Eligible {
			recommendation.SelectedPartner = option.Partner.Name
			recommendation.Summary = fmt.Sprintf(
				"Recommended %s based on active capability match, routing priority %d, and %d-day SLA.",
				option.Partner.Name,
				option.Partner.RoutingPriority,
				option.Partner.SLADays,
			)
			return recommendation
		}
	}

	recommendation.Summary = fmt.Sprintf(
		"No active partner matches %s in %s.",
		fallbackRoutingLabel(productType, "the requested product"),
		fallbackRoutingLabel(shipRegion, "the requested region"),
	)
	return recommendation
}

func evaluatePartnerEligibility(
	partner entity.PartnerRoutingProfile,
	productType string,
	shipRegion string,
) (bool, string) {
	if partner.Status != "active" {
		return false, "partner is inactive"
	}
	if len(partner.SupportedProductTypes) > 0 && productType != "" &&
		!containsNormalized(partner.SupportedProductTypes, productType) {
		return false, fmt.Sprintf("does not support %s", productType)
	}
	if len(partner.SupportedRegions) > 0 && shipRegion != "" &&
		!containsNormalized(partner.SupportedRegions, shipRegion) {
		return false, fmt.Sprintf("does not ship to %s", shipRegion)
	}
	return true, fmt.Sprintf(
		"eligible with priority %d and %d-day SLA",
		partner.RoutingPriority,
		partner.SLADays,
	)
}

func containsNormalized(items []string, expected string) bool {
	normalizedExpected := normalizeRoutingLabel(expected)
	for _, item := range items {
		if normalizeRoutingLabel(item) == normalizedExpected {
			return true
		}
	}
	return false
}

func normalizeRoutingLabel(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func fallbackRoutingLabel(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func multiplyMoney(raw string, qty int) string {
	value, ok := parseMoney(raw)
	if !ok {
		return "TBD"
	}
	return formatMoney(value * float64(qty))
}

func calculateMargin(total, fulfillmentCost, shippingCost, issueCost string) string {
	totalValue, ok := parseMoney(total)
	if !ok {
		return "TBD"
	}
	fulfillmentValue, ok := parseMoney(fulfillmentCost)
	if !ok {
		return "TBD"
	}
	shippingValue, ok := parseMoney(shippingCost)
	if !ok {
		return "TBD"
	}
	issueValue, ok := parseMoney(issueCost)
	if !ok {
		return "TBD"
	}
	return formatMoney(totalValue - fulfillmentValue - shippingValue - issueValue)
}

func normalizeMoney(raw string) (string, error) {
	value, ok := parseMoney(raw)
	if !ok {
		return "", fmt.Errorf("invalid money")
	}
	return formatMoney(value), nil
}

func parseMoney(raw string) (float64, bool) {
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' {
			return r
		}
		return -1
	}, raw)
	if cleaned == "" {
		return 0, false
	}
	var value float64
	if _, err := fmt.Sscanf(cleaned, "%f", &value); err != nil {
		return 0, false
	}
	return value, true
}

func formatMoney(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}
