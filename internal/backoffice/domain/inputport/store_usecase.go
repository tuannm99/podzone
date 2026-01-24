package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
)

type StoreUsecase interface {
	GetAllStores(ctx context.Context) ([]entity.Store, error)
	GetStoreByID(ctx context.Context, id string) (*entity.Store, error)
	CreateStore(ctx context.Context, name, description string) (*entity.Store, error)
	UpdateStoreStatus(ctx context.Context, id string, active bool) (*entity.Store, error)
}
