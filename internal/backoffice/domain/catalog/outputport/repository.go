package outputport

import (
	"context"

	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
)

type ProductSetupRepository interface {
	ListDrafts(ctx context.Context) ([]catalogentity.ProductSetupDraft, error)
	GetDraftByID(ctx context.Context, id string) (*catalogentity.ProductSetupDraft, error)
	CreateDraft(ctx context.Context, draft catalogentity.ProductSetupDraft) (*catalogentity.ProductSetupDraft, error)
	ListCandidates(ctx context.Context) ([]catalogentity.ProductSetupCandidate, error)
	GetCandidateByID(ctx context.Context, id string) (*catalogentity.ProductSetupCandidate, error)
	GetCandidateByDraftID(ctx context.Context, draftID string) (*catalogentity.ProductSetupCandidate, error)
	CreateCandidate(ctx context.Context, candidate catalogentity.ProductSetupCandidate) (*catalogentity.ProductSetupCandidate, error)
	UpdateCandidateStatus(ctx context.Context, id, status string) (*catalogentity.ProductSetupCandidate, error)
}
