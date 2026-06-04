package resolver

import (
	backofficeoperations "github.com/tuannm99/podzone/internal/backoffice/application/operations"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	cataloginputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	StoreUsecase        storectx.StoreUsecase
	ProductSetupUsecase cataloginputport.ProductSetupUsecase
	OrderRoutingUsecase backofficeoperations.OrderRoutingUsecase
}

func NewResolver(
	storeUC storectx.StoreUsecase,
	productSetupUC cataloginputport.ProductSetupUsecase,
	orderRoutingUC backofficeoperations.OrderRoutingUsecase,
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

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
