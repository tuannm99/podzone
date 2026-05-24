package routing

import (
	"context"
	"fmt"
	"strings"
	"time"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
)

func (i *OrderRoutingInteractor) AdvanceRoutedOrder(
	ctx context.Context,
	orderID string,
) (*routingentity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}
	if order.ExceptionStatus == routingentity.RoutedOrderExceptionStatusOpen ||
		order.ExceptionStatus == routingentity.RoutedOrderExceptionStatusEscalated {
		return nil, fmt.Errorf("resolve the active exception before advancing the routed order")
	}

	var nextStatus string
	switch order.Status {
	case routingentity.RoutedOrderStatusRoutingBlocked:
		return nil, fmt.Errorf("resolve the routing block before advancing the routed order")
	case routingentity.RoutedOrderStatusQueued:
		nextStatus = routingentity.RoutedOrderStatusInProduction
	case routingentity.RoutedOrderStatusInProduction:
		nextStatus = routingentity.RoutedOrderStatusShipped
	case routingentity.RoutedOrderStatusShipped:
		nextStatus = routingentity.RoutedOrderStatusShipped
	default:
		return nil, fmt.Errorf("invalid routed order status")
	}
	if nextStatus == order.Status && order.Status == routingentity.RoutedOrderStatusShipped {
		return order, nil
	}

	order.Status = nextStatus
	entry := routingTimelineEntry(nextStatus, order.Partner)
	now := time.Now().UTC()
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		routingentity.RoutedOrderActivityTypeSystem,
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
	cmd routinginputport.OpenOrderExceptionCmd,
) (*routingentity.RoutedOrder, error) {
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
	if order.ExceptionStatus == routingentity.RoutedOrderExceptionStatusOpen {
		return order, nil
	}

	order.ExceptionType = exceptionType
	order.ExceptionStatus = routingentity.RoutedOrderExceptionStatusOpen
	entry := fmt.Sprintf("Exception opened: %s", strings.ReplaceAll(exceptionType, "_", " "))
	now := time.Now().UTC()
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		routingentity.RoutedOrderActivityTypeSystem,
		activityActorFromContext(ctx),
		entry,
		now,
		activityDetails(
			"exception_type",
			exceptionType,
			"exception_status",
			routingentity.RoutedOrderExceptionStatusOpen,
		),
	))
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderExceptionStatus(
	ctx context.Context,
	cmd routinginputport.UpdateOrderExceptionStatusCmd,
) (*routingentity.RoutedOrder, error) {
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
		routingentity.RoutedOrderActivityTypeSystem,
		activityActorFromContext(ctx),
		entry,
		now,
		activityDetails("exception_type", order.ExceptionType, "exception_status", status),
	))
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}
