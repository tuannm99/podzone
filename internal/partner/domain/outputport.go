package domain

import "context"

type PartnerRepository interface {
	Create(ctx context.Context, partner Partner) (*Partner, error)
	GetByID(ctx context.Context, id string) (*Partner, error)
	List(ctx context.Context, query ListPartnersQuery) ([]Partner, error)
	Update(ctx context.Context, partner Partner) (*Partner, error)
	UpdateStatus(ctx context.Context, id, status string) (*Partner, error)
}
