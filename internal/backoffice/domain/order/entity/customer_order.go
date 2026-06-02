package entity

import (
	"fmt"
	"strings"
	"time"
)

const (
	StatusQueued         = "queued"
	StatusRoutingBlocked = "routing_blocked"
	StatusInProduction   = "in_production"
	StatusShipped        = "shipped"

	ExceptionStatusOpen      = "open"
	ExceptionStatusEscalated = "escalated"
)

type ActivityDetail struct {
	Key   string
	Value string
}

type Change struct {
	Message string
	Details []ActivityDetail
}

type CustomerOrderSnapshot struct {
	ID                 string
	StoreID            string
	Status             string
	Partner            string
	OperatorAssignee   string
	ShipmentSlaDueAt   *time.Time
	IssueSlaDueAt      *time.Time
	ExceptionStatus    string
	RoutingBlockCode   string
	RoutingBlockReason string
	SettlementStatus   string
	UpdatedAt          time.Time
}

type CustomerOrder struct {
	id                 string
	storeID            string
	status             string
	partner            string
	operatorAssignee   string
	shipmentSlaDueAt   *time.Time
	issueSlaDueAt      *time.Time
	exceptionStatus    string
	routingBlockCode   string
	routingBlockReason string
	settlementStatus   string
	updatedAt          time.Time
}

func RehydrateCustomerOrder(snapshot CustomerOrderSnapshot) (*CustomerOrder, error) {
	if strings.TrimSpace(snapshot.ID) == "" {
		return nil, fmt.Errorf("customer order id is required")
	}
	if strings.TrimSpace(snapshot.StoreID) == "" {
		return nil, fmt.Errorf("customer order store id is required")
	}
	return &CustomerOrder{
		id:                 snapshot.ID,
		storeID:            snapshot.StoreID,
		status:             snapshot.Status,
		partner:            snapshot.Partner,
		operatorAssignee:   snapshot.OperatorAssignee,
		shipmentSlaDueAt:   snapshot.ShipmentSlaDueAt,
		issueSlaDueAt:      snapshot.IssueSlaDueAt,
		exceptionStatus:    snapshot.ExceptionStatus,
		routingBlockCode:   snapshot.RoutingBlockCode,
		routingBlockReason: snapshot.RoutingBlockReason,
		settlementStatus:   snapshot.SettlementStatus,
		updatedAt:          snapshot.UpdatedAt,
	}, nil
}

func (o *CustomerOrder) Snapshot() CustomerOrderSnapshot {
	return CustomerOrderSnapshot{
		ID:                 o.id,
		StoreID:            o.storeID,
		Status:             o.status,
		Partner:            o.partner,
		OperatorAssignee:   o.operatorAssignee,
		ShipmentSlaDueAt:   o.shipmentSlaDueAt,
		IssueSlaDueAt:      o.issueSlaDueAt,
		ExceptionStatus:    o.exceptionStatus,
		RoutingBlockCode:   o.routingBlockCode,
		RoutingBlockReason: o.routingBlockReason,
		SettlementStatus:   o.settlementStatus,
		UpdatedAt:          o.updatedAt,
	}
}

func (o *CustomerOrder) Advance(now time.Time) (Change, error) {
	if o.exceptionStatus == ExceptionStatusOpen || o.exceptionStatus == ExceptionStatusEscalated {
		return Change{}, fmt.Errorf("resolve the active exception before advancing the routed order")
	}

	var nextStatus string
	switch o.status {
	case StatusRoutingBlocked:
		return Change{}, fmt.Errorf("resolve the routing block before advancing the routed order")
	case StatusQueued:
		nextStatus = StatusInProduction
	case StatusInProduction:
		nextStatus = StatusShipped
	case StatusShipped:
		return Change{}, nil
	default:
		return Change{}, fmt.Errorf("invalid routed order status")
	}

	o.status = nextStatus
	o.updatedAt = now
	return Change{
		Message: routingTimelineEntry(nextStatus, o.partner),
		Details: details("status", nextStatus, "partner", o.partner),
	}, nil
}

func (o *CustomerOrder) MarkShipped(now time.Time) (Change, bool) {
	if o.status == StatusShipped {
		return Change{}, false
	}
	o.status = StatusShipped
	o.updatedAt = now
	return Change{
		Message: routingTimelineEntry(StatusShipped, o.partner),
		Details: details("status", StatusShipped, "partner", o.partner),
	}, true
}

func (o *CustomerOrder) UpdateQueueControl(
	assignee string,
	shipmentSlaDueAt *time.Time,
	issueSlaDueAt *time.Time,
	now time.Time,
) Change {
	assignee = strings.TrimSpace(assignee)
	if assignee == "" {
		assignee = "unassigned"
	}
	o.operatorAssignee = assignee
	o.shipmentSlaDueAt = shipmentSlaDueAt
	o.issueSlaDueAt = issueSlaDueAt
	o.updatedAt = now

	return Change{
		Message: queueControlTimelineEntry(o),
		Details: details(
			"operator_assignee", o.operatorAssignee,
			"shipment_sla_due_at", formatOptionalTime(o.shipmentSlaDueAt),
			"issue_sla_due_at", formatOptionalTime(o.issueSlaDueAt),
		),
	}
}

func (o *CustomerOrder) ApplyBulkQueueControl(
	assignee *string,
	shipmentSlaDueAt *time.Time,
	settlementStatus *string,
	now time.Time,
) (Change, error) {
	if assignee != nil {
		normalizedAssignee := strings.TrimSpace(*assignee)
		if normalizedAssignee == "" {
			normalizedAssignee = "unassigned"
		}
		o.operatorAssignee = normalizedAssignee
	}
	if shipmentSlaDueAt != nil {
		o.shipmentSlaDueAt = shipmentSlaDueAt
	}
	if settlementStatus != nil {
		normalizedSettlementStatus := strings.TrimSpace(*settlementStatus)
		if normalizedSettlementStatus == "" {
			return Change{}, fmt.Errorf("invalid settlement status")
		}
		o.settlementStatus = normalizedSettlementStatus
	}
	o.updatedAt = now

	return Change{
		Message: bulkUpdateTimelineEntry(o, assignee, shipmentSlaDueAt, settlementStatus),
		Details: details(
			"operator_assignee", o.operatorAssignee,
			"shipment_sla_due_at", formatOptionalTime(o.shipmentSlaDueAt),
			"settlement_status", o.settlementStatus,
		),
	}, nil
}

func routingTimelineEntry(status, partner string) string {
	switch status {
	case StatusInProduction:
		return fmt.Sprintf("Sent to %s for POD production", partner)
	case StatusShipped:
		return fmt.Sprintf("Marked as shipped from %s", partner)
	default:
		return fmt.Sprintf("Queued for %s", partner)
	}
}

func queueControlTimelineEntry(order *CustomerOrder) string {
	shipmentDue := "none"
	if order.shipmentSlaDueAt != nil {
		shipmentDue = order.shipmentSlaDueAt.UTC().Format(time.RFC3339)
	}
	issueDue := "none"
	if order.issueSlaDueAt != nil {
		issueDue = order.issueSlaDueAt.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf(
		"Queue ownership updated: %s · shipment SLA %s · issue SLA %s",
		order.operatorAssignee,
		shipmentDue,
		issueDue,
	)
}

func bulkUpdateTimelineEntry(
	order *CustomerOrder,
	assignee *string,
	shipmentSlaDueAt *time.Time,
	settlementStatus *string,
) string {
	parts := make([]string, 0, 3)
	if assignee != nil {
		parts = append(parts, fmt.Sprintf("owner %s", order.operatorAssignee))
	}
	if shipmentSlaDueAt != nil {
		parts = append(parts, fmt.Sprintf("shipment SLA %s", shipmentSlaDueAt.UTC().Format(time.RFC3339)))
	}
	if settlementStatus != nil {
		parts = append(parts, fmt.Sprintf("settlement %s", order.settlementStatus))
	}
	return fmt.Sprintf("Bulk queue update applied: %s", strings.Join(parts, " · "))
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

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
