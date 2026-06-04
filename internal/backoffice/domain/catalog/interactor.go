package catalog

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ProductSetupInteractor struct {
	repo ProductSetupRepository
}

var _ ProductSetupUsecase = (*ProductSetupInteractor)(nil)

func NewProductSetupInteractor(repo ProductSetupRepository) ProductSetupUsecase {
	return &ProductSetupInteractor{repo: repo}
}

func (i *ProductSetupInteractor) GetSnapshot(ctx context.Context, storeID string) (*ProductSetupSnapshot, error) {
	storeID = strings.TrimSpace(storeID)
	if storeID == "" {
		return nil, fmt.Errorf("store id is required")
	}

	drafts, err := i.repo.ListDrafts(ctx, storeID)
	if err != nil {
		return nil, err
	}
	candidates, err := i.repo.ListCandidates(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return &ProductSetupSnapshot{Drafts: drafts, Candidates: candidates}, nil
}

func (i *ProductSetupInteractor) CreateDraft(
	ctx context.Context,
	cmd CreateProductSetupDraftCmd,
) (*ProductSetupDraft, error) {
	storeID := strings.TrimSpace(cmd.StoreID)
	if storeID == "" {
		return nil, fmt.Errorf("store id is required")
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, fmt.Errorf("product name is required")
	}

	status := normalizeDraftStatus(cmd.Status)
	if status == "" {
		return nil, fmt.Errorf("invalid draft status")
	}

	now := time.Now().UTC()
	cmd.StoreID = storeID
	cmd.Name = name
	draft, _, err := NewProductSetupDraft(cmd, status, now)
	if err != nil {
		return nil, err
	}
	return i.repo.CreateDraft(ctx, draft)
}

func (i *ProductSetupInteractor) PromoteCandidate(
	ctx context.Context,
	cmd PromoteProductSetupCandidateCmd,
) (*ProductSetupCandidate, error) {
	storeID := strings.TrimSpace(cmd.StoreID)
	if storeID == "" {
		return nil, fmt.Errorf("store id is required")
	}

	draftID := strings.TrimSpace(cmd.DraftID)
	if draftID == "" {
		return nil, fmt.Errorf("draft id is required")
	}

	draft, err := i.repo.GetDraftByID(ctx, storeID, draftID)
	if err != nil {
		return nil, err
	}

	existing, err := i.repo.GetCandidateByDraftID(ctx, storeID, draftID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("catalog candidate for %s already exists", draft.Name)
	}

	candidate, _, err := draft.PromoteCandidate(cmd, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return i.repo.CreateCandidate(ctx, candidate)
}

func (i *ProductSetupInteractor) UpdateCandidateStatus(
	ctx context.Context,
	storeID string,
	id string,
	status string,
) (*ProductSetupCandidate, error) {
	storeID = strings.TrimSpace(storeID)
	if storeID == "" {
		return nil, fmt.Errorf("store id is required")
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("candidate id is required")
	}

	status = normalizeCandidateStatus(status)
	if status == "" {
		return nil, fmt.Errorf("invalid candidate status")
	}
	return i.repo.UpdateCandidateStatus(ctx, storeID, id, status)
}

func normalizeCandidateStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case ProductSetupCandidateStatusReady,
		ProductSetupCandidateStatusPublishedMock,
		ProductSetupCandidateStatusArchived:
		return raw
	default:
		return ""
	}
}
