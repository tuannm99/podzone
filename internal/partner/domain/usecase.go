package domain

import (
	"context"

	"github.com/tuannm99/podzone/pkg/collection"
)

type PartnerUsecase interface {
	CreatePartner(ctx context.Context, cmd CreatePartnerCmd) (*Partner, error)
	GetPartner(ctx context.Context, id string) (*Partner, error)
	ListPartners(ctx context.Context, query ListPartnersQuery) (collection.Page[Partner], error)
	UpdatePartner(ctx context.Context, cmd UpdatePartnerCmd) (*Partner, error)
	UpdatePartnerStatus(ctx context.Context, id, status string) (*Partner, error)
}
