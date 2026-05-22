package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	StoreStatusDraft    = "draft"
	StoreStatusActive   = "active"
	StoreStatusInactive = "inactive"
)

type Store struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OwnerID     string    `json:"owner_id"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewStore(name, description, ownerID string) Store {
	now := time.Now().UTC()
	return Store{
		ID:          uuid.NewString(),
		Name:        name,
		OwnerID:     ownerID,
		Description: description,
		Status:      StoreStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
