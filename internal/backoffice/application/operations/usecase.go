package operations

import (
	"context"

	exceptionctx "github.com/tuannm99/podzone/internal/backoffice/domain/exception"
	fulfillmentctx "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	settlementctx "github.com/tuannm99/podzone/internal/backoffice/domain/settlement"
)

type (
	CreateRoutedOrderCmd          = orderctx.CreateCustomerOrderCmd
	OpenOrderExceptionCmd         = exceptionctx.OpenOrderExceptionCmd
	UpdateOrderExceptionStatusCmd = exceptionctx.UpdateOrderExceptionStatusCmd
	UpdateOrderShipmentCmd        = fulfillmentctx.UpdateOrderShipmentCmd
	UpdateOrderSettlementCmd      = settlementctx.UpdateOrderSettlementCmd
	UpdateOrderIssueHandlingCmd   = settlementctx.UpdateOrderIssueHandlingCmd
	UpdateOrderQueueControlCmd    = orderctx.UpdateOrderQueueControlCmd
	BulkUpdateRoutedOrdersCmd     = orderctx.BulkUpdateOrdersCmd
	ListRoutedOrdersQuery         = orderctx.ListCustomerOrdersQuery
)

type OrderRoutingUsecase interface {
	routingctx.RoutingCommandUsecase
	routingctx.RoutingQueryUsecase

	ListCustomerOrders(
		ctx context.Context,
		query orderctx.ListCustomerOrdersQuery,
	) ([]routingctx.RoutedOrder, error)
	ListRoutedOrders(ctx context.Context, query ListRoutedOrdersQuery) ([]routingctx.RoutedOrder, error)
	ListRoutedOrderActivities(
		ctx context.Context,
		query routingctx.RoutedOrderActivityFeedQuery,
	) (*routingctx.RoutedOrderActivityFeedPage, error)
	CreateCustomerOrder(
		ctx context.Context,
		cmd orderctx.CreateCustomerOrderCmd,
	) (*routingctx.RoutedOrder, error)
	CreateRoutedOrder(ctx context.Context, cmd CreateRoutedOrderCmd) (*routingctx.RoutedOrder, error)
	AdvanceCustomerOrder(
		ctx context.Context,
		cmd orderctx.AdvanceCustomerOrderCmd,
	) (*routingctx.RoutedOrder, error)
	AdvanceRoutedOrder(ctx context.Context, storeID, orderID string) (*routingctx.RoutedOrder, error)
	OpenOrderException(ctx context.Context, cmd OpenOrderExceptionCmd) (*routingctx.RoutedOrder, error)
	UpdateOrderExceptionStatus(
		ctx context.Context,
		cmd UpdateOrderExceptionStatusCmd,
	) (*routingctx.RoutedOrder, error)
	UpdateOrderShipment(ctx context.Context, cmd UpdateOrderShipmentCmd) (*routingctx.RoutedOrder, error)
	UpdateOrderSettlement(ctx context.Context, cmd UpdateOrderSettlementCmd) (*routingctx.RoutedOrder, error)
	UpdateOrderIssueHandling(ctx context.Context, cmd UpdateOrderIssueHandlingCmd) (*routingctx.RoutedOrder, error)
	UpdateOrderQueueControl(ctx context.Context, cmd UpdateOrderQueueControlCmd) (*routingctx.RoutedOrder, error)
	BulkUpdateOrders(ctx context.Context, cmd orderctx.BulkUpdateOrdersCmd) ([]routingctx.RoutedOrder, error)
	BulkUpdateRoutedOrders(ctx context.Context, cmd BulkUpdateRoutedOrdersCmd) ([]routingctx.RoutedOrder, error)
}
