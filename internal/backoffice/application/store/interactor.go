package store

import (
	"context"
	"errors"
	"strings"

	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/ddd"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type Interactor struct {
	repo  storectx.StoreRepository
	ids   ddd.IDGenerator
	clock ddd.Clock
}

var _ storectx.StoreUsecase = (*Interactor)(nil)

func NewInteractor(
	repo storectx.StoreRepository,
	ids ddd.IDGenerator,
	clock ddd.Clock,
) storectx.StoreUsecase {
	return &Interactor{repo: repo, ids: ids, clock: clock}
}

func (i *Interactor) ListStores(
	ctx context.Context,
	query storectx.ListStoresQuery,
) (collection.Page[storectx.Store], error) {
	return i.repo.FindPage(ctx, query.Collection.Normalize())
}

func (i *Interactor) GetStore(ctx context.Context, query storectx.GetStoreQuery) (*storectx.Store, error) {
	return i.repo.FindByID(ctx, strings.TrimSpace(query.ID))
}

func (i *Interactor) CreateStoreFromCommand(
	ctx context.Context,
	cmd storectx.CreateStoreCmd,
) (*storectx.Store, error) {
	if strings.TrimSpace(cmd.Name) == "" {
		return nil, ddd.NewDomainError("STORE_NAME_REQUIRED", "store name is required")
	}
	ownerID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	storeID, err := i.ids.NewID("store")
	if err != nil {
		return nil, err
	}
	aggregate, _, err := storectx.CreateStore(
		storeID.String(),
		cmd.Name,
		cmd.Description,
		ownerID,
		i.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	store := aggregate.Snapshot()
	if err := i.repo.Create(ctx, store); err != nil {
		return nil, err
	}
	return &store, nil
}

func (i *Interactor) UpdateStoreStatusFromCommand(
	ctx context.Context,
	cmd storectx.UpdateStoreStatusCmd,
) (*storectx.Store, error) {
	store, err := i.repo.FindByID(ctx, strings.TrimSpace(cmd.ID))
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.Join(ddd.ErrNotFound, ddd.NewDomainError("STORE_NOT_FOUND", "store not found"))
	}
	aggregate, err := storectx.RehydrateStore(*store)
	if err != nil {
		return nil, err
	}
	if cmd.Active {
		aggregate.Activate(i.clock.Now())
	} else {
		aggregate.Deactivate(i.clock.Now())
	}
	updated := aggregate.Snapshot()
	if err := i.repo.UpdateStatus(ctx, updated.ID, updated.Status); err != nil {
		return nil, err
	}
	return i.repo.FindByID(ctx, updated.ID)
}

func (i *Interactor) BootstrapStore(
	ctx context.Context,
	cmd storectx.BootstrapStoreCmd,
) (*storectx.Store, error) {
	aggregate, _, err := storectx.CreateStore(
		strings.TrimSpace(cmd.ID),
		cmd.Name,
		"",
		strings.TrimSpace(cmd.OwnerID),
		i.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	aggregate.Activate(i.clock.Now())
	store := aggregate.Snapshot()
	if err := i.repo.Bootstrap(ctx, store); err != nil {
		return nil, err
	}
	return i.repo.FindByID(ctx, store.ID)
}

func (i *Interactor) GetStoreByID(ctx context.Context, id string) (*storectx.Store, error) {
	return i.GetStore(ctx, storectx.GetStoreQuery{ID: id})
}

func (i *Interactor) CreateStore(ctx context.Context, name, description string) (*storectx.Store, error) {
	return i.CreateStoreFromCommand(ctx, storectx.CreateStoreCmd{Name: name, Description: description})
}

func (i *Interactor) UpdateStoreStatus(ctx context.Context, id string, active bool) (*storectx.Store, error) {
	return i.UpdateStoreStatusFromCommand(ctx, storectx.UpdateStoreStatusCmd{ID: id, Active: active})
}
