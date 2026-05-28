package inputport

import (
	"context"
	"time"
)

type StoreStatus string

const (
	StoreStatusDraft     StoreStatus = "draft"
	StoreStatusPending   StoreStatus = "pending"
	StoreStatusActive    StoreStatus = "active"
	StoreStatusRejected  StoreStatus = "rejected"
	StoreStatusSuspended StoreStatus = "suspended"
)

type Store struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Subdomain   string      `json:"subdomain"`
	OwnerID     string      `json:"owner_id"`
	Status      StoreStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

type Usecase interface {
	CreateStore(ctx context.Context, name, subdomain, ownerID string) (*Store, error)
	GetStore(ctx context.Context, id string) (*Store, error)
	GetStoresByOwner(ctx context.Context, ownerID string) ([]*Store, error)
	UpdateStoreStatus(ctx context.Context, id string, status StoreStatus) error
}
