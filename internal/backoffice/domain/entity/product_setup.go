package entity

import "time"

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
