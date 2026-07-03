package store

import (
	"context"

	"github.com/tuannm99/podzone/pkg/collection"
)

type StoreRepository interface {
	FindPage(ctx context.Context, query collection.Query) (collection.Page[Store], error)
	FindByID(ctx context.Context, id string) (*Store, error)
	Create(ctx context.Context, store Store) error
	Bootstrap(ctx context.Context, store Store) error
	UpdateStatus(ctx context.Context, id string, status string) error
}
