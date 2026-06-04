package catalog

import (
	"context"
)

type ProductSetupRepository interface {
	ListDrafts(ctx context.Context, storeID string) ([]ProductSetupDraft, error)
	GetDraftByID(ctx context.Context, storeID, id string) (*ProductSetupDraft, error)
	CreateDraft(ctx context.Context, draft ProductSetupDraft) (*ProductSetupDraft, error)
	ListCandidates(ctx context.Context, storeID string) ([]ProductSetupCandidate, error)
	GetCandidateByID(ctx context.Context, storeID, id string) (*ProductSetupCandidate, error)
	GetCandidateByDraftID(ctx context.Context, storeID, draftID string) (*ProductSetupCandidate, error)
	CreateCandidate(
		ctx context.Context,
		candidate ProductSetupCandidate,
	) (*ProductSetupCandidate, error)
	UpdateCandidateStatus(ctx context.Context, storeID, id, status string) (*ProductSetupCandidate, error)
}
