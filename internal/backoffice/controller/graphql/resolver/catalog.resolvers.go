package resolver

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
)

// CreateProductSetupDraft is the resolver for the createProductSetupDraft field.
func (r *mutationResolver) CreateProductSetupDraft(ctx context.Context, input model.CreateProductSetupDraftInput) (*model.ProductSetupDraft, error) {
	draft, err := r.ProductSetupUsecase.CreateDraft(ctx, inputport.CreateProductSetupDraftCmd{
		Name:        input.Name,
		Partner:     input.Partner,
		BaseCost:    input.BaseCost,
		RetailPrice: input.RetailPrice,
		Status:      input.Status,
		Notes:       input.Notes,
	})
	if err != nil {
		return nil, err
	}
	return toGraphQLProductSetupDraft(*draft), nil
}

// PromoteProductSetupCandidate is the resolver for the promoteProductSetupCandidate field.
func (r *mutationResolver) PromoteProductSetupCandidate(ctx context.Context, input model.PromoteProductSetupCandidateInput) (*model.ProductSetupCandidate, error) {
	candidate, err := r.ProductSetupUsecase.PromoteCandidate(ctx, inputport.PromoteProductSetupCandidateCmd{
		DraftID:            input.DraftID,
		Channel:            input.Channel,
		VariantColor:       input.VariantColor,
		VariantSize:        input.VariantSize,
		ArtworkChecklist:   toEntityArtworkChecklistInput(input.ArtworkChecklist),
		MerchandisingNotes: input.MerchandisingNotes,
	})
	if err != nil {
		return nil, err
	}
	return toGraphQLProductSetupCandidate(*candidate), nil
}

// UpdateProductSetupCandidateStatus is the resolver for the updateProductSetupCandidateStatus field.
func (r *mutationResolver) UpdateProductSetupCandidateStatus(ctx context.Context, id string, status string) (*model.ProductSetupCandidate, error) {
	candidate, err := r.ProductSetupUsecase.UpdateCandidateStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}
	return toGraphQLProductSetupCandidate(*candidate), nil
}

// ProductSetupSnapshot is the resolver for the productSetupSnapshot field.
func (r *queryResolver) ProductSetupSnapshot(ctx context.Context) (*model.ProductSetupSnapshot, error) {
	snapshot, err := r.ProductSetupUsecase.GetSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	return toGraphQLProductSetupSnapshot(*snapshot), nil
}
