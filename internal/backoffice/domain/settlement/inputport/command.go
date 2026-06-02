package inputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type UpdateOrderSettlementCmd struct {
	StoreID          string
	OrderID          string
	FulfillmentCost  string
	ShippingCost     string
	SettlementStatus string
	Notes            string
}

type UpdateOrderIssueHandlingCmd struct {
	StoreID         string
	OrderID         string
	IssueCost       string
	IssueResolution string
	Notes           string
}

type SettlementCommandUsecase interface {
	UpdateOrderSettlement(ctx context.Context, cmd UpdateOrderSettlementCmd) (*routingentity.RoutedOrder, error)
	UpdateOrderIssueHandling(ctx context.Context, cmd UpdateOrderIssueHandlingCmd) (*routingentity.RoutedOrder, error)
}
