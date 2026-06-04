package store

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type StoreCreated struct {
	StoreID    string
	OwnerID    string
	Name       string
	OccurredAt time.Time
}

func (e StoreCreated) EventType() string {
	return "StoreCreated"
}

func (e StoreCreated) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type StoreActivated struct {
	StoreID    string
	OccurredAt time.Time
}

func (e StoreActivated) EventType() string {
	return "StoreActivated"
}

func (e StoreActivated) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type StoreDeactivated struct {
	StoreID    string
	OccurredAt time.Time
}

func (e StoreDeactivated) EventType() string {
	return "StoreDeactivated"
}

func (e StoreDeactivated) OccurredAtTime() time.Time {
	return e.OccurredAt
}
