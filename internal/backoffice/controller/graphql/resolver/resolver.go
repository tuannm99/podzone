package resolver

import (
	backofficeoperations "github.com/tuannm99/podzone/internal/backoffice/application/operations"
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
