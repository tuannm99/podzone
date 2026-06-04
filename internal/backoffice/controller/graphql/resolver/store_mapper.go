package resolver

import (
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
)

func toGraphQLStore(store storectx.Store) *model.Store {
	return &model.Store{
		ID:          store.ID,
		Name:        store.Name,
		OwnerID:     store.OwnerID,
		IsActive:    store.IsActive,
		Description: store.Description,
		Status:      store.Status,
		CreatedAt:   store.CreatedAt,
		UpdatedAt:   store.UpdatedAt,
	}
}
