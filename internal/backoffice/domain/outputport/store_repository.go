package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/infrastructure/model"
)

type StoreRepository interface {
	FindAll() ([]entity.Store, error)
	FindByID(id string) (*entity.Store, error)
	Create(ctx context.Context, s *model.Store) error
	UpdateStatus(ctx context.Context, id string, status model.StoreStatus) error
}
