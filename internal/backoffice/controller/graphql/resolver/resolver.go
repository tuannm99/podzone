package resolver

import "github.com/tuannm99/podzone/internal/backoffice/domain/inputport"

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
