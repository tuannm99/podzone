package resolver

import (
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	cataloginputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/inputport"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
	storeinputport "github.com/tuannm99/podzone/internal/backoffice/domain/store/inputport"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	StoreUsecase        storeinputport.StoreUsecase
	ProductSetupUsecase cataloginputport.ProductSetupUsecase
	OrderRoutingUsecase routinginputport.OrderRoutingUsecase
}

func NewResolver(
	storeUC storeinputport.StoreUsecase,
	productSetupUC cataloginputport.ProductSetupUsecase,
	orderRoutingUC routinginputport.OrderRoutingUsecase,
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
