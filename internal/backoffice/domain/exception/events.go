package exception

import (
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type DomainEvent = ddd.DomainEvent

type OrderExceptionOpened struct {
	OrderID       string
	ExceptionType string
	OccurredAt    time.Time
}

var _ DomainEvent = OrderExceptionOpened{}

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

var _ DomainEvent = OrderExceptionStatusChanged{}

func (e OrderExceptionStatusChanged) EventType() string {
	return "OrderExceptionStatusChanged"
}

func (e OrderExceptionStatusChanged) OccurredAtTime() time.Time {
	return e.OccurredAt
}
