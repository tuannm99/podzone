package store

import (
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type DomainEvent = ddd.DomainEvent

type StoreCreated struct {
	StoreID    string
	OwnerID    string
	Name       string
	OccurredAt time.Time
}

var _ DomainEvent = StoreCreated{}

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

var _ DomainEvent = StoreActivated{}

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

var _ DomainEvent = StoreDeactivated{}

func (e StoreDeactivated) EventType() string {
	return "StoreDeactivated"
}

func (e StoreDeactivated) OccurredAtTime() time.Time {
	return e.OccurredAt
}
