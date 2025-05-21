package services

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/services/storeportal/domain/repositories"
)

// Module exports all service components
var Module = fx.Options(
	fx.Provide(
		NewStoreService,
	),
)

// StoreServiceParams contains dependencies for StoreService
type StoreServiceParams struct {
	fx.In

	StoreRepo repositories.StoreRepository `name:"store-repository"`
}

// NewStoreService creates a new store service
func NewStoreService(params StoreServiceParams) *StoreService {
	return &StoreService{
		storeRepo: params.StoreRepo,
	}
}
