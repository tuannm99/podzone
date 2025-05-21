package graphql

import (
	"github.com/tuannm99/podzone/services/storeportal/domain/entities"
	"github.com/tuannm99/podzone/services/storeportal/interfaces/graphql/model"
)

// mapDomainStoreToGraphQL maps a domain Store entity to a GraphQL Store model
func mapDomainStoreToGraphQL(store *entities.Store) *model.Store {
	description := store.Description
	return &model.Store{
		ID:          store.ID,
		Name:        store.Name,
		Description: &description,
		Status:      model.StoreStatus(store.Status),
		CreatedAt:   store.CreatedAt,
		UpdatedAt:   store.UpdatedAt,
	}
}

// mapDomainStoresToGraphQL maps a slice of domain Store entities to GraphQL Store models
func mapDomainStoresToGraphQL(stores []*entities.Store) []*model.Store {
	result := make([]*model.Store, len(stores))
	for i, store := range stores {
		result[i] = mapDomainStoreToGraphQL(store)
	}
	return result
}
