package catalog

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ProductSetupDraftStatusDraft          = "draft"
	ProductSetupDraftStatusReadyForReview = "ready_for_review"
	ProductSetupDraftStatusPublishReady   = "publish_candidate"

	ProductSetupCandidateStatusReady         = "ready"
	ProductSetupCandidateStatusPublishedMock = "published_mock"
	ProductSetupCandidateStatusArchived      = "archived"
)

type ProductSetupDraft struct {
	ID          string    `json:"id"`
	StoreID     string    `json:"storeId"`
	Name        string    `json:"name"`
	Partner     string    `json:"partner"`
	BaseCost    string    `json:"baseCost"`
	RetailPrice string    `json:"retailPrice"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProductSetupVariant struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Color  string `json:"color"`
	Size   string `json:"size"`
	Status string `json:"status"`
}

type ProductSetupArtworkChecklist struct {
	FrontArtwork     bool `json:"frontArtwork"`
	BackArtwork      bool `json:"backArtwork"`
	MockupReady      bool `json:"mockupReady"`
	PrintSpecChecked bool `json:"printSpecChecked"`
}

type ProductSetupCandidate struct {
	ID                 string                       `json:"id"`
	StoreID            string                       `json:"storeId"`
	DraftID            string                       `json:"draftId"`
	Title              string                       `json:"title"`
	SKU                string                       `json:"sku"`
	Partner            string                       `json:"partner"`
	BaseCost           string                       `json:"baseCost"`
	RetailPrice        string                       `json:"retailPrice"`
	EstimatedMargin    string                       `json:"estimatedMargin"`
	Status             string                       `json:"status"`
	Channel            string                       `json:"channel"`
	UpdatedAt          time.Time                    `json:"updatedAt"`
	Variants           []ProductSetupVariant        `json:"variants"`
	ArtworkChecklist   ProductSetupArtworkChecklist `json:"artworkChecklist"`
	MerchandisingNotes string                       `json:"merchandisingNotes"`
}

type ProductSetupSnapshot struct {
	Drafts     []ProductSetupDraft     `json:"drafts"`
	Candidates []ProductSetupCandidate `json:"candidates"`
}

func NewProductSetupDraft(
	cmd CreateProductSetupDraftCmd,
	status string,
	now time.Time,
) (ProductSetupDraft, []DomainEvent, error) {
	storeID := strings.TrimSpace(cmd.StoreID)
	if storeID == "" {
		return ProductSetupDraft{}, nil, fmt.Errorf("product setup store id is required")
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return ProductSetupDraft{}, nil, fmt.Errorf("product setup name is required")
	}
	status = normalizeDraftStatus(status)
	if status == "" {
		return ProductSetupDraft{}, nil, fmt.Errorf("invalid product setup draft status")
	}
	now = now.UTC()
	draft := ProductSetupDraft{
		ID:          uuid.NewString(),
		StoreID:     storeID,
		Name:        name,
		Partner:     strings.TrimSpace(cmd.Partner),
		BaseCost:    fallbackValue(cmd.BaseCost, "TBD"),
		RetailPrice: fallbackValue(cmd.RetailPrice, "TBD"),
		Status:      status,
		Notes:       strings.TrimSpace(cmd.Notes),
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC(),
	}
	if draft.Partner == "" {
		draft.Partner = "Unassigned"
	}
	return draft, []DomainEvent{
		ProductSetupDraftCreated{
			DraftID:    draft.ID,
			StoreID:    draft.StoreID,
			Name:       draft.Name,
			OccurredAt: now,
		},
	}, nil
}

func (d ProductSetupDraft) PromoteCandidate(
	cmd PromoteProductSetupCandidateCmd,
	now time.Time,
) (ProductSetupCandidate, []DomainEvent, error) {
	if strings.TrimSpace(d.ID) == "" {
		return ProductSetupCandidate{}, nil, fmt.Errorf("product setup draft id is required")
	}
	if strings.TrimSpace(d.StoreID) == "" {
		return ProductSetupCandidate{}, nil, fmt.Errorf("product setup store id is required")
	}
	if strings.TrimSpace(d.Name) == "" {
		return ProductSetupCandidate{}, nil, fmt.Errorf("product setup name is required")
	}
	color := fallbackValue(cmd.VariantColor, "Black")
	size := fallbackValue(cmd.VariantSize, "M")
	notes := strings.TrimSpace(cmd.MerchandisingNotes)
	if notes == "" {
		if d.Notes != "" {
			notes = d.Notes
		} else {
			notes = "No extra merchandising notes yet."
		}
	}

	candidate := ProductSetupCandidate{
		ID:              uuid.NewString(),
		StoreID:         d.StoreID,
		DraftID:         d.ID,
		Title:           d.Name,
		SKU:             buildSKU(d.Name),
		Partner:         d.Partner,
		BaseCost:        d.BaseCost,
		RetailPrice:     d.RetailPrice,
		EstimatedMargin: estimateMargin(d.BaseCost, d.RetailPrice),
		Status:          ProductSetupCandidateStatusReady,
		Channel:         fallbackValue(cmd.Channel, "website_store"),
		UpdatedAt:       now.UTC(),
		Variants: []ProductSetupVariant{
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
	return candidate, []DomainEvent{
		ProductSetupCandidatePromoted{
			CandidateID: candidate.ID,
			DraftID:     d.ID,
			StoreID:     d.StoreID,
			OccurredAt:  now.UTC(),
		},
	}, nil
}

func normalizeDraftStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case ProductSetupDraftStatusDraft,
		ProductSetupDraftStatusReadyForReview,
		ProductSetupDraftStatusPublishReady:
		return strings.TrimSpace(raw)
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
