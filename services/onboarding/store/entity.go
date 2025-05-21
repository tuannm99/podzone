package store

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Store represents a store in the onboarding process
type Store struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"          json:"id"`
	Name        string             `bson:"name"                   json:"name"`
	Subdomain   string             `bson:"subdomain"              json:"subdomain"`
	OwnerID     string             `bson:"owner_id"               json:"owner_id"`
	Status      StoreStatus        `bson:"status"                 json:"status"`
	CreatedAt   time.Time          `bson:"created_at"             json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"             json:"updated_at"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// StoreStatus represents the current status of store onboarding
type StoreStatus string

const (
	StoreStatusDraft     StoreStatus = "draft"
	StoreStatusPending   StoreStatus = "pending"
	StoreStatusActive    StoreStatus = "active"
	StoreStatusRejected  StoreStatus = "rejected"
	StoreStatusSuspended StoreStatus = "suspended"
)
