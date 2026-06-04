package catalog

import (
	"context"
)

type CreateProductSetupDraftCmd struct {
	StoreID     string
	Name        string
	Partner     string
	BaseCost    string
	RetailPrice string
	Status      string
	Notes       string
}

type PromoteProductSetupCandidateCmd struct {
	StoreID            string
	DraftID            string
	Channel            string
	VariantColor       string
	VariantSize        string
	ArtworkChecklist   ProductSetupArtworkChecklist
	MerchandisingNotes string
}

type ProductSetupUsecase interface {
	GetSnapshot(ctx context.Context, storeID string) (*ProductSetupSnapshot, error)
	CreateDraft(ctx context.Context, cmd CreateProductSetupDraftCmd) (*ProductSetupDraft, error)
	PromoteCandidate(
		ctx context.Context,
		cmd PromoteProductSetupCandidateCmd,
	) (*ProductSetupCandidate, error)
	UpdateCandidateStatus(ctx context.Context, storeID, id, status string) (*ProductSetupCandidate, error)
}
