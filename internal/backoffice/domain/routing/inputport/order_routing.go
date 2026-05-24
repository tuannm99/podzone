package inputport

import (
	"context"
	"time"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type CreateRoutedOrderCmd struct {
	CandidateID      string
	CustomerName     string
	Quantity         int
	ProductType      string
	ShipRegion       string
	PreferredPartner string
}

type RecommendRoutedOrderPartnerQuery struct {
	CandidateID      string
	ProductType      string
	ShipRegion       string
	PreferredPartner string
}

type OpenOrderExceptionCmd struct {
	OrderID       string
	ExceptionType string
}

type ForceRerouteBlockedOrderCmd struct {
	OrderID          string
	PreferredPartner string
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

type UpdateOrderQueueControlCmd struct {
	OrderID          string
	OperatorAssignee string
	ShipmentSlaDueAt *time.Time
	IssueSlaDueAt    *time.Time
}

type BulkUpdateRoutedOrdersCmd struct {
	OrderIDs         []string
	OperatorAssignee *string
	ShipmentSlaDueAt *time.Time
	SettlementStatus *string
}

type OrderRoutingUsecase interface {
	ListRoutedOrders(ctx context.Context) ([]routingentity.RoutedOrder, error)
	ListRoutedOrderActivities(
		ctx context.Context,
		query routingentity.RoutedOrderActivityFeedQuery,
	) (*routingentity.RoutedOrderActivityFeedPage, error)
	RecommendRoutedOrderPartner(
		ctx context.Context,
		query RecommendRoutedOrderPartnerQuery,
	) (*routingentity.RoutedOrderRecommendation, error)
	CreateRoutedOrder(ctx context.Context, cmd CreateRoutedOrderCmd) (*routingentity.RoutedOrder, error)
	ForceRerouteBlockedOrder(ctx context.Context, cmd ForceRerouteBlockedOrderCmd) (*routingentity.RoutedOrder, error)
	AdvanceRoutedOrder(ctx context.Context, orderID string) (*routingentity.RoutedOrder, error)
	OpenOrderException(ctx context.Context, cmd OpenOrderExceptionCmd) (*routingentity.RoutedOrder, error)
	UpdateOrderExceptionStatus(
		ctx context.Context,
		cmd UpdateOrderExceptionStatusCmd,
	) (*routingentity.RoutedOrder, error)
	UpdateOrderShipment(ctx context.Context, cmd UpdateOrderShipmentCmd) (*routingentity.RoutedOrder, error)
	UpdateOrderSettlement(ctx context.Context, cmd UpdateOrderSettlementCmd) (*routingentity.RoutedOrder, error)
	UpdateOrderIssueHandling(ctx context.Context, cmd UpdateOrderIssueHandlingCmd) (*routingentity.RoutedOrder, error)
	UpdateOrderQueueControl(ctx context.Context, cmd UpdateOrderQueueControlCmd) (*routingentity.RoutedOrder, error)
	BulkUpdateRoutedOrders(ctx context.Context, cmd BulkUpdateRoutedOrdersCmd) ([]routingentity.RoutedOrder, error)
}
