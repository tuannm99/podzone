package domain

import "context"

type SupplierRepository interface {
	Create(ctx context.Context, supplier Supplier) (*Supplier, error)
	GetByID(ctx context.Context, id string) (*Supplier, error)
	List(ctx context.Context, query ListSuppliersQuery) ([]Supplier, error)
	Update(ctx context.Context, supplier Supplier) (*Supplier, error)
	UpdateStatus(ctx context.Context, id, status string) (*Supplier, error)
}
