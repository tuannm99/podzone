package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
)

type ProductSetupInteractor struct {
	repo outputport.ProductSetupRepository
}

func NewProductSetupInteractor(repo outputport.ProductSetupRepository) inputport.ProductSetupUsecase {
	return &ProductSetupInteractor{repo: repo}
}

func (i *ProductSetupInteractor) GetSnapshot(ctx context.Context) (*entity.ProductSetupSnapshot, error) {
	drafts, err := i.repo.ListDrafts(ctx)
	if err != nil {
		return nil, err
	}
	candidates, err := i.repo.ListCandidates(ctx)
	if err != nil {
		return nil, err
	}
	return &entity.ProductSetupSnapshot{Drafts: drafts, Candidates: candidates}, nil
}

func (i *ProductSetupInteractor) CreateDraft(
	ctx context.Context,
	cmd inputport.CreateProductSetupDraftCmd,
) (*entity.ProductSetupDraft, error) {
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, fmt.Errorf("product name is required")
	}

	status := normalizeDraftStatus(cmd.Status)
	if status == "" {
		return nil, fmt.Errorf("invalid draft status")
	}

	now := time.Now().UTC()
	draft := entity.ProductSetupDraft{
		ID:          uuid.NewString(),
		Name:        name,
		Partner:     strings.TrimSpace(cmd.Partner),
		BaseCost:    fallbackValue(cmd.BaseCost, "TBD"),
		RetailPrice: fallbackValue(cmd.RetailPrice, "TBD"),
		Status:      status,
		Notes:       strings.TrimSpace(cmd.Notes),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if draft.Partner == "" {
		draft.Partner = "Unassigned"
	}
	return i.repo.CreateDraft(ctx, draft)
}

func (i *ProductSetupInteractor) PromoteCandidate(
	ctx context.Context,
	cmd inputport.PromoteProductSetupCandidateCmd,
) (*entity.ProductSetupCandidate, error) {
	draftID := strings.TrimSpace(cmd.DraftID)
	if draftID == "" {
		return nil, fmt.Errorf("draft id is required")
	}

	draft, err := i.repo.GetDraftByID(ctx, draftID)
	if err != nil {
		return nil, err
	}

	existing, err := i.repo.GetCandidateByDraftID(ctx, draftID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("catalog candidate for %s already exists", draft.Name)
	}

	now := time.Now().UTC()
	color := fallbackValue(cmd.VariantColor, "Black")
	size := fallbackValue(cmd.VariantSize, "M")
	channel := fallbackValue(cmd.Channel, "website_store")
	notes := strings.TrimSpace(cmd.MerchandisingNotes)
	if notes == "" {
		if draft.Notes != "" {
			notes = draft.Notes
		} else {
			notes = "No extra merchandising notes yet."
		}
	}

	candidate := entity.ProductSetupCandidate{
		ID:              uuid.NewString(),
		DraftID:         draft.ID,
		Title:           draft.Name,
		SKU:             buildSKU(draft.Name),
		Partner:         draft.Partner,
		BaseCost:        draft.BaseCost,
		RetailPrice:     draft.RetailPrice,
		EstimatedMargin: estimateMargin(draft.BaseCost, draft.RetailPrice),
		Status:          entity.ProductSetupCandidateStatusReady,
		Channel:         channel,
		UpdatedAt:       now,
		Variants: []entity.ProductSetupVariant{
			{
				ID:     uuid.NewString(),
				Label:  fmt.Sprintf("%s / %s", color, size),
				Color:  color,
				Size:   size,
				Status: "ready",
			},
		},
		ArtworkChecklist:   cmd.ArtworkChecklist,
		MerchandisingNotes: notes,
	}
	return i.repo.CreateCandidate(ctx, candidate)
}

func (i *ProductSetupInteractor) UpdateCandidateStatus(
	ctx context.Context,
	id string,
	status string,
) (*entity.ProductSetupCandidate, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("candidate id is required")
	}

	status = normalizeCandidateStatus(status)
	if status == "" {
		return nil, fmt.Errorf("invalid candidate status")
	}
	return i.repo.UpdateCandidateStatus(ctx, id, status)
}

func normalizeDraftStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case entity.ProductSetupDraftStatusDraft,
		entity.ProductSetupDraftStatusReadyForReview,
		entity.ProductSetupDraftStatusPublishReady:
		return raw
	default:
		return ""
	}
}

func normalizeCandidateStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case entity.ProductSetupCandidateStatusReady,
		entity.ProductSetupCandidateStatusPublishedMock,
		entity.ProductSetupCandidateStatusArchived:
		return raw
	default:
		return ""
	}
}

func fallbackValue(raw, fallback string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	return raw
}

func buildSKU(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '-'
		}
	}, slug)
	slug = strings.Trim(strings.Join(strings.FieldsFunc(slug, func(r rune) bool { return r == '-' }), "-"), "-")
	if slug == "" {
		slug = "product"
	}
	return fmt.Sprintf("%s-%s", slug, strings.ToLower(uuid.NewString()[:8]))
}

func estimateMargin(baseCost, retailPrice string) string {
	parse := func(value string) (float64, bool) {
		cleaned := strings.Map(func(r rune) rune {
			if (r >= '0' && r <= '9') || r == '.' {
				return r
			}
			return -1
		}, value)
		if cleaned == "" {
			return 0, false
		}
		var whole float64
		_, err := fmt.Sscanf(cleaned, "%f", &whole)
		return whole, err == nil
	}

	base, okBase := parse(baseCost)
	retail, okRetail := parse(retailPrice)
	if !okBase || !okRetail {
		return "TBD"
	}
	return fmt.Sprintf("$%.2f", retail-base)
}
