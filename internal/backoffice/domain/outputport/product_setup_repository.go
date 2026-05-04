package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
)

type ProductSetupRepository interface {
	ListDrafts(ctx context.Context) ([]entity.ProductSetupDraft, error)
	GetDraftByID(ctx context.Context, id string) (*entity.ProductSetupDraft, error)
	CreateDraft(ctx context.Context, draft entity.ProductSetupDraft) (*entity.ProductSetupDraft, error)
	ListCandidates(ctx context.Context) ([]entity.ProductSetupCandidate, error)
	GetCandidateByDraftID(ctx context.Context, draftID string) (*entity.ProductSetupCandidate, error)
	CreateCandidate(ctx context.Context, candidate entity.ProductSetupCandidate) (*entity.ProductSetupCandidate, error)
	UpdateCandidateStatus(ctx context.Context, id, status string) (*entity.ProductSetupCandidate, error)
}
