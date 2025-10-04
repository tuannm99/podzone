package model

import (
	"time"

	"github.com/google/uuid"
)

type StoreStatus string

const (
	StoreStatusDraft    StoreStatus = "draft"
	StoreStatusActive   StoreStatus = "active"
	StoreStatusInactive StoreStatus = "inactive"
)

type Store struct {
	ID          string      `db:"id"          json:"id"`
	Name        string      `db:"name"        json:"name"`
	Description string      `db:"description" json:"description"`
	OwnerID     string      `db:"owner_id"    json:"ownerId"`
	Status      StoreStatus `db:"status"      json:"status"`
	CreatedAt   time.Time   `db:"created_at"  json:"createdAt"`
	UpdatedAt   time.Time   `db:"updated_at"  json:"updatedAt"`
}

func (Store) TableName() string {
	return "stores"
}

func NewStore(name, description, ownerID string) *Store {
	now := time.Now()
	return &Store{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Status:      StoreStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (s *Store) Activate() {
	s.Status = StoreStatusActive
	s.UpdatedAt = time.Now()
}

func (s *Store) Deactivate() {
	s.Status = StoreStatusInactive
	s.UpdatedAt = time.Now()
}
