package operations

import (
	"context"

	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	"github.com/tuannm99/podzone/pkg/ddd"
)

type OrderRoutingInteractor struct {
	orderOperations    *OrderInteractor
	activityQueries    *ActivityInteractor
	routingCommands    routingctx.RoutingCommandUsecase
	routingQueries     routingctx.RoutingQueryUsecase
	exceptionCommands  *ExceptionInteractor
	fulfillmentCommand *FulfillmentInteractor
	settlementCommands *SettlementInteractor
}

var _ OrderRoutingUsecase = (*OrderRoutingInteractor)(nil)

func NewOrderRoutingInteractor(
	orders routingctx.OrderRoutingRepository,
	customerOrders orderctx.CustomerOrderQueryRepository,
	products catalogctx.ProductSetupRepository,
	partners routingctx.PartnerDirectory,
	dispatcher ddd.EventDispatcher,
	ids ddd.IDGenerator,
	clock ddd.Clock,
) OrderRoutingUsecase {
	orderUsecase := NewOrderInteractor(orders, customerOrders, products, partners, dispatcher, ids, clock)
	routingUsecase := NewRoutingInteractor(orders, customerOrders, products, partners, dispatcher, clock)

	return &OrderRoutingInteractor{
		orderOperations:    orderUsecase,
		activityQueries:    NewActivityInteractor(orders),
		routingCommands:    routingUsecase,
		routingQueries:     routingUsecase,
		exceptionCommands:  NewExceptionInteractor(orders, dispatcher, clock),
		fulfillmentCommand: NewFulfillmentInteractor(orders, customerOrders, dispatcher, clock),
		settlementCommands: NewSettlementInteractor(orders, dispatcher, clock),
	}
}

func (i *OrderRoutingInteractor) ListCustomerOrders(
	ctx context.Context,
	query orderctx.ListCustomerOrdersQuery,
) ([]routingctx.RoutedOrder, error) {
	return i.orderOperations.ListCustomerOrders(ctx, query)
}

func (i *OrderRoutingInteractor) ListRoutedOrders(
	ctx context.Context,
	query ListRoutedOrdersQuery,
) ([]routingctx.RoutedOrder, error) {
	return i.orderOperations.ListCustomerOrders(ctx, query)
}

func (i *OrderRoutingInteractor) ListRoutedOrderActivities(
	ctx context.Context,
	query routingctx.RoutedOrderActivityFeedQuery,
) (*routingctx.RoutedOrderActivityFeedPage, error) {
	return i.activityQueries.ListRoutedOrderActivities(ctx, query)
}

func (i *OrderRoutingInteractor) RecommendRoutedOrderPartner(
	ctx context.Context,
	query routingctx.RecommendRoutedOrderPartnerQuery,
) (*routingctx.RoutedOrderRecommendation, error) {
	return i.routingQueries.RecommendRoutedOrderPartner(ctx, query)
}

func (i *OrderRoutingInteractor) CreateCustomerOrder(
	ctx context.Context,
	cmd orderctx.CreateCustomerOrderCmd,
) (*routingctx.RoutedOrder, error) {
	return i.orderOperations.CreateCustomerOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) CreateRoutedOrder(
	ctx context.Context,
	cmd CreateRoutedOrderCmd,
) (*routingctx.RoutedOrder, error) {
	return i.orderOperations.CreateCustomerOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) ForceRerouteBlockedOrder(
	ctx context.Context,
	cmd routingctx.ForceRerouteBlockedOrderCmd,
) (*routingctx.RoutedOrder, error) {
	return i.routingCommands.ForceRerouteBlockedOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) AdvanceCustomerOrder(
	ctx context.Context,
	cmd orderctx.AdvanceCustomerOrderCmd,
) (*routingctx.RoutedOrder, error) {
	return i.orderOperations.AdvanceCustomerOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) AdvanceRoutedOrder(
	ctx context.Context,
	storeID string,
	orderID string,
) (*routingctx.RoutedOrder, error) {
	return i.orderOperations.AdvanceCustomerOrder(ctx, orderctx.AdvanceCustomerOrderCmd{
		StoreID: storeID,
		OrderID: orderID,
	})
}

func (i *OrderRoutingInteractor) OpenOrderException(
	ctx context.Context,
	cmd OpenOrderExceptionCmd,
) (*routingctx.RoutedOrder, error) {
	return i.exceptionCommands.OpenOrderException(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderExceptionStatus(
	ctx context.Context,
	cmd UpdateOrderExceptionStatusCmd,
) (*routingctx.RoutedOrder, error) {
	return i.exceptionCommands.UpdateOrderExceptionStatus(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderShipment(
	ctx context.Context,
	cmd UpdateOrderShipmentCmd,
) (*routingctx.RoutedOrder, error) {
	return i.fulfillmentCommand.UpdateOrderShipment(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderSettlement(
	ctx context.Context,
	cmd UpdateOrderSettlementCmd,
) (*routingctx.RoutedOrder, error) {
	return i.settlementCommands.UpdateOrderSettlement(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderIssueHandling(
	ctx context.Context,
	cmd UpdateOrderIssueHandlingCmd,
) (*routingctx.RoutedOrder, error) {
	return i.settlementCommands.UpdateOrderIssueHandling(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderQueueControl(
	ctx context.Context,
	cmd UpdateOrderQueueControlCmd,
) (*routingctx.RoutedOrder, error) {
	return i.orderOperations.UpdateOrderQueueControl(ctx, cmd)
}

func (i *OrderRoutingInteractor) BulkUpdateOrders(
	ctx context.Context,
	cmd orderctx.BulkUpdateOrdersCmd,
) ([]routingctx.RoutedOrder, error) {
	return i.orderOperations.BulkUpdateOrders(ctx, cmd)
}

func (i *OrderRoutingInteractor) BulkUpdateRoutedOrders(
	ctx context.Context,
	cmd BulkUpdateRoutedOrdersCmd,
) ([]routingctx.RoutedOrder, error) {
	return i.orderOperations.BulkUpdateOrders(ctx, cmd)
}
