package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type partnerRepoFake struct {
	createFunc       func(ctx context.Context, partner Partner) (*Partner, error)
	getByIDFunc      func(ctx context.Context, id string) (*Partner, error)
	listFunc         func(ctx context.Context, query ListPartnersQuery) ([]Partner, error)
	updateFunc       func(ctx context.Context, partner Partner) (*Partner, error)
	updateStatusFunc func(ctx context.Context, id, status string) (*Partner, error)
}

func (f *partnerRepoFake) Create(ctx context.Context, partner Partner) (*Partner, error) {
	return f.createFunc(ctx, partner)
}

func (f *partnerRepoFake) GetByID(ctx context.Context, id string) (*Partner, error) {
	return f.getByIDFunc(ctx, id)
}

func (f *partnerRepoFake) List(ctx context.Context, query ListPartnersQuery) ([]Partner, error) {
	return f.listFunc(ctx, query)
}

func (f *partnerRepoFake) Update(ctx context.Context, partner Partner) (*Partner, error) {
	return f.updateFunc(ctx, partner)
}

func (f *partnerRepoFake) UpdateStatus(ctx context.Context, id, status string) (*Partner, error) {
	return f.updateStatusFunc(ctx, id, status)
}

func TestCreatePartner_NormalizesCodeFromName(t *testing.T) {
	t.Parallel()

	uc := NewPartnerUsecase(&partnerRepoFake{
		createFunc: func(ctx context.Context, partner Partner) (*Partner, error) {
			return &partner, nil
		},
	})

	out, err := uc.CreatePartner(context.Background(), CreatePartnerCmd{
		TenantID:    "tenant-1",
		Name:        "Acme Supply Co",
		PartnerType: PartnerTypePrintOnDemand,
	})
	require.NoError(t, err)
	require.Equal(t, "acme-supply-co", out.Code)
	require.Equal(t, PartnerTypePrintOnDemand, out.PartnerType)
	require.Equal(t, PartnerStatusActive, out.Status)
}

func TestUpdatePartnerStatus_RejectsUnknownStatus(t *testing.T) {
	t.Parallel()

	uc := NewPartnerUsecase(&partnerRepoFake{})
	out, err := uc.UpdatePartnerStatus(context.Background(), "prt-1", "paused")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrInvalidPartnerStatus)
}

func TestCreatePartner_RejectsUnknownPartnerType(t *testing.T) {
	t.Parallel()

	uc := NewPartnerUsecase(&partnerRepoFake{})
	out, err := uc.CreatePartner(context.Background(), CreatePartnerCmd{
		TenantID:    "tenant-1",
		Name:        "Acme",
		PartnerType: "supplier",
	})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrInvalidPartnerType)
}
