package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Store represents a store in the system
type Store struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name"          json:"name"`
	Description string             `bson:"description"   json:"description"`
	OwnerID     string             `bson:"owner_id"      json:"ownerId"`
	Status      string             `bson:"status"        json:"status"`
	CreatedAt   time.Time          `bson:"created_at"    json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at"    json:"updatedAt"`
}

// NewStore creates a new store
func NewStore(name, description string) *Store {
	now := time.Now()
	return &Store{
		Name:        name,
		Description: description,
		Status:      "inactive",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Activate activates the store
func (s *Store) Activate() {
	s.Status = "active"
	s.UpdatedAt = time.Now()
}

// Deactivate deactivates the store
func (s *Store) Deactivate() {
	s.Status = "inactive"
	s.UpdatedAt = time.Now()
}
