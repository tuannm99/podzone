package graphql

import (
	"github.com/tuannm99/podzone/services/storeportal/application/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	storeService *services.StoreService
}

// NewResolver creates a new resolver with required dependencies
func NewResolver(storeService *services.StoreService) *Resolver {
	return &Resolver{
		storeService: storeService,
	}
}
