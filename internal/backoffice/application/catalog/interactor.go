package catalog

import (
	"context"
	"fmt"
	"strings"

	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	"github.com/tuannm99/podzone/pkg/ddd"
)

type Interactor struct {
	repo  catalogctx.ProductSetupRepository
	ids   ddd.IDGenerator
	clock ddd.Clock
}

var _ catalogctx.ProductSetupUsecase = (*Interactor)(nil)

func NewInteractor(
	repo catalogctx.ProductSetupRepository,
	ids ddd.IDGenerator,
	clock ddd.Clock,
) catalogctx.ProductSetupUsecase {
	return &Interactor{repo: repo, ids: ids, clock: clock}
}

func (i *Interactor) GetSnapshot(
	ctx context.Context,
	storeID string,
) (*catalogctx.ProductSetupSnapshot, error) {
	storeID = strings.TrimSpace(storeID)
	if storeID == "" {
		return nil, ddd.NewDomainError("STORE_ID_REQUIRED", "store id is required")
	}
	drafts, err := i.repo.ListDrafts(ctx, storeID)
	if err != nil {
		return nil, err
	}
	candidates, err := i.repo.ListCandidates(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return &catalogctx.ProductSetupSnapshot{Drafts: drafts, Candidates: candidates}, nil
}

func (i *Interactor) CreateDraft(
	ctx context.Context,
	cmd catalogctx.CreateProductSetupDraftCmd,
) (*catalogctx.ProductSetupDraft, error) {
	storeID := strings.TrimSpace(cmd.StoreID)
	if storeID == "" {
		return nil, ddd.NewDomainError("STORE_ID_REQUIRED", "store id is required")
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ddd.NewDomainError("PRODUCT_NAME_REQUIRED", "product name is required")
	}
	status := catalogctx.NormalizeDraftStatus(cmd.Status)
	if status == "" {
		return nil, ddd.NewDomainError("PRODUCT_DRAFT_STATUS_INVALID", "invalid draft status")
	}
	cmd.StoreID = storeID
	cmd.Name = name
	draftID, err := i.ids.NewID("product_setup_draft")
	if err != nil {
		return nil, err
	}
	aggregate, _, err := catalogctx.NewProductSetupDraft(draftID.String(), cmd, status, i.clock.Now())
	if err != nil {
		return nil, err
	}
	return i.repo.CreateDraft(ctx, aggregate.Snapshot())
}

func (i *Interactor) PromoteCandidate(
	ctx context.Context,
	cmd catalogctx.PromoteProductSetupCandidateCmd,
) (*catalogctx.ProductSetupCandidate, error) {
	storeID := strings.TrimSpace(cmd.StoreID)
	if storeID == "" {
		return nil, ddd.NewDomainError("STORE_ID_REQUIRED", "store id is required")
	}
	draftID := strings.TrimSpace(cmd.DraftID)
	if draftID == "" {
		return nil, ddd.NewDomainError("PRODUCT_DRAFT_ID_REQUIRED", "draft id is required")
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
		return nil, ddd.NewDomainError(
			"PRODUCT_CANDIDATE_EXISTS",
			fmt.Sprintf("catalog candidate for %s already exists", draft.Name),
		)
	}
	candidateID, err := i.ids.NewID("product_setup_candidate")
	if err != nil {
		return nil, err
	}
	variantID, err := i.ids.NewID("product_setup_variant")
	if err != nil {
		return nil, err
	}
	aggregate, err := catalogctx.RehydrateProductSetupDraft(*draft)
	if err != nil {
		return nil, err
	}
	candidate, _, err := aggregate.PromoteCandidate(
		candidateID.String(),
		variantID.String(),
		cmd,
		i.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	return i.repo.CreateCandidate(ctx, candidate)
}

func (i *Interactor) UpdateCandidateStatus(
	ctx context.Context,
	storeID string,
	id string,
	status string,
) (*catalogctx.ProductSetupCandidate, error) {
	storeID = strings.TrimSpace(storeID)
	if storeID == "" {
		return nil, ddd.NewDomainError("STORE_ID_REQUIRED", "store id is required")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ddd.NewDomainError("PRODUCT_CANDIDATE_ID_REQUIRED", "candidate id is required")
	}
	candidate, err := i.repo.GetCandidateByID(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, ddd.ErrNotFound
	}
	aggregate, err := catalogctx.RehydrateProductSetupCandidate(*candidate)
	if err != nil {
		return nil, err
	}
	if err := aggregate.ChangeStatus(status, i.clock.Now()); err != nil {
		return nil, err
	}
	updated := aggregate.Snapshot()
	return i.repo.UpdateCandidateStatus(ctx, storeID, id, updated.Status)
}
