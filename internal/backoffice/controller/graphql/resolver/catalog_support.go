package resolver

import (
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
)

func toEntityArtworkChecklistInput(input *model.ProductSetupArtworkChecklistInput) entity.ProductSetupArtworkChecklist {
	if input == nil {
		return entity.ProductSetupArtworkChecklist{}
	}
	return entity.ProductSetupArtworkChecklist{
		FrontArtwork:     input.FrontArtwork,
		BackArtwork:      input.BackArtwork,
		MockupReady:      input.MockupReady,
		PrintSpecChecked: input.PrintSpecChecked,
	}
}

func toGraphQLProductSetupSnapshot(snapshot entity.ProductSetupSnapshot) *model.ProductSetupSnapshot {
	drafts := make([]*model.ProductSetupDraft, 0, len(snapshot.Drafts))
	for _, draft := range snapshot.Drafts {
		drafts = append(drafts, toGraphQLProductSetupDraft(draft))
	}
	candidates := make([]*model.ProductSetupCandidate, 0, len(snapshot.Candidates))
	for _, candidate := range snapshot.Candidates {
		candidates = append(candidates, toGraphQLProductSetupCandidate(candidate))
	}
	return &model.ProductSetupSnapshot{
		Drafts:     drafts,
		Candidates: candidates,
	}
}

func toGraphQLProductSetupDraft(draft entity.ProductSetupDraft) *model.ProductSetupDraft {
	return &model.ProductSetupDraft{
		ID:          draft.ID,
		Name:        draft.Name,
		Partner:     draft.Partner,
		BaseCost:    draft.BaseCost,
		RetailPrice: draft.RetailPrice,
		Status:      draft.Status,
		Notes:       draft.Notes,
		CreatedAt:   draft.CreatedAt,
		UpdatedAt:   draft.UpdatedAt,
	}
}

func toGraphQLProductSetupCandidate(candidate entity.ProductSetupCandidate) *model.ProductSetupCandidate {
	variants := make([]*model.ProductSetupVariant, 0, len(candidate.Variants))
	for _, variant := range candidate.Variants {
		variants = append(variants, &model.ProductSetupVariant{
			ID:     variant.ID,
			Label:  variant.Label,
			Color:  variant.Color,
			Size:   variant.Size,
			Status: variant.Status,
		})
	}
	return &model.ProductSetupCandidate{
		ID:              candidate.ID,
		DraftID:         candidate.DraftID,
		Title:           candidate.Title,
		Sku:             candidate.SKU,
		Partner:         candidate.Partner,
		BaseCost:        candidate.BaseCost,
		RetailPrice:     candidate.RetailPrice,
		EstimatedMargin: candidate.EstimatedMargin,
		Status:          candidate.Status,
		Channel:         candidate.Channel,
		UpdatedAt:       candidate.UpdatedAt,
		Variants:        variants,
		ArtworkChecklist: &model.ProductSetupArtworkChecklist{
			FrontArtwork:     candidate.ArtworkChecklist.FrontArtwork,
			BackArtwork:      candidate.ArtworkChecklist.BackArtwork,
			MockupReady:      candidate.ArtworkChecklist.MockupReady,
			PrintSpecChecked: candidate.ArtworkChecklist.PrintSpecChecked,
		},
		MerchandisingNotes: candidate.MerchandisingNotes,
	}
}
