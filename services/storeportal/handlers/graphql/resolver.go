package graphql

import (
	"github.com/tuannm99/podzone/services/storeportal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the resolver root.
type Resolver struct {
	storeService *service.StoreService
}

func NewResolver(storeService *service.StoreService) *Resolver {
	return &Resolver{
		storeService: storeService,
	}
}
