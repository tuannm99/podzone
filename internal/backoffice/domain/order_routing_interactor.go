package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
)

type OrderRoutingInteractor struct {
	orders   outputport.OrderRoutingRepository
	products outputport.ProductSetupRepository
}

func NewOrderRoutingInteractor(
	orders outputport.OrderRoutingRepository,
	products outputport.ProductSetupRepository,
) inputport.OrderRoutingUsecase {
	return &OrderRoutingInteractor{orders: orders, products: products}
}

func (i *OrderRoutingInteractor) ListRoutedOrders(ctx context.Context) ([]entity.RoutedOrder, error) {
	return i.orders.List(ctx)
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

	qty := cmd.Quantity
	if qty < 1 {
		qty = 1
	}
	customerName := strings.TrimSpace(cmd.CustomerName)
	if customerName == "" {
		customerName = "Sample customer"
	}
	now := time.Now().UTC()
	order := entity.RoutedOrder{
		ID:               fmt.Sprintf("ORD-%s", strings.ToUpper(uuid.NewString()[:8])),
		CandidateID:      candidate.ID,
		ProductTitle:     candidate.Title,
		Partner:          candidate.Partner,
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
	order.Timeline = append(order.Timeline, routingTimelineEntry(nextStatus, order.Partner))
	order.UpdatedAt = time.Now().UTC()
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
	order.Timeline = append(order.Timeline, fmt.Sprintf("Exception opened: %s", strings.ReplaceAll(exceptionType, "_", " ")))
	order.UpdatedAt = time.Now().UTC()
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
	order.Timeline = append(order.Timeline, fmt.Sprintf("Exception %s: %s", status, strings.ReplaceAll(order.ExceptionType, "_", " ")))
	order.UpdatedAt = time.Now().UTC()
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
	order.ShipmentStatus = shipmentStatus
	order.ShipmentCarrier = strings.TrimSpace(cmd.Carrier)
	order.ShipmentTrackingNumber = strings.TrimSpace(cmd.TrackingNumber)
	order.ShipmentTrackingURL = strings.TrimSpace(cmd.TrackingURL)
	order.ShipmentNotes = strings.TrimSpace(cmd.Notes)
	switch shipmentStatus {
	case entity.RoutedOrderShipmentStatusInTransit:
		if order.Status != entity.RoutedOrderStatusShipped {
			order.Status = entity.RoutedOrderStatusShipped
			order.Timeline = append(order.Timeline, routingTimelineEntry(entity.RoutedOrderStatusShipped, order.Partner))
		}
		if order.ShippedAt == nil {
			order.ShippedAt = &now
		}
		order.DeliveredAt = nil
	case entity.RoutedOrderShipmentStatusDelivered:
		if order.Status != entity.RoutedOrderStatusShipped {
			order.Status = entity.RoutedOrderStatusShipped
			order.Timeline = append(order.Timeline, routingTimelineEntry(entity.RoutedOrderStatusShipped, order.Partner))
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

	order.Timeline = append(order.Timeline, shipmentTimelineEntry(order))
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
	order.Timeline = append(order.Timeline, queueControlTimelineEntry(order))
	order.UpdatedAt = time.Now().UTC()
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

		order.Timeline = append(order.Timeline, bulkUpdateTimelineEntry(order, cmd))
		order.UpdatedAt = time.Now().UTC()
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
	order.Timeline = append(order.Timeline, settlementTimelineEntry(order))
	order.UpdatedAt = time.Now().UTC()
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
	order.Timeline = append(order.Timeline, issueHandlingTimelineEntry(order))
	order.UpdatedAt = time.Now().UTC()
	return i.orders.Update(ctx, *order)
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
