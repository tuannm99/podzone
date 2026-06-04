package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit"
)

type StoreInteractor struct {
	repo StoreRepository
}

var _ StoreUsecase = (*StoreInteractor)(nil)

func NewStoreInteractor(repo StoreRepository) StoreUsecase {
	return &StoreInteractor{repo: repo}
}

func (i *StoreInteractor) ListStores(ctx context.Context, _ ListStoresQuery) ([]Store, error) {
	return i.repo.FindAll(ctx)
}

func (i *StoreInteractor) GetStore(ctx context.Context, query GetStoreQuery) (*Store, error) {
	return i.repo.FindByID(ctx, strings.TrimSpace(query.ID))
}

func (i *StoreInteractor) CreateStoreFromCommand(ctx context.Context, cmd CreateStoreCmd) (*Store, error) {
	if strings.TrimSpace(cmd.Name) == "" {
		return nil, errors.New("store name is required")
	}
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	store, _, err := CreateStore(cmd.Name, cmd.Description, ownerID, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := i.repo.Create(ctx, store); err != nil {
		return nil, err
	}
	return &store, nil
}

func (i *StoreInteractor) UpdateStoreStatusFromCommand(
	ctx context.Context,
	cmd UpdateStoreStatusCmd,
) (*Store, error) {
	store, err := i.repo.FindByID(ctx, strings.TrimSpace(cmd.ID))
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("store not found")
	}
	now := time.Now().UTC()
	if cmd.Active {
		store.Activate(now)
	} else {
		store.Deactivate(now)
	}
	if err := i.repo.UpdateStatus(ctx, store.ID, store.Status); err != nil {
		return nil, err
	}
	return i.repo.FindByID(ctx, store.ID)
}

func (i *StoreInteractor) GetAllStores(ctx context.Context) ([]Store, error) {
	return i.ListStores(ctx, ListStoresQuery{})
}

func (i *StoreInteractor) GetStoreByID(ctx context.Context, id string) (*Store, error) {
	return i.GetStore(ctx, GetStoreQuery{ID: id})
}

func (i *StoreInteractor) CreateStore(ctx context.Context, name, description string) (*Store, error) {
	return i.CreateStoreFromCommand(ctx, CreateStoreCmd{Name: name, Description: description})
}

func (i *StoreInteractor) UpdateStoreStatus(ctx context.Context, id string, active bool) (*Store, error) {
	return i.UpdateStoreStatusFromCommand(ctx, UpdateStoreStatusCmd{ID: id, Active: active})
}
