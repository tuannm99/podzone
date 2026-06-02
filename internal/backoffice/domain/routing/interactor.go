package routing

import (
	"context"

	activityinputport "github.com/tuannm99/podzone/internal/backoffice/domain/activity/inputport"
	activityinteractor "github.com/tuannm99/podzone/internal/backoffice/domain/activity/interactor"
	catalogoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/outputport"
	exceptioninputport "github.com/tuannm99/podzone/internal/backoffice/domain/exception/inputport"
	exceptioninteractor "github.com/tuannm99/podzone/internal/backoffice/domain/exception/interactor"
	fulfillmentinputport "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/inputport"
	fulfillmentinteractor "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/interactor"
	orderinputport "github.com/tuannm99/podzone/internal/backoffice/domain/order/inputport"
	orderinteractor "github.com/tuannm99/podzone/internal/backoffice/domain/order/interactor"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
	routinginteractor "github.com/tuannm99/podzone/internal/backoffice/domain/routing/interactor"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	settlementinputport "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/inputport"
	settlementinteractor "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/interactor"
)

type OrderRoutingInteractor struct {
	orderCommands      orderinputport.CustomerOrderCommandUsecase
	orderQueries       orderinputport.CustomerOrderQueryUsecase
	activityQueries    activityinputport.ActivityFeedQueryUsecase
	routingCommands    routinginputport.RoutingCommandUsecase
	routingQueries     routinginputport.RoutingQueryUsecase
	exceptionCommands  exceptioninputport.ExceptionCommandUsecase
	fulfillmentCommand fulfillmentinputport.FulfillmentCommandUsecase
	settlementCommands settlementinputport.SettlementCommandUsecase
}

var _ routinginputport.OrderRoutingUsecase = (*OrderRoutingInteractor)(nil)

func NewOrderRoutingInteractor(
	orders routingoutputport.OrderRoutingRepository,
	products catalogoutputport.ProductSetupRepository,
	partners routingoutputport.PartnerDirectory,
) routinginputport.OrderRoutingUsecase {
	orderUsecase := orderinteractor.New(orders, products, partners)
	routingUsecase := routinginteractor.New(orders, products, partners)

	return &OrderRoutingInteractor{
		orderCommands:      orderUsecase,
		orderQueries:       orderUsecase,
		activityQueries:    activityinteractor.New(orders),
		routingCommands:    routingUsecase,
		routingQueries:     routingUsecase,
		exceptionCommands:  exceptioninteractor.New(orders),
		fulfillmentCommand: fulfillmentinteractor.New(orders),
		settlementCommands: settlementinteractor.New(orders),
	}
}

func (i *OrderRoutingInteractor) ListCustomerOrders(
	ctx context.Context,
	query orderinputport.ListCustomerOrdersQuery,
) ([]routingentity.RoutedOrder, error) {
	return i.orderQueries.ListCustomerOrders(ctx, query)
}

func (i *OrderRoutingInteractor) ListRoutedOrders(
	ctx context.Context,
	query routinginputport.ListRoutedOrdersQuery,
) ([]routingentity.RoutedOrder, error) {
	return i.orderQueries.ListCustomerOrders(ctx, query)
}

func (i *OrderRoutingInteractor) ListRoutedOrderActivities(
	ctx context.Context,
	query routingentity.RoutedOrderActivityFeedQuery,
) (*routingentity.RoutedOrderActivityFeedPage, error) {
	return i.activityQueries.ListRoutedOrderActivities(ctx, query)
}

func (i *OrderRoutingInteractor) RecommendRoutedOrderPartner(
	ctx context.Context,
	query routinginputport.RecommendRoutedOrderPartnerQuery,
) (*routingentity.RoutedOrderRecommendation, error) {
	return i.routingQueries.RecommendRoutedOrderPartner(ctx, query)
}

func (i *OrderRoutingInteractor) CreateCustomerOrder(
	ctx context.Context,
	cmd orderinputport.CreateCustomerOrderCmd,
) (*routingentity.RoutedOrder, error) {
	return i.orderCommands.CreateCustomerOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) CreateRoutedOrder(
	ctx context.Context,
	cmd routinginputport.CreateRoutedOrderCmd,
) (*routingentity.RoutedOrder, error) {
	return i.orderCommands.CreateCustomerOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) ForceRerouteBlockedOrder(
	ctx context.Context,
	cmd routinginputport.ForceRerouteBlockedOrderCmd,
) (*routingentity.RoutedOrder, error) {
	return i.routingCommands.ForceRerouteBlockedOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) AdvanceCustomerOrder(
	ctx context.Context,
	cmd orderinputport.AdvanceCustomerOrderCmd,
) (*routingentity.RoutedOrder, error) {
	return i.orderCommands.AdvanceCustomerOrder(ctx, cmd)
}

func (i *OrderRoutingInteractor) AdvanceRoutedOrder(
	ctx context.Context,
	storeID string,
	orderID string,
) (*routingentity.RoutedOrder, error) {
	return i.orderCommands.AdvanceCustomerOrder(ctx, orderinputport.AdvanceCustomerOrderCmd{
		StoreID: storeID,
		OrderID: orderID,
	})
}

func (i *OrderRoutingInteractor) OpenOrderException(
	ctx context.Context,
	cmd routinginputport.OpenOrderExceptionCmd,
) (*routingentity.RoutedOrder, error) {
	return i.exceptionCommands.OpenOrderException(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderExceptionStatus(
	ctx context.Context,
	cmd routinginputport.UpdateOrderExceptionStatusCmd,
) (*routingentity.RoutedOrder, error) {
	return i.exceptionCommands.UpdateOrderExceptionStatus(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderShipment(
	ctx context.Context,
	cmd routinginputport.UpdateOrderShipmentCmd,
) (*routingentity.RoutedOrder, error) {
	return i.fulfillmentCommand.UpdateOrderShipment(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderSettlement(
	ctx context.Context,
	cmd routinginputport.UpdateOrderSettlementCmd,
) (*routingentity.RoutedOrder, error) {
	return i.settlementCommands.UpdateOrderSettlement(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderIssueHandling(
	ctx context.Context,
	cmd routinginputport.UpdateOrderIssueHandlingCmd,
) (*routingentity.RoutedOrder, error) {
	return i.settlementCommands.UpdateOrderIssueHandling(ctx, cmd)
}

func (i *OrderRoutingInteractor) UpdateOrderQueueControl(
	ctx context.Context,
	cmd routinginputport.UpdateOrderQueueControlCmd,
) (*routingentity.RoutedOrder, error) {
	return i.orderCommands.UpdateOrderQueueControl(ctx, cmd)
}

func (i *OrderRoutingInteractor) BulkUpdateOrders(
	ctx context.Context,
	cmd orderinputport.BulkUpdateOrdersCmd,
) ([]routingentity.RoutedOrder, error) {
	return i.orderCommands.BulkUpdateOrders(ctx, cmd)
}

func (i *OrderRoutingInteractor) BulkUpdateRoutedOrders(
	ctx context.Context,
	cmd routinginputport.BulkUpdateRoutedOrdersCmd,
) ([]routingentity.RoutedOrder, error) {
	return i.orderCommands.BulkUpdateOrders(ctx, cmd)
}
