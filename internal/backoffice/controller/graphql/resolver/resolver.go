package resolver

import "github.com/tuannm99/podzone/internal/backoffice/domain/inputport"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	StoreUsecase inputport.StoreUsecase
}

func NewResolver(storeUC inputport.StoreUsecase) *Resolver {
	return &Resolver{StoreUsecase: storeUC}
}
