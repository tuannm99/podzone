package resolver

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	cataloginputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
)

// CreateProductSetupDraft is the resolver for the createProductSetupDraft field.
func (r *mutationResolver) CreateProductSetupDraft(
	ctx context.Context,
	input model.CreateProductSetupDraftInput,
) (*model.ProductSetupDraft, error) {
	storeID, err := requiredStoreID(ctx)
	if err != nil {
		return nil, err
	}

	draft, err := r.ProductSetupUsecase.CreateDraft(ctx, cataloginputport.CreateProductSetupDraftCmd{
		StoreID:     storeID,
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
func (r *mutationResolver) PromoteProductSetupCandidate(
	ctx context.Context,
	input model.PromoteProductSetupCandidateInput,
) (*model.ProductSetupCandidate, error) {
	storeID, err := requiredStoreID(ctx)
	if err != nil {
		return nil, err
	}

	candidate, err := r.ProductSetupUsecase.PromoteCandidate(ctx, cataloginputport.PromoteProductSetupCandidateCmd{
		StoreID:            storeID,
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
func (r *mutationResolver) UpdateProductSetupCandidateStatus(
	ctx context.Context,
	id string,
	status string,
) (*model.ProductSetupCandidate, error) {
	storeID, err := requiredStoreID(ctx)
	if err != nil {
		return nil, err
	}

	candidate, err := r.ProductSetupUsecase.UpdateCandidateStatus(ctx, storeID, id, status)
	if err != nil {
		return nil, err
	}
	return toGraphQLProductSetupCandidate(*candidate), nil
}

// ProductSetupSnapshot is the resolver for the productSetupSnapshot field.
func (r *queryResolver) ProductSetupSnapshot(ctx context.Context) (*model.ProductSetupSnapshot, error) {
	storeID, err := requiredStoreID(ctx)
	if err != nil {
		return nil, err
	}

	snapshot, err := r.ProductSetupUsecase.GetSnapshot(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return toGraphQLProductSetupSnapshot(*snapshot), nil
}
