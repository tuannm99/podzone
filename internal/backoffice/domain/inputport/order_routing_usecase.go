package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
)

type CreateRoutedOrderCmd struct {
	CandidateID  string
	CustomerName string
	Quantity     int
}

type OpenOrderExceptionCmd struct {
	OrderID       string
	ExceptionType string
}

type UpdateOrderExceptionStatusCmd struct {
	OrderID string
	Status  string
}

type UpdateOrderShipmentCmd struct {
	OrderID        string
	ShipmentStatus string
	Carrier        string
	TrackingNumber string
	TrackingURL    string
	Notes          string
}

type UpdateOrderSettlementCmd struct {
	OrderID          string
	FulfillmentCost  string
	ShippingCost     string
	SettlementStatus string
	Notes            string
}

type UpdateOrderIssueHandlingCmd struct {
	OrderID         string
	IssueCost       string
	IssueResolution string
	Notes           string
}

type OrderRoutingUsecase interface {
	ListRoutedOrders(ctx context.Context) ([]entity.RoutedOrder, error)
	CreateRoutedOrder(ctx context.Context, cmd CreateRoutedOrderCmd) (*entity.RoutedOrder, error)
	AdvanceRoutedOrder(ctx context.Context, orderID string) (*entity.RoutedOrder, error)
	OpenOrderException(ctx context.Context, cmd OpenOrderExceptionCmd) (*entity.RoutedOrder, error)
	UpdateOrderExceptionStatus(ctx context.Context, cmd UpdateOrderExceptionStatusCmd) (*entity.RoutedOrder, error)
	UpdateOrderShipment(ctx context.Context, cmd UpdateOrderShipmentCmd) (*entity.RoutedOrder, error)
	UpdateOrderSettlement(ctx context.Context, cmd UpdateOrderSettlementCmd) (*entity.RoutedOrder, error)
	UpdateOrderIssueHandling(ctx context.Context, cmd UpdateOrderIssueHandlingCmd) (*entity.RoutedOrder, error)
}
