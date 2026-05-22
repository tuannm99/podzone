package routing

import (
	"context"
	"fmt"
	"strings"
	"time"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func (i *OrderRoutingInteractor) UpdateOrderShipment(
	ctx context.Context,
	cmd routinginputport.UpdateOrderShipmentCmd,
) (*routingentity.RoutedOrder, error) {
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
	case routingentity.RoutedOrderShipmentStatusInTransit:
		if order.Status != routingentity.RoutedOrderStatusShipped {
			order.Status = routingentity.RoutedOrderStatusShipped
			entry := routingTimelineEntry(routingentity.RoutedOrderStatusShipped, order.Partner)
			order.Timeline = append(order.Timeline, entry)
			order.ActivityLog = append(order.ActivityLog, newActivity(
				routingentity.RoutedOrderActivityTypeSystem,
				actor,
				entry,
				now,
				activityDetails("status", routingentity.RoutedOrderStatusShipped, "partner", order.Partner),
			))
		}
		if order.ShippedAt == nil {
			order.ShippedAt = &now
		}
		order.DeliveredAt = nil
	case routingentity.RoutedOrderShipmentStatusDelivered:
		if order.Status != routingentity.RoutedOrderStatusShipped {
			order.Status = routingentity.RoutedOrderStatusShipped
			entry := routingTimelineEntry(routingentity.RoutedOrderStatusShipped, order.Partner)
			order.Timeline = append(order.Timeline, entry)
			order.ActivityLog = append(order.ActivityLog, newActivity(
				routingentity.RoutedOrderActivityTypeSystem,
				actor,
				entry,
				now,
				activityDetails("status", routingentity.RoutedOrderStatusShipped, "partner", order.Partner),
			))
		}
		if order.ShippedAt == nil {
			order.ShippedAt = &now
		}
		order.DeliveredAt = &now
	case routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		routingentity.RoutedOrderShipmentStatusLabelReady,
		routingentity.RoutedOrderShipmentStatusDeliveryIssue:
		if shipmentStatus != routingentity.RoutedOrderShipmentStatusDelivered {
			order.DeliveredAt = nil
		}
	}

	shipmentEntry := shipmentTimelineEntry(order)
	order.Timeline = append(order.Timeline, shipmentEntry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		routingentity.RoutedOrderActivityTypeSystem,
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
			routingentity.RoutedOrderActivityTypeShipmentNote,
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
	cmd routinginputport.UpdateOrderQueueControlCmd,
) (*routingentity.RoutedOrder, error) {
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
		routingentity.RoutedOrderActivityTypeSystem,
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
	cmd routinginputport.BulkUpdateRoutedOrdersCmd,
) ([]routingentity.RoutedOrder, error) {
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

	updated := make([]routingentity.RoutedOrder, 0, len(cmd.OrderIDs))
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
			routingentity.RoutedOrderActivityTypeSystem,
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
	cmd routinginputport.UpdateOrderSettlementCmd,
) (*routingentity.RoutedOrder, error) {
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
		routingentity.RoutedOrderActivityTypeSystem,
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
			routingentity.RoutedOrderActivityTypeSettlementNote,
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
	cmd routinginputport.UpdateOrderIssueHandlingCmd,
) (*routingentity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}
	if order.ExceptionType == "" && order.ShipmentStatus != routingentity.RoutedOrderShipmentStatusDeliveryIssue {
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
		routingentity.RoutedOrderActivityTypeSystem,
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
			routingentity.RoutedOrderActivityTypeIssueNote,
			actor,
			order.IssueNotes,
			now,
			activityDetails("issue_resolution", order.IssueResolution, "issue_cost", order.IssueCost),
		))
	}
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
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
