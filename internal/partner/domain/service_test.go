package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domain "github.com/tuannm99/podzone/internal/partner/domain"
	domainmocks "github.com/tuannm99/podzone/internal/partner/domain/mocks"
)

func TestCreatePartner_NormalizesCodeFromName(t *testing.T) {
	t.Parallel()

	repo := domainmocks.NewMockPartnerRepository(t)
	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, partner domain.Partner) (*domain.Partner, error) {
			return &partner, nil
		}).
		Once()

	uc := domain.NewPartnerUsecase(repo)
	out, err := uc.CreatePartner(context.Background(), domain.CreatePartnerCmd{
		TenantID:    "tenant-1",
		Name:        "Acme Supply Co",
		PartnerType: domain.PartnerTypePrintOnDemand,
	})
	require.NoError(t, err)
	require.Equal(t, "acme-supply-co", out.Code)
	require.Equal(t, domain.PartnerTypePrintOnDemand, out.PartnerType)
	require.Equal(t, domain.PartnerStatusActive, out.Status)
}

func TestUpdatePartnerStatus_RejectsUnknownStatus(t *testing.T) {
	t.Parallel()

	repo := domainmocks.NewMockPartnerRepository(t)
	uc := domain.NewPartnerUsecase(repo)
	out, err := uc.UpdatePartnerStatus(context.Background(), "prt-1", "paused")
	require.Nil(t, out)
	require.ErrorIs(t, err, domain.ErrInvalidPartnerStatus)
}

func TestCreatePartner_RejectsUnknownPartnerType(t *testing.T) {
	t.Parallel()

	repo := domainmocks.NewMockPartnerRepository(t)
	uc := domain.NewPartnerUsecase(repo)
	out, err := uc.CreatePartner(context.Background(), domain.CreatePartnerCmd{
		TenantID:    "tenant-1",
		Name:        "Acme",
		PartnerType: "supplier",
	})
	require.Nil(t, out)
	require.ErrorIs(t, err, domain.ErrInvalidPartnerType)
}
