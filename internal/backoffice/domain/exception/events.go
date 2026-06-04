package exception

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type OrderExceptionOpened struct {
	OrderID       string
	ExceptionType string
	OccurredAt    time.Time
}

func (e OrderExceptionOpened) EventType() string {
	return "OrderExceptionOpened"
}

func (e OrderExceptionOpened) OccurredAtTime() time.Time {
	return e.OccurredAt
}

type OrderExceptionStatusChanged struct {
	OrderID       string
	ExceptionType string
	Status        string
	OccurredAt    time.Time
}

func (e OrderExceptionStatusChanged) EventType() string {
	return "OrderExceptionStatusChanged"
}

func (e OrderExceptionStatusChanged) OccurredAtTime() time.Time {
	return e.OccurredAt
}
