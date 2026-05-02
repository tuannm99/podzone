package domain

import "context"

type SupplierUsecase interface {
	CreateSupplier(ctx context.Context, cmd CreateSupplierCmd) (*Supplier, error)
	GetSupplier(ctx context.Context, id string) (*Supplier, error)
	ListSuppliers(ctx context.Context, query ListSuppliersQuery) ([]Supplier, error)
	UpdateSupplier(ctx context.Context, cmd UpdateSupplierCmd) (*Supplier, error)
	UpdateSupplierStatus(ctx context.Context, id, status string) (*Supplier, error)
}
