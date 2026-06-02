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
	storeID, err := requiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := ensureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := order.UpdateShipment(
		cmd.ShipmentStatus,
		cmd.Carrier,
		cmd.TrackingNumber,
		cmd.TrackingURL,
		cmd.Notes,
		activityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderQueueControl(
	ctx context.Context,
	cmd routinginputport.UpdateOrderQueueControlCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := requiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := ensureOrderStore(order, storeID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	order.UpdateQueueControl(
		cmd.OperatorAssignee,
		cmd.ShipmentSlaDueAt,
		cmd.IssueSlaDueAt,
		activityActorFromContext(ctx),
		now,
	)
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) BulkUpdateRoutedOrders(
	ctx context.Context,
	cmd routinginputport.BulkUpdateRoutedOrdersCmd,
) ([]routingentity.RoutedOrder, error) {
	storeID, err := requiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	if len(cmd.OrderIDs) == 0 {
		return nil, fmt.Errorf("at least one routed order id is required")
	}
	if cmd.OperatorAssignee == nil && cmd.ShipmentSlaDueAt == nil && cmd.SettlementStatus == nil {
		return nil, fmt.Errorf("at least one bulk update field is required")
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
		if err := ensureOrderStore(order, storeID); err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		if err := order.ApplyBulkQueueControl(
			cmd.OperatorAssignee,
			cmd.ShipmentSlaDueAt,
			cmd.SettlementStatus,
			activityActorFromContext(ctx),
			now,
		); err != nil {
			return nil, err
		}
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
	storeID, err := requiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := ensureOrderStore(order, storeID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := order.UpdateSettlement(
		cmd.FulfillmentCost,
		cmd.ShippingCost,
		cmd.SettlementStatus,
		cmd.Notes,
		activityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderIssueHandling(
	ctx context.Context,
	cmd routinginputport.UpdateOrderIssueHandlingCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := requiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := ensureOrderStore(order, storeID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := order.UpdateIssueHandling(
		cmd.IssueCost,
		cmd.IssueResolution,
		cmd.Notes,
		activityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
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
