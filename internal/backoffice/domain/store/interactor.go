package store

import (
	"context"
	"errors"

	storeentity "github.com/tuannm99/podzone/internal/backoffice/domain/store/entity"
	storeinputport "github.com/tuannm99/podzone/internal/backoffice/domain/store/inputport"
	storeoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/store/outputport"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type StoreInteractor struct {
	repo storeoutputport.StoreRepository
}

var _ storeinputport.StoreUsecase = (*StoreInteractor)(nil)

func NewStoreInteractor(repo storeoutputport.StoreRepository) storeinputport.StoreUsecase {
	return &StoreInteractor{repo: repo}
}

func (i *StoreInteractor) GetAllStores(ctx context.Context) ([]storeentity.Store, error) {
	return i.repo.FindAll(ctx)
}

func (i *StoreInteractor) GetStoreByID(ctx context.Context, id string) (*storeentity.Store, error) {
	return i.repo.FindByID(ctx, id)
}

func (i *StoreInteractor) CreateStore(ctx context.Context, name, description string) (*storeentity.Store, error) {
	if name == "" {
		return nil, errors.New("store name is required")
	}
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	store := storeentity.NewStore(name, description, ownerID)
	if err := i.repo.Create(ctx, store); err != nil {
		return nil, err
	}
	return &store, nil
}

func (i *StoreInteractor) UpdateStoreStatus(ctx context.Context, id string, active bool) (*storeentity.Store, error) {
	status := storeentity.StoreStatusInactive
	if active {
		status = storeentity.StoreStatusActive
	}
	if err := i.repo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	return i.repo.FindByID(ctx, id)
}
