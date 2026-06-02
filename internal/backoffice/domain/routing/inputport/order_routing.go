package inputport

import (
	"context"

	activityinputport "github.com/tuannm99/podzone/internal/backoffice/domain/activity/inputport"
	exceptioninputport "github.com/tuannm99/podzone/internal/backoffice/domain/exception/inputport"
	fulfillmentinputport "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/inputport"
	orderinputport "github.com/tuannm99/podzone/internal/backoffice/domain/order/inputport"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	settlementinputport "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/inputport"
)

type (
	CreateRoutedOrderCmd          = orderinputport.CreateCustomerOrderCmd
	OpenOrderExceptionCmd         = exceptioninputport.OpenOrderExceptionCmd
	UpdateOrderExceptionStatusCmd = exceptioninputport.UpdateOrderExceptionStatusCmd
	UpdateOrderShipmentCmd        = fulfillmentinputport.UpdateOrderShipmentCmd
	UpdateOrderSettlementCmd      = settlementinputport.UpdateOrderSettlementCmd
	UpdateOrderIssueHandlingCmd   = settlementinputport.UpdateOrderIssueHandlingCmd
	UpdateOrderQueueControlCmd    = orderinputport.UpdateOrderQueueControlCmd
	BulkUpdateRoutedOrdersCmd     = orderinputport.BulkUpdateOrdersCmd
	ListRoutedOrdersQuery         = orderinputport.ListCustomerOrdersQuery
)

type OrderRoutingUsecase interface {
	orderinputport.CustomerOrderCommandUsecase
	orderinputport.CustomerOrderQueryUsecase
	activityinputport.ActivityFeedQueryUsecase
	RoutingCommandUsecase
	RoutingQueryUsecase
	exceptioninputport.ExceptionCommandUsecase
	fulfillmentinputport.FulfillmentCommandUsecase
	settlementinputport.SettlementCommandUsecase

	// Compatibility facade for the current GraphQL surface while bounded contexts are split.
	ListRoutedOrders(ctx context.Context, query ListRoutedOrdersQuery) ([]routingentity.RoutedOrder, error)
	CreateRoutedOrder(ctx context.Context, cmd CreateRoutedOrderCmd) (*routingentity.RoutedOrder, error)
	AdvanceRoutedOrder(ctx context.Context, storeID, orderID string) (*routingentity.RoutedOrder, error)
	BulkUpdateRoutedOrders(ctx context.Context, cmd BulkUpdateRoutedOrdersCmd) ([]routingentity.RoutedOrder, error)
}
