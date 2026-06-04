package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StoreStatusDraft    = "draft"
	StoreStatusActive   = "active"
	StoreStatusInactive = "inactive"
)

type Store struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	OwnerID       string    `json:"owner_id"`
	IsActive      bool      `json:"is_active"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	pendingEvents []DomainEvent
}

func NewStore(name, description, ownerID string) Store {
	store, _, err := CreateStore(name, description, ownerID, time.Now().UTC())
	if err == nil {
		return store
	}
	now := time.Now().UTC()
	return Store{
		ID:          uuid.NewString(),
		Name:        strings.TrimSpace(name),
		OwnerID:     strings.TrimSpace(ownerID),
		Description: strings.TrimSpace(description),
		Status:      StoreStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func CreateStore(name, description, ownerID string, now time.Time) (Store, []DomainEvent, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Store{}, nil, fmt.Errorf("store name is required")
	}
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return Store{}, nil, fmt.Errorf("store owner id is required")
	}
	now = now.UTC()
	store := Store{
		ID:          uuid.NewString(),
		Name:        name,
		OwnerID:     ownerID,
		Description: strings.TrimSpace(description),
		Status:      StoreStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	event := StoreCreated{
		StoreID:    store.ID,
		OwnerID:    store.OwnerID,
		Name:       store.Name,
		OccurredAt: now,
	}
	store.record(event)
	return store, []DomainEvent{event}, nil
}

func (s *Store) PullEvents() []DomainEvent {
	events := s.pendingEvents
	s.pendingEvents = nil
	return events
}

func (s *Store) Activate(now time.Time) {
	s.Status = StoreStatusActive
	s.IsActive = true
	s.UpdatedAt = now.UTC()
	s.record(StoreActivated{StoreID: s.ID, OccurredAt: s.UpdatedAt})
}

func (s *Store) Deactivate(now time.Time) {
	s.Status = StoreStatusInactive
	s.IsActive = false
	s.UpdatedAt = now.UTC()
	s.record(StoreDeactivated{StoreID: s.ID, OccurredAt: s.UpdatedAt})
}

func (s *Store) record(event DomainEvent) {
	s.pendingEvents = append(s.pendingEvents, event)
}
