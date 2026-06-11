package store

import (
	"strings"
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
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

type StoreAggregate struct {
	aggregate   ddd.AggregateBase
	id          string
	name        string
	ownerID     string
	isActive    bool
	description string
	status      string
	createdAt   time.Time
	updatedAt   time.Time
}

var _ ddd.AggregateRoot = (*StoreAggregate)(nil)

func CreateStore(
	id string,
	name string,
	description string,
	ownerID string,
	now time.Time,
) (*StoreAggregate, []DomainEvent, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, ErrStoreIDRequired
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil, ErrStoreNameRequired
	}
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, nil, ErrStoreOwnerRequired
	}
	now = now.UTC()
	if now.IsZero() {
		return nil, nil, ErrStoreTimeRequired
	}
	aggregate, err := newAggregate(id)
	if err != nil {
		return nil, nil, err
	}
	store := &StoreAggregate{
		aggregate:   aggregate,
		id:          id,
		name:        name,
		ownerID:     ownerID,
		description: strings.TrimSpace(description),
		status:      StoreStatusDraft,
		createdAt:   now,
		updatedAt:   now,
	}
	event := StoreCreated{
		StoreID:    store.id,
		OwnerID:    store.ownerID,
		Name:       store.name,
		OccurredAt: now,
	}
	store.record(event)
	return store, []DomainEvent{event}, nil
}

func RehydrateStore(snapshot Store) (*StoreAggregate, error) {
	aggregate, err := newAggregate(snapshot.ID)
	if err != nil {
		return nil, err
	}
	return &StoreAggregate{
		aggregate:   aggregate,
		id:          snapshot.ID,
		name:        snapshot.Name,
		ownerID:     snapshot.OwnerID,
		isActive:    snapshot.IsActive,
		description: snapshot.Description,
		status:      snapshot.Status,
		createdAt:   snapshot.CreatedAt,
		updatedAt:   snapshot.UpdatedAt,
	}, nil
}

func (s *StoreAggregate) PullEvents() []DomainEvent {
	if s == nil {
		return nil
	}
	return s.aggregate.PullEvents()
}

func (s *StoreAggregate) AggregateID() ddd.ID {
	if s == nil {
		return ""
	}
	return s.aggregate.AggregateID()
}

func (s *StoreAggregate) AggregateVersion() ddd.Version {
	if s == nil {
		return 0
	}
	return s.aggregate.AggregateVersion()
}

func (s *StoreAggregate) Snapshot() Store {
	return Store{
		ID:          s.id,
		Name:        s.name,
		OwnerID:     s.ownerID,
		IsActive:    s.isActive,
		Description: s.description,
		Status:      s.status,
		CreatedAt:   s.createdAt,
		UpdatedAt:   s.updatedAt,
	}
}

func (s *StoreAggregate) Activate(now time.Time) {
	s.status = StoreStatusActive
	s.isActive = true
	s.updatedAt = now.UTC()
	s.record(StoreActivated{StoreID: s.id, OccurredAt: s.updatedAt})
}

func (s *StoreAggregate) Deactivate(now time.Time) {
	s.status = StoreStatusInactive
	s.isActive = false
	s.updatedAt = now.UTC()
	s.record(StoreDeactivated{StoreID: s.id, OccurredAt: s.updatedAt})
}

func (s *StoreAggregate) record(event DomainEvent) {
	s.aggregate.RecordEvent(event)
}

func newAggregate(rawID string) (ddd.AggregateBase, error) {
	id, err := ddd.ParseID(rawID)
	if err != nil {
		return ddd.AggregateBase{}, err
	}
	return ddd.NewAggregateBase(id, 0)
}
