package order

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type CustomerOrderReceived struct {
	OrderID     string
	StoreID     string
	CandidateID string
	Quantity    int
	OccurredAt  time.Time
}

func (e CustomerOrderReceived) EventType() string {
	return "CustomerOrderReceived"
}

func (e CustomerOrderReceived) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type CustomerOrderQueued struct {
	OrderID    string
	StoreID    string
	Partner    string
	OccurredAt time.Time
}

func (e CustomerOrderQueued) EventType() string {
	return "CustomerOrderQueued"
}

func (e CustomerOrderQueued) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type CustomerOrderRoutingBlocked struct {
	OrderID    string
	StoreID    string
	ReasonCode string
	Reason     string
	OccurredAt time.Time
}

func (e CustomerOrderRoutingBlocked) EventType() string {
	return "CustomerOrderRoutingBlocked"
}

func (e CustomerOrderRoutingBlocked) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type CustomerOrderShipped struct {
	OrderID    string
	StoreID    string
	Partner    string
	OccurredAt time.Time
}

func (e CustomerOrderShipped) EventType() string {
	return "CustomerOrderShipped"
}

func (e CustomerOrderShipped) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type CustomerOrderAdvanced struct {
	OrderID    string
	StoreID    string
	Status     string
	Partner    string
	OccurredAt time.Time
}

func (e CustomerOrderAdvanced) EventType() string {
	return "CustomerOrderAdvanced"
}

func (e CustomerOrderAdvanced) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type CustomerOrderQueueControlUpdated struct {
	OrderID          string
	StoreID          string
	OperatorAssignee string
	OccurredAt       time.Time
}

func (e CustomerOrderQueueControlUpdated) EventType() string {
	return "CustomerOrderQueueControlUpdated"
}

func (e CustomerOrderQueueControlUpdated) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type CustomerOrderRoutingResolved struct {
	OrderID                  string
	StoreID                  string
	Partner                  string
	PreviousPartner          string
	PreviousRoutingBlockCode string
	OccurredAt               time.Time
}

func (e CustomerOrderRoutingResolved) EventType() string {
	return "CustomerOrderRoutingResolved"
}

func (e CustomerOrderRoutingResolved) OccurredAtTime() time.Time {
	return e.OccurredAt
}
