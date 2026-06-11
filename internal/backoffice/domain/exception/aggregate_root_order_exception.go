package exception

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
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
	aggregate ddd.AggregateBase
	orderID   string
	typ       string
	status    string
}

var _ ddd.AggregateRoot = (*OrderException)(nil)

func RehydrateOrderException(snapshot OrderExceptionSnapshot) (*OrderException, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, ErrOrderIDRequired
	}
	status := normalizeStatus(snapshot.Status)
	if strings.TrimSpace(snapshot.Status) != "" && status == "" {
		return nil, ErrStatusInvalid
	}
	aggregate, err := newAggregate(snapshot.OrderID)
	if err != nil {
		return nil, err
	}
	return &OrderException{aggregate: aggregate, orderID: snapshot.OrderID, typ: snapshot.Type, status: status}, nil
}

func (e *OrderException) AggregateID() ddd.ID {
	if e == nil {
		return ""
	}
	return e.aggregate.AggregateID()
}

func (e *OrderException) AggregateVersion() ddd.Version {
	if e == nil {
		return 0
	}
	return e.aggregate.AggregateVersion()
}

func (e *OrderException) Snapshot() OrderExceptionSnapshot {
	return OrderExceptionSnapshot{OrderID: e.orderID, Type: e.typ, Status: e.status}
}

func (e *OrderException) PullEvents() []DomainEvent {
	if e == nil {
		return nil
	}
	return e.aggregate.PullEvents()
}

func (e *OrderException) Open(exceptionType string, now time.Time) (Change, error) {
	exceptionType = normalizeType(exceptionType)
	if exceptionType == "" {
		return Change{}, ErrTypeInvalid
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
		return Change{}, ErrNoActiveType
	}
	status = normalizeStatus(status)
	if status == "" {
		return Change{}, ErrStatusInvalid
	}
	if e.status == StatusResolved && status != StatusResolved {
		return Change{}, ErrResolvedCannotReopen
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
	e.aggregate.RecordEvent(event)
}

func newAggregate(rawID string) (ddd.AggregateBase, error) {
	id, err := ddd.ParseID(rawID)
	if err != nil {
		return ddd.AggregateBase{}, err
	}
	return ddd.NewAggregateBase(id, 0)
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
