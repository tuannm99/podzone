package entities

import (
	"time"
)

// Store represents a store entity in the domain
type Store struct {
	ID          string
	Name        string
	Description string
	Status      StoreStatus
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StoreStatus represents the possible states of a store
type StoreStatus string

const (
	StoreStatusActive   StoreStatus = "active"
	StoreStatusInactive StoreStatus = "inactive"
	StoreStatusPending  StoreStatus = "pending"
)

// NewStore creates a new store instance
func NewStore(name, description string) *Store {
	now := time.Now()
	return &Store{
		Name:        name,
		Description: description,
		Status:      StoreStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Activate activates the store
func (s *Store) Activate() {
	s.Status = StoreStatusActive
	s.UpdatedAt = time.Now()
}

// Deactivate deactivates the store
func (s *Store) Deactivate() {
	s.Status = StoreStatusInactive
	s.UpdatedAt = time.Now()
}
