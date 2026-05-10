package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
)

type OrderRoutingRepository interface {
	List(ctx context.Context) ([]entity.RoutedOrder, error)
	GetByID(ctx context.Context, id string) (*entity.RoutedOrder, error)
	Create(ctx context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error)
	Update(ctx context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error)
}
