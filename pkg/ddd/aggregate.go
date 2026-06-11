package ddd

import "fmt"

// Version is the aggregate version used by repositories for optimistic locking
// or by event stores as the current stream version.
type Version uint64

// AggregateRoot is the minimal contract shared by persisted aggregate roots.
// EntityID is intentionally not required here: in this toolkit the aggregate ID
// is the root entity ID, and aggregate repositories should depend on this
// aggregate-specific name to avoid mixing child entities with aggregate roots.
type AggregateRoot interface {
	AggregateID() ID
	AggregateVersion() Version
	PullEvents() []DomainEvent
}

// AggregateBase stores common aggregate plumbing. It wraps EntityBase instead
// of embedding it so domain structs do not expose base methods accidentally.
type AggregateBase struct {
	entity  EntityBase
	version Version
	events  EventRecorder
}

var _ AggregateRoot = (*AggregateBase)(nil)

func NewAggregateBase(id ID, version Version) (AggregateBase, error) {
	entity, err := NewEntityBase(id)
	if err != nil {
		return AggregateBase{}, fmt.Errorf("aggregate id is required: %w", err)
	}
	return AggregateBase{entity: entity, version: version}, nil
}

func (a *AggregateBase) EntityID() ID {
	if a == nil {
		return ""
	}
	return a.entity.EntityID()
}

func (a *AggregateBase) AggregateID() ID {
	return a.EntityID()
}

func (a *AggregateBase) AggregateVersion() Version {
	if a == nil {
		return 0
	}
	return a.version
}

func (a *AggregateBase) SetAggregateVersion(version Version) {
	if a == nil {
		return
	}
	a.version = version
}

func (a *AggregateBase) RecordEvent(event DomainEvent) {
	if a == nil {
		return
	}
	a.events.Record(event)
}

// RecordAndApply applies a domain event to an aggregate before recording it.
// This is optional plumbing for aggregates that want event-sourced style state
// transitions; regular state mutation plus RecordEvent is still supported.
func (a *AggregateBase) RecordAndApply(applier EventApplier, event DomainEvent) error {
	if a == nil || event == nil {
		return nil
	}
	if applier == nil {
		return ErrMissingEventApplier
	}
	if err := applier.Apply(event); err != nil {
		return fmt.Errorf("apply domain event %s: %w", event.EventType(), err)
	}
	a.RecordEvent(event)
	return nil
}

func (a *AggregateBase) PullEvents() []DomainEvent {
	if a == nil {
		return nil
	}
	return a.events.Pull()
}
