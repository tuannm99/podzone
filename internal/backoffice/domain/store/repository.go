package store

import "context"

type StoreRepository interface {
	FindAll(ctx context.Context) ([]Store, error)
	FindByID(ctx context.Context, id string) (*Store, error)
	Create(ctx context.Context, store Store) error
	Bootstrap(ctx context.Context, store Store) error
	UpdateStatus(ctx context.Context, id string, status string) error
}
