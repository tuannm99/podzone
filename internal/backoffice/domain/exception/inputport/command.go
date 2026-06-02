package inputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type OpenOrderExceptionCmd struct {
	StoreID       string
	OrderID       string
	ExceptionType string
}

type UpdateOrderExceptionStatusCmd struct {
	StoreID string
	OrderID string
	Status  string
}

type ExceptionCommandUsecase interface {
	OpenOrderException(ctx context.Context, cmd OpenOrderExceptionCmd) (*routingentity.RoutedOrder, error)
	UpdateOrderExceptionStatus(
		ctx context.Context,
		cmd UpdateOrderExceptionStatusCmd,
	) (*routingentity.RoutedOrder, error)
}
