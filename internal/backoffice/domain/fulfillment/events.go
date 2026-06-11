package fulfillment

import (
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type DomainEvent = ddd.DomainEvent

type ShipmentStatusUpdated struct {
	OrderID    string
	Partner    string
	Status     string
	OccurredAt time.Time
}

var _ DomainEvent = ShipmentStatusUpdated{}

func (e ShipmentStatusUpdated) EventType() string {
	return "ShipmentStatusUpdated"
}

func (e ShipmentStatusUpdated) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type ShipmentInTransit struct {
	OrderID        string
	Partner        string
	Carrier        string
	TrackingNumber string
	OccurredAt     time.Time
}

var _ DomainEvent = ShipmentInTransit{}

func (e ShipmentInTransit) EventType() string {
	return "ShipmentInTransit"
}

func (e ShipmentInTransit) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type ShipmentDelivered struct {
	OrderID        string
	Partner        string
	Carrier        string
	TrackingNumber string
	OccurredAt     time.Time
}

var _ DomainEvent = ShipmentDelivered{}

func (e ShipmentDelivered) EventType() string {
	return "ShipmentDelivered"
}

func (e ShipmentDelivered) OccurredAtTime() time.Time {
	return e.OccurredAt
}
