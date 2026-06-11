package catalog

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
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

type ProductSetupDraftAggregate struct {
	aggregate ddd.AggregateBase
	draft     ProductSetupDraft
}

var _ ddd.AggregateRoot = (*ProductSetupDraftAggregate)(nil)

type ProductSetupCandidateAggregate struct {
	aggregate ddd.AggregateBase
	candidate ProductSetupCandidate
}

var _ ddd.AggregateRoot = (*ProductSetupCandidateAggregate)(nil)

func NewProductSetupDraft(
	id string,
	cmd CreateProductSetupDraftCmd,
	status string,
	now time.Time,
) (*ProductSetupDraftAggregate, []DomainEvent, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, ErrDraftIDRequired
	}
	storeID := strings.TrimSpace(cmd.StoreID)
	if storeID == "" {
		return nil, nil, ErrStoreIDRequired
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, nil, ErrProductNameRequired
	}
	status = NormalizeDraftStatus(status)
	if status == "" {
		return nil, nil, ErrDraftStatusInvalid
	}
	now = now.UTC()
	aggregate, err := newDraftAggregate(id)
	if err != nil {
		return nil, nil, err
	}
	draft := ProductSetupDraft{
		ID:          id,
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
	root := &ProductSetupDraftAggregate{aggregate: aggregate, draft: draft}
	return root, []DomainEvent{
		ProductSetupDraftCreated{
			DraftID:    draft.ID,
			StoreID:    draft.StoreID,
			Name:       draft.Name,
			OccurredAt: now,
		},
	}, nil
}

func RehydrateProductSetupDraft(draft ProductSetupDraft) (*ProductSetupDraftAggregate, error) {
	aggregate, err := newDraftAggregate(draft.ID)
	if err != nil {
		return nil, err
	}
	return &ProductSetupDraftAggregate{aggregate: aggregate, draft: draft}, nil
}

func (d *ProductSetupDraftAggregate) AggregateID() ddd.ID {
	if d == nil {
		return ""
	}
	return d.aggregate.AggregateID()
}

func (d *ProductSetupDraftAggregate) AggregateVersion() ddd.Version {
	if d == nil {
		return 0
	}
	return d.aggregate.AggregateVersion()
}

func (d *ProductSetupDraftAggregate) Snapshot() ProductSetupDraft {
	return d.draft
}

func (d *ProductSetupDraftAggregate) PullEvents() []DomainEvent {
	if d == nil {
		return nil
	}
	return d.aggregate.PullEvents()
}

func (d *ProductSetupDraftAggregate) PromoteCandidate(
	candidateID string,
	variantID string,
	cmd PromoteProductSetupCandidateCmd,
	now time.Time,
) (ProductSetupCandidate, []DomainEvent, error) {
	draft := d.draft
	candidateID = strings.TrimSpace(candidateID)
	if candidateID == "" {
		return ProductSetupCandidate{}, nil, ErrCandidateIDRequired
	}
	variantID = strings.TrimSpace(variantID)
	if variantID == "" {
		return ProductSetupCandidate{}, nil, ErrVariantIDRequired
	}
	if strings.TrimSpace(draft.ID) == "" {
		return ProductSetupCandidate{}, nil, ErrDraftIDRequired
	}
	if strings.TrimSpace(draft.StoreID) == "" {
		return ProductSetupCandidate{}, nil, ErrStoreIDRequired
	}
	if strings.TrimSpace(draft.Name) == "" {
		return ProductSetupCandidate{}, nil, ErrProductNameRequired
	}
	color := fallbackValue(cmd.VariantColor, "Black")
	size := fallbackValue(cmd.VariantSize, "M")
	notes := strings.TrimSpace(cmd.MerchandisingNotes)
	if notes == "" {
		if draft.Notes != "" {
			notes = draft.Notes
		} else {
			notes = "No extra merchandising notes yet."
		}
	}

	candidate := ProductSetupCandidate{
		ID:              candidateID,
		StoreID:         draft.StoreID,
		DraftID:         draft.ID,
		Title:           draft.Name,
		SKU:             buildSKU(draft.Name, candidateID),
		Partner:         draft.Partner,
		BaseCost:        draft.BaseCost,
		RetailPrice:     draft.RetailPrice,
		EstimatedMargin: estimateMargin(draft.BaseCost, draft.RetailPrice),
		Status:          ProductSetupCandidateStatusReady,
		Channel:         fallbackValue(cmd.Channel, "website_store"),
		UpdatedAt:       now.UTC(),
		Variants: []ProductSetupVariant{
			{
				ID:     variantID,
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
			DraftID:     draft.ID,
			StoreID:     draft.StoreID,
			OccurredAt:  now.UTC(),
		},
	}, nil
}

func RehydrateProductSetupCandidate(
	candidate ProductSetupCandidate,
) (*ProductSetupCandidateAggregate, error) {
	id, err := ddd.ParseID(candidate.ID)
	if err != nil {
		return nil, err
	}
	aggregate, err := ddd.NewAggregateBase(id, 0)
	if err != nil {
		return nil, err
	}
	return &ProductSetupCandidateAggregate{aggregate: aggregate, candidate: candidate}, nil
}

func (c *ProductSetupCandidateAggregate) AggregateID() ddd.ID {
	if c == nil {
		return ""
	}
	return c.aggregate.AggregateID()
}

func (c *ProductSetupCandidateAggregate) AggregateVersion() ddd.Version {
	if c == nil {
		return 0
	}
	return c.aggregate.AggregateVersion()
}

func (c *ProductSetupCandidateAggregate) Snapshot() ProductSetupCandidate {
	return c.candidate
}

func (c *ProductSetupCandidateAggregate) PullEvents() []DomainEvent {
	if c == nil {
		return nil
	}
	return c.aggregate.PullEvents()
}

func (c *ProductSetupCandidateAggregate) ChangeStatus(status string, now time.Time) error {
	status = NormalizeCandidateStatus(status)
	if status == "" {
		return ddd.NewDomainError("PRODUCT_CANDIDATE_STATUS_INVALID", "invalid candidate status")
	}
	c.candidate.Status = status
	c.candidate.UpdatedAt = now.UTC()
	return nil
}

func newDraftAggregate(rawID string) (ddd.AggregateBase, error) {
	id, err := ddd.ParseID(rawID)
	if err != nil {
		return ddd.AggregateBase{}, err
	}
	return ddd.NewAggregateBase(id, 0)
}

func NormalizeDraftStatus(raw string) string {
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

func buildSKU(name string, suffix string) string {
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
	suffix = strings.Trim(strings.ToLower(strings.TrimSpace(suffix)), "_-")
	if len(suffix) > 8 {
		suffix = suffix[len(suffix)-8:]
	}
	if suffix == "" {
		suffix = "draft"
	}
	return fmt.Sprintf("%s-%s", slug, suffix)
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
