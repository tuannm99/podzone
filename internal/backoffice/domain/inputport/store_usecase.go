package inputport

import "github.com/tuannm99/podzone/internal/backoffice/domain/entity"

type StoreUsecase interface {
	GetAllStores() ([]entity.Store, error)
	GetStoreByID(id string) (*entity.Store, error)
	CreateStore(name, description string) (*entity.Store, error)
	UpdateStoreStatus(id string, active bool) (*entity.Store, error)
}
