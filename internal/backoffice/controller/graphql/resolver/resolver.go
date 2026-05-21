package resolver

import (
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	StoreUsecase        inputport.StoreUsecase
	ProductSetupUsecase inputport.ProductSetupUsecase
	OrderRoutingUsecase inputport.OrderRoutingUsecase
}

func NewResolver(
	storeUC inputport.StoreUsecase,
	productSetupUC inputport.ProductSetupUsecase,
	orderRoutingUC inputport.OrderRoutingUsecase,
) *Resolver {
	return &Resolver{
		StoreUsecase:        storeUC,
		ProductSetupUsecase: productSetupUC,
		OrderRoutingUsecase: orderRoutingUC,
	}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
