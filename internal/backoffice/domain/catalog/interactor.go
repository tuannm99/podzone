package catalog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
	cataloginputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/inputport"
	catalogoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/outputport"
)

type ProductSetupInteractor struct {
	repo catalogoutputport.ProductSetupRepository
}

var _ cataloginputport.ProductSetupUsecase = (*ProductSetupInteractor)(nil)

func NewProductSetupInteractor(repo catalogoutputport.ProductSetupRepository) cataloginputport.ProductSetupUsecase {
	return &ProductSetupInteractor{repo: repo}
}

func (i *ProductSetupInteractor) GetSnapshot(ctx context.Context, storeID string) (*catalogentity.ProductSetupSnapshot, error) {
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
	return &catalogentity.ProductSetupSnapshot{Drafts: drafts, Candidates: candidates}, nil
}

func (i *ProductSetupInteractor) CreateDraft(
	ctx context.Context,
	cmd cataloginputport.CreateProductSetupDraftCmd,
) (*catalogentity.ProductSetupDraft, error) {
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
	draft := catalogentity.ProductSetupDraft{
		ID:          uuid.NewString(),
		StoreID:     storeID,
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
	cmd cataloginputport.PromoteProductSetupCandidateCmd,
) (*catalogentity.ProductSetupCandidate, error) {
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

	candidate := catalogentity.ProductSetupCandidate{
		ID:              uuid.NewString(),
		StoreID:         storeID,
		DraftID:         draft.ID,
		Title:           draft.Name,
		SKU:             buildSKU(draft.Name),
		Partner:         draft.Partner,
		BaseCost:        draft.BaseCost,
		RetailPrice:     draft.RetailPrice,
		EstimatedMargin: estimateMargin(draft.BaseCost, draft.RetailPrice),
		Status:          catalogentity.ProductSetupCandidateStatusReady,
		Channel:         channel,
		UpdatedAt:       now,
		Variants: []catalogentity.ProductSetupVariant{
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
	storeID string,
	id string,
	status string,
) (*catalogentity.ProductSetupCandidate, error) {
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

func normalizeDraftStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case catalogentity.ProductSetupDraftStatusDraft,
		catalogentity.ProductSetupDraftStatusReadyForReview,
		catalogentity.ProductSetupDraftStatusPublishReady:
		return raw
	default:
		return ""
	}
}

func normalizeCandidateStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case catalogentity.ProductSetupCandidateStatusReady,
		catalogentity.ProductSetupCandidateStatusPublishedMock,
		catalogentity.ProductSetupCandidateStatusArchived:
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
