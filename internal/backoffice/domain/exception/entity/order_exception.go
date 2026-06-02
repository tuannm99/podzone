package entity

import (
	"fmt"
	"strings"
	"time"
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
	orderID string
	typ     string
	status  string
}

func RehydrateOrderException(snapshot OrderExceptionSnapshot) (*OrderException, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, fmt.Errorf("order id is required")
	}
	return &OrderException{orderID: snapshot.OrderID, typ: snapshot.Type, status: snapshot.Status}, nil
}

func (e *OrderException) Snapshot() OrderExceptionSnapshot {
	return OrderExceptionSnapshot{OrderID: e.orderID, Type: e.typ, Status: e.status}
}

func (e *OrderException) Open(exceptionType string, _ time.Time) (Change, error) {
	exceptionType = normalizeType(exceptionType)
	if exceptionType == "" {
		return Change{}, fmt.Errorf("invalid exception type")
	}
	if e.status == StatusOpen {
		return Change{}, nil
	}
	e.typ = exceptionType
	e.status = StatusOpen
	return Change{
		Message: fmt.Sprintf("Exception opened: %s", strings.ReplaceAll(exceptionType, "_", " ")),
		Details: details("exception_type", exceptionType, "exception_status", StatusOpen),
	}, nil
}

func (e *OrderException) UpdateStatus(status string, _ time.Time) (Change, error) {
	if e.typ == "" {
		return Change{}, fmt.Errorf("routed order has no active exception type")
	}
	status = normalizeStatus(status)
	if status == "" {
		return Change{}, fmt.Errorf("invalid exception status")
	}
	e.status = status
	return Change{
		Message: fmt.Sprintf("Exception %s: %s", status, strings.ReplaceAll(e.typ, "_", " ")),
		Details: details("exception_type", e.typ, "exception_status", status),
	}, nil
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
