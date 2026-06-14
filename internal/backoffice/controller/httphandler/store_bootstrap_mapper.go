package httphandler

import (
	"time"

	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
)

type BootstrapStoreRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
	StoreID     string `json:"store_id"     binding:"required"`
	Name        string `json:"name"         binding:"required"`
	OwnerID     string `json:"owner_id"     binding:"required"`
}

type BootstrapStoreResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r BootstrapStoreRequest) toCommand() storectx.BootstrapStoreCmd {
	return storectx.BootstrapStoreCmd{
		ID:      r.StoreID,
		Name:    r.Name,
		OwnerID: r.OwnerID,
	}
}

func newBootstrapStoreResponse(store *storectx.Store) *BootstrapStoreResponse {
	if store == nil {
		return nil
	}
	return &BootstrapStoreResponse{
		ID:        store.ID,
		Name:      store.Name,
		OwnerID:   store.OwnerID,
		Status:    store.Status,
		CreatedAt: store.CreatedAt,
		UpdatedAt: store.UpdatedAt,
	}
}
