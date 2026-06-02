package inputport

import (
	"context"
	"time"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type CreateCustomerOrderCmd struct {
	StoreID          string
	CandidateID      string
	CustomerName     string
	Quantity         int
	ProductType      string
	ShipRegion       string
	PreferredPartner string
}

type AdvanceCustomerOrderCmd struct {
	StoreID string
	OrderID string
}

type UpdateOrderQueueControlCmd struct {
	StoreID          string
	OrderID          string
	OperatorAssignee string
	ShipmentSlaDueAt *time.Time
	IssueSlaDueAt    *time.Time
}

type BulkUpdateOrdersCmd struct {
	StoreID          string
	OrderIDs         []string
	OperatorAssignee *string
	ShipmentSlaDueAt *time.Time
	SettlementStatus *string
}

type CustomerOrderCommandUsecase interface {
	CreateCustomerOrder(ctx context.Context, cmd CreateCustomerOrderCmd) (*routingentity.RoutedOrder, error)
	AdvanceCustomerOrder(ctx context.Context, cmd AdvanceCustomerOrderCmd) (*routingentity.RoutedOrder, error)
	UpdateOrderQueueControl(ctx context.Context, cmd UpdateOrderQueueControlCmd) (*routingentity.RoutedOrder, error)
	BulkUpdateOrders(ctx context.Context, cmd BulkUpdateOrdersCmd) ([]routingentity.RoutedOrder, error)
}
