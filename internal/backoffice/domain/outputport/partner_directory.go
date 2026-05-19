package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
)

type PartnerDirectory interface {
	ListActivePartners(ctx context.Context, tenantID string) ([]entity.PartnerRoutingProfile, error)
}
