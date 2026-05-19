package inputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
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

type ListRoutedOrderActivitiesQuery struct {
	ActivityType  string
	ActorContains string
	OrderID       string
	Partner       string
	Assignee      string
	Since         *time.Time
	Limit         int
	After         string
	IncludeSystem bool
}

type OrderRoutingUsecase interface {
	ListRoutedOrders(ctx context.Context) ([]entity.RoutedOrder, error)
	ListRoutedOrderActivities(
		ctx context.Context,
		query ListRoutedOrderActivitiesQuery,
	) (*entity.RoutedOrderActivityFeedPage, error)
	RecommendRoutedOrderPartner(
		ctx context.Context,
		query RecommendRoutedOrderPartnerQuery,
	) (*entity.RoutedOrderRecommendation, error)
	CreateRoutedOrder(ctx context.Context, cmd CreateRoutedOrderCmd) (*entity.RoutedOrder, error)
	AdvanceRoutedOrder(ctx context.Context, orderID string) (*entity.RoutedOrder, error)
	OpenOrderException(ctx context.Context, cmd OpenOrderExceptionCmd) (*entity.RoutedOrder, error)
	UpdateOrderExceptionStatus(ctx context.Context, cmd UpdateOrderExceptionStatusCmd) (*entity.RoutedOrder, error)
	UpdateOrderShipment(ctx context.Context, cmd UpdateOrderShipmentCmd) (*entity.RoutedOrder, error)
	UpdateOrderSettlement(ctx context.Context, cmd UpdateOrderSettlementCmd) (*entity.RoutedOrder, error)
	UpdateOrderIssueHandling(ctx context.Context, cmd UpdateOrderIssueHandlingCmd) (*entity.RoutedOrder, error)
	UpdateOrderQueueControl(ctx context.Context, cmd UpdateOrderQueueControlCmd) (*entity.RoutedOrder, error)
	BulkUpdateRoutedOrders(ctx context.Context, cmd BulkUpdateRoutedOrdersCmd) ([]entity.RoutedOrder, error)
}
