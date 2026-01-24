package interactor

import (
	"context"
	"errors"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	"github.com/tuannm99/podzone/internal/backoffice/infrastructure/model"
)

type StoreInteractor struct {
	repo outputport.StoreRepository
}

func NewStoreInteractor(repo outputport.StoreRepository) inputport.StoreUsecase {
	return &StoreInteractor{repo: repo}
}

func (i *StoreInteractor) GetAllStores(ctx context.Context) ([]entity.Store, error) {
	return i.repo.FindAll(ctx)
}

func (i *StoreInteractor) GetStoreByID(ctx context.Context, id string) (*entity.Store, error) {
	return i.repo.FindByID(ctx, id)
}

func (i *StoreInteractor) CreateStore(ctx context.Context, name, description string) (*entity.Store, error) {
	if name == "" {
		return nil, errors.New("store name is required")
	}
	s := model.NewStore(name, description, "owner_1") // later: extract from ctx
	if err := i.repo.Create(ctx, s); err != nil {
		return nil, err
	}
	return &entity.Store{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		IsActive:    false,
	}, nil
}

func (i *StoreInteractor) UpdateStoreStatus(ctx context.Context, id string, active bool) (*entity.Store, error) {
	status := model.StoreStatusInactive
	if active {
		status = model.StoreStatusActive
	}
	if err := i.repo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	return i.repo.FindByID(ctx, id)
}
