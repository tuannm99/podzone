package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StoreStatus string

const (
	StoreStatusDraft    StoreStatus = "draft"
	StoreStatusActive   StoreStatus = "active"
	StoreStatusInactive StoreStatus = "inactive"
)

// Store represents a store in the system
type Store struct {
	ID          string      `gorm:"primaryKey"                json:"id"`
	Name        string      `gorm:"not null"                  json:"name"`
	Description string      `gorm:"not null"                  json:"description"`
	OwnerID     string      `gorm:"not null;index"            json:"ownerId"`
	Status      StoreStatus `gorm:"not null;type:varchar(20)" json:"status"`
	CreatedAt   time.Time   `gorm:"not null"                  json:"createdAt"`
	UpdatedAt   time.Time   `gorm:"not null"                  json:"updatedAt"`
}

// TableName specifies the table name for the Store model
func (Store) TableName() string {
	return "stores"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (s *Store) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// NewStore creates a new store
func NewStore(name, description string) *Store {
	now := time.Now()
	return &Store{
		Name:        name,
		Description: description,
		Status:      StoreStatusDraft,
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
