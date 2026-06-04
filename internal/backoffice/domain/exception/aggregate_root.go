package exception

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/backoffice/domain/shared"
)

const (
	StatusOpen      = "open"
	StatusEscalated = "escalated"
	StatusResolved  = "resolved"
)

type ActivityDetail struct {
	Key   string
	Value string
}

type Change struct {
	Message string
	Details []ActivityDetail
}

type OrderExceptionSnapshot struct {
	OrderID string
	Type    string
	Status  string
}

type OrderException struct {
	orderID       string
	typ           string
	status        string
	pendingEvents []DomainEvent
}

var _ shared.AggregateRoot = (*OrderException)(nil)

func RehydrateOrderException(snapshot OrderExceptionSnapshot) (*OrderException, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	status := normalizeStatus(snapshot.Status)
	if strings.TrimSpace(snapshot.Status) != "" && status == "" {
		return nil, fmt.Errorf("invalid exception status")
	}
	return &OrderException{orderID: snapshot.OrderID, typ: snapshot.Type, status: status}, nil
}

func (e *OrderException) AggregateID() string {
	return e.orderID
}

func (e *OrderException) Snapshot() OrderExceptionSnapshot {
	return OrderExceptionSnapshot{OrderID: e.orderID, Type: e.typ, Status: e.status}
}

func (e *OrderException) PullEvents() []DomainEvent {
	events := e.pendingEvents
	e.pendingEvents = nil
	return events
}

func (e *OrderException) Open(exceptionType string, now time.Time) (Change, error) {
	exceptionType = normalizeType(exceptionType)
	if exceptionType == "" {
		return Change{}, fmt.Errorf("invalid exception type")
	}
	if e.status == StatusOpen {
		return Change{}, nil
	}
	e.typ = exceptionType
	e.status = StatusOpen
	e.record(OrderExceptionOpened{
		OrderID:       e.orderID,
		ExceptionType: e.typ,
		OccurredAt:    now.UTC(),
	})
	return Change{
		Message: fmt.Sprintf("Exception opened: %s", strings.ReplaceAll(exceptionType, "_", " ")),
		Details: details("exception_type", exceptionType, "exception_status", StatusOpen),
	}, nil
}

func (e *OrderException) UpdateStatus(status string, now time.Time) (Change, error) {
	if e.typ == "" {
		return Change{}, fmt.Errorf("routed order has no active exception type")
	}
	status = normalizeStatus(status)
	if status == "" {
		return Change{}, fmt.Errorf("invalid exception status")
	}
	if e.status == StatusResolved && status != StatusResolved {
		return Change{}, fmt.Errorf("resolved exception cannot be reopened by status update")
	}
	e.status = status
	e.record(OrderExceptionStatusChanged{
		OrderID:       e.orderID,
		ExceptionType: e.typ,
		Status:        e.status,
		OccurredAt:    now.UTC(),
	})
	return Change{
		Message: fmt.Sprintf("Exception %s: %s", status, strings.ReplaceAll(e.typ, "_", " ")),
		Details: details("exception_type", e.typ, "exception_status", status),
	}, nil
}

func (e *OrderException) record(event DomainEvent) {
	e.pendingEvents = append(e.pendingEvents, event)
}

func normalizeType(raw string) string {
	switch strings.TrimSpace(raw) {
	case "artwork_issue", "partner_delay", "address_hold", "reprint_request":
		return raw
	default:
		return ""
	}
}

func normalizeStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case StatusOpen, StatusEscalated, StatusResolved:
		return raw
	default:
		return ""
	}
}

func details(pairs ...string) []ActivityDetail {
	out := make([]ActivityDetail, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key := strings.TrimSpace(pairs[i])
		value := strings.TrimSpace(pairs[i+1])
		if key == "" || value == "" {
			continue
		}
		out = append(out, ActivityDetail{Key: key, Value: value})
	}
	return out
}
