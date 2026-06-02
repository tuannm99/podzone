package interactor

import (
	"context"
	"strings"
	"time"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	routingsupport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/support"
	settlementinputport "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/inputport"
)

type Interactor struct {
	orders routingoutputport.OrderRoutingRepository
}

var _ settlementinputport.SettlementCommandUsecase = (*Interactor)(nil)

func New(orders routingoutputport.OrderRoutingRepository) *Interactor {
	return &Interactor{orders: orders}
}

func (i *Interactor) UpdateOrderSettlement(
	ctx context.Context,
	cmd settlementinputport.UpdateOrderSettlementCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingsupport.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := order.UpdateSettlement(
		cmd.FulfillmentCost,
		cmd.ShippingCost,
		cmd.SettlementStatus,
		cmd.Notes,
		routingsupport.ActivityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *Interactor) UpdateOrderIssueHandling(
	ctx context.Context,
	cmd settlementinputport.UpdateOrderIssueHandlingCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingsupport.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := order.UpdateIssueHandling(
		cmd.IssueCost,
		cmd.IssueResolution,
		cmd.Notes,
		routingsupport.ActivityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
