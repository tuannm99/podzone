package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
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
	ArtworkChecklist   entity.ProductSetupArtworkChecklist
	MerchandisingNotes string
}

type ProductSetupUsecase interface {
	GetSnapshot(ctx context.Context) (*entity.ProductSetupSnapshot, error)
	CreateDraft(ctx context.Context, cmd CreateProductSetupDraftCmd) (*entity.ProductSetupDraft, error)
	PromoteCandidate(ctx context.Context, cmd PromoteProductSetupCandidateCmd) (*entity.ProductSetupCandidate, error)
	UpdateCandidateStatus(ctx context.Context, id, status string) (*entity.ProductSetupCandidate, error)
}
