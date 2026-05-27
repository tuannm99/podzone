package outputport

import (
	"context"

	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
)

type ProductSetupRepository interface {
	ListDrafts(ctx context.Context, storeID string) ([]catalogentity.ProductSetupDraft, error)
	GetDraftByID(ctx context.Context, storeID, id string) (*catalogentity.ProductSetupDraft, error)
	CreateDraft(ctx context.Context, draft catalogentity.ProductSetupDraft) (*catalogentity.ProductSetupDraft, error)
	ListCandidates(ctx context.Context, storeID string) ([]catalogentity.ProductSetupCandidate, error)
	GetCandidateByID(ctx context.Context, storeID, id string) (*catalogentity.ProductSetupCandidate, error)
	GetCandidateByDraftID(ctx context.Context, storeID, draftID string) (*catalogentity.ProductSetupCandidate, error)
	CreateCandidate(
		ctx context.Context,
		candidate catalogentity.ProductSetupCandidate,
	) (*catalogentity.ProductSetupCandidate, error)
	UpdateCandidateStatus(ctx context.Context, storeID, id, status string) (*catalogentity.ProductSetupCandidate, error)
}
