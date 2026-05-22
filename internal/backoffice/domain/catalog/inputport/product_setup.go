package inputport

import (
	"context"

	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
)

type CreateProductSetupDraftCmd struct {
	Name        string
	Partner     string
	BaseCost    string
	RetailPrice string
	Status      string
	Notes       string
}

type PromoteProductSetupCandidateCmd struct {
	DraftID            string
	Channel            string
	VariantColor       string
	VariantSize        string
	ArtworkChecklist   catalogentity.ProductSetupArtworkChecklist
	MerchandisingNotes string
}

type ProductSetupUsecase interface {
	GetSnapshot(ctx context.Context) (*catalogentity.ProductSetupSnapshot, error)
	CreateDraft(ctx context.Context, cmd CreateProductSetupDraftCmd) (*catalogentity.ProductSetupDraft, error)
	PromoteCandidate(ctx context.Context, cmd PromoteProductSetupCandidateCmd) (*catalogentity.ProductSetupCandidate, error)
	UpdateCandidateStatus(ctx context.Context, id, status string) (*catalogentity.ProductSetupCandidate, error)
}
