package fulfillment

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type ShipmentStatusUpdated struct {
	OrderID    string
	Partner    string
	Status     string
	OccurredAt time.Time
}

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

func (e ShipmentDelivered) EventType() string {
	return "ShipmentDelivered"
}

func (e ShipmentDelivered) OccurredAtTime() time.Time {
	return e.OccurredAt
}
