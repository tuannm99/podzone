package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type supplierRepoFake struct {
	createFunc       func(ctx context.Context, supplier Supplier) (*Supplier, error)
	getByIDFunc      func(ctx context.Context, id string) (*Supplier, error)
	listFunc         func(ctx context.Context, query ListSuppliersQuery) ([]Supplier, error)
	updateFunc       func(ctx context.Context, supplier Supplier) (*Supplier, error)
	updateStatusFunc func(ctx context.Context, id, status string) (*Supplier, error)
}

func (f *supplierRepoFake) Create(ctx context.Context, supplier Supplier) (*Supplier, error) {
	return f.createFunc(ctx, supplier)
}

func (f *supplierRepoFake) GetByID(ctx context.Context, id string) (*Supplier, error) {
	return f.getByIDFunc(ctx, id)
}

func (f *supplierRepoFake) List(ctx context.Context, query ListSuppliersQuery) ([]Supplier, error) {
	return f.listFunc(ctx, query)
}

func (f *supplierRepoFake) Update(ctx context.Context, supplier Supplier) (*Supplier, error) {
	return f.updateFunc(ctx, supplier)
}

func (f *supplierRepoFake) UpdateStatus(ctx context.Context, id, status string) (*Supplier, error) {
	return f.updateStatusFunc(ctx, id, status)
}

func TestCreateSupplier_NormalizesCodeFromName(t *testing.T) {
	t.Parallel()

	uc := NewSupplierUsecase(&supplierRepoFake{
		createFunc: func(ctx context.Context, supplier Supplier) (*Supplier, error) {
			return &supplier, nil
		},
	})

	out, err := uc.CreateSupplier(context.Background(), CreateSupplierCmd{
		TenantID:    "tenant-1",
		Name:        "Acme Supply Co",
		PartnerType: PartnerTypePrintOnDemand,
	})
	require.NoError(t, err)
	require.Equal(t, "acme-supply-co", out.Code)
	require.Equal(t, PartnerTypePrintOnDemand, out.PartnerType)
	require.Equal(t, SupplierStatusActive, out.Status)
}

func TestUpdateSupplierStatus_RejectsUnknownStatus(t *testing.T) {
	t.Parallel()

	uc := NewSupplierUsecase(&supplierRepoFake{})
	out, err := uc.UpdateSupplierStatus(context.Background(), "sup-1", "paused")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrInvalidSupplierStatus)
}

func TestCreateSupplier_RejectsUnknownPartnerType(t *testing.T) {
	t.Parallel()

	uc := NewSupplierUsecase(&supplierRepoFake{})
	out, err := uc.CreateSupplier(context.Background(), CreateSupplierCmd{
		TenantID:    "tenant-1",
		Name:        "Acme",
		PartnerType: "supplier",
	})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrInvalidPartnerType)
}
