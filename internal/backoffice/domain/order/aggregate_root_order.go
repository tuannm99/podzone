package order

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/pkg/ddd"
)

const (
	StatusQueued         = "queued"
	StatusRoutingBlocked = "routing_blocked"
	StatusInProduction   = "in_production"
	StatusShipped        = "shipped"

	ExceptionStatusOpen      = "open"
	ExceptionStatusEscalated = "escalated"

	SettlementStatusPending = "pending"
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
	Version            ddd.Version
	StoreID            string
	CandidateID        string
	ProductTitle       string
	Quantity           int
	Total              string
	CustomerName       string
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
	aggregate          ddd.AggregateBase
	id                 string
	storeID            string
	candidateID        string
	productTitle       string
	quantity           int
	total              string
	customerName       string
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

var _ ddd.AggregateRoot = (*CustomerOrder)(nil)

func RehydrateCustomerOrder(snapshot CustomerOrderSnapshot) (*CustomerOrder, error) {
	if strings.TrimSpace(snapshot.ID) == "" {
		return nil, ErrOrderIDRequired
	}
	if strings.TrimSpace(snapshot.StoreID) == "" {
		return nil, ErrStoreIDRequired
	}
	if snapshot.Quantity < 0 {
		return nil, ErrQuantityInvalid
	}
	if err := validateStatus(snapshot.Status); err != nil {
		return nil, err
	}
	aggregate, err := newAggregate(snapshot.ID, snapshot.Version)
	if err != nil {
		return nil, err
	}
	return &CustomerOrder{
		aggregate:          aggregate,
		id:                 snapshot.ID,
		storeID:            snapshot.StoreID,
		candidateID:        snapshot.CandidateID,
		productTitle:       snapshot.ProductTitle,
		quantity:           snapshot.Quantity,
		total:              snapshot.Total,
		customerName:       snapshot.CustomerName,
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

type ReceiveCustomerOrderInput struct {
	ID                 string
	StoreID            string
	CandidateID        string
	ProductTitle       string
	Quantity           int
	Total              string
	CustomerName       string
	Partner            string
	RoutingBlockCode   string
	RoutingBlockReason string
	Now                time.Time
}

func ReceiveCustomerOrder(input ReceiveCustomerOrderInput) (*CustomerOrder, []Change, error) {
	id := strings.TrimSpace(input.ID)
	if id == "" {
		return nil, nil, ErrOrderIDRequired
	}
	storeID := strings.TrimSpace(input.StoreID)
	if storeID == "" {
		return nil, nil, ErrStoreIDRequired
	}
	candidateID := strings.TrimSpace(input.CandidateID)
	if candidateID == "" {
		return nil, nil, ErrCandidateIDRequired
	}
	productTitle := strings.TrimSpace(input.ProductTitle)
	if productTitle == "" {
		return nil, nil, ErrProductTitleRequired
	}
	quantity := input.Quantity
	if quantity < 1 {
		quantity = 1
	}
	customerName := strings.TrimSpace(input.CustomerName)
	if customerName == "" {
		customerName = "Sample customer"
	}
	now := input.Now.UTC()
	if now.IsZero() {
		return nil, nil, ddd.NewDomainError("ORDER_TIME_REQUIRED", "customer order time is required")
	}

	partner := strings.TrimSpace(input.Partner)
	routingBlockCode := strings.TrimSpace(input.RoutingBlockCode)
	routingBlockReason := strings.TrimSpace(input.RoutingBlockReason)
	status := StatusQueued
	if partner == "" {
		if routingBlockReason == "" {
			return nil, nil, ErrRoutingReasonRequired
		}
		status = StatusRoutingBlocked
	}

	aggregate, err := newAggregate(id, 0)
	if err != nil {
		return nil, nil, err
	}
	order := &CustomerOrder{
		aggregate:          aggregate,
		id:                 id,
		storeID:            storeID,
		candidateID:        candidateID,
		productTitle:       productTitle,
		quantity:           quantity,
		total:              strings.TrimSpace(input.Total),
		customerName:       customerName,
		status:             status,
		partner:            partner,
		operatorAssignee:   "unassigned",
		routingBlockCode:   routingBlockCode,
		routingBlockReason: routingBlockReason,
		settlementStatus:   SettlementStatusPending,
		updatedAt:          now,
	}
	order.record(CustomerOrderReceived{
		OrderID:     order.id,
		StoreID:     order.storeID,
		CandidateID: order.candidateID,
		Quantity:    order.quantity,
		OccurredAt:  now,
	})

	changes := []Change{{
		Message: fmt.Sprintf("Order created for %s", order.productTitle),
		Details: details(
			"candidate_id", order.candidateID,
			"quantity", fmt.Sprintf("%d", order.quantity),
			"status", order.status,
		),
	}}
	if status == StatusRoutingBlocked {
		order.record(CustomerOrderRoutingBlocked{
			OrderID:    order.id,
			StoreID:    order.storeID,
			ReasonCode: order.routingBlockCode,
			Reason:     order.routingBlockReason,
			OccurredAt: now,
		})
		changes = append(changes, Change{
			Message: fmt.Sprintf("Routing blocked: %s", order.routingBlockReason),
			Details: details(
				"status", order.status,
				"routing_block_code", order.routingBlockCode,
				"routing_block_reason", order.routingBlockReason,
			),
		})
		return order, changes, nil
	}

	order.record(CustomerOrderQueued{
		OrderID:    order.id,
		StoreID:    order.storeID,
		Partner:    order.partner,
		OccurredAt: now,
	})
	changes = append(changes, Change{
		Message: routingTimelineEntry(order.status, order.partner),
		Details: details("status", order.status, "partner", order.partner),
	})
	return order, changes, nil
}

func (o *CustomerOrder) AggregateID() ddd.ID {
	if o == nil {
		return ""
	}
	return o.aggregate.AggregateID()
}

func (o *CustomerOrder) AggregateVersion() ddd.Version {
	if o == nil {
		return 0
	}
	return o.aggregate.AggregateVersion()
}

func (o *CustomerOrder) Snapshot() CustomerOrderSnapshot {
	return CustomerOrderSnapshot{
		ID:                 o.id,
		Version:            o.aggregate.AggregateVersion(),
		StoreID:            o.storeID,
		CandidateID:        o.candidateID,
		ProductTitle:       o.productTitle,
		Quantity:           o.quantity,
		Total:              o.total,
		CustomerName:       o.customerName,
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

func (o *CustomerOrder) PullEvents() []DomainEvent {
	if o == nil {
		return nil
	}
	return o.aggregate.PullEvents()
}

func (o *CustomerOrder) Advance(now time.Time) (Change, error) {
	if o.exceptionStatus == ExceptionStatusOpen || o.exceptionStatus == ExceptionStatusEscalated {
		return Change{}, ErrActiveException
	}

	var nextStatus string
	switch o.status {
	case StatusRoutingBlocked:
		return Change{}, ErrRoutingBlocked
	case StatusQueued:
		nextStatus = StatusInProduction
	case StatusInProduction:
		nextStatus = StatusShipped
	case StatusShipped:
		return Change{}, nil
	default:
		return Change{}, ErrStatusInvalid
	}

	o.status = nextStatus
	o.updatedAt = now
	o.record(CustomerOrderAdvanced{
		OrderID:    o.id,
		StoreID:    o.storeID,
		Status:     o.status,
		Partner:    o.partner,
		OccurredAt: now.UTC(),
	})
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
	o.record(CustomerOrderShipped{
		OrderID:    o.id,
		StoreID:    o.storeID,
		Partner:    o.partner,
		OccurredAt: now.UTC(),
	})
	return Change{
		Message: routingTimelineEntry(StatusShipped, o.partner),
		Details: details("status", StatusShipped, "partner", o.partner),
	}, true
}

func (o *CustomerOrder) record(event DomainEvent) {
	o.aggregate.RecordEvent(event)
}

func newAggregate(rawID string, version ddd.Version) (ddd.AggregateBase, error) {
	id, err := ddd.ParseID(rawID)
	if err != nil {
		return ddd.AggregateBase{}, err
	}
	return ddd.NewAggregateBase(id, version)
}

func validateStatus(status string) error {
	switch status {
	case "", StatusQueued, StatusRoutingBlocked, StatusInProduction, StatusShipped:
		return nil
	default:
		return ErrStatusInvalid
	}
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
	o.record(CustomerOrderQueueControlUpdated{
		OrderID:          o.id,
		StoreID:          o.storeID,
		OperatorAssignee: o.operatorAssignee,
		OccurredAt:       now.UTC(),
	})

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
			return Change{}, ErrSettlementStatusInvalid
		}
		o.settlementStatus = normalizedSettlementStatus
	}
	o.updatedAt = now
	o.record(CustomerOrderQueueControlUpdated{
		OrderID:          o.id,
		StoreID:          o.storeID,
		OperatorAssignee: o.operatorAssignee,
		OccurredAt:       now.UTC(),
	})

	return Change{
		Message: bulkUpdateTimelineEntry(o, assignee, shipmentSlaDueAt, settlementStatus),
		Details: details(
			"operator_assignee", o.operatorAssignee,
			"shipment_sla_due_at", formatOptionalTime(o.shipmentSlaDueAt),
			"settlement_status", o.settlementStatus,
		),
	}, nil
}

func (o *CustomerOrder) RouteManually(partner string, now time.Time) (Change, error) {
	partner = strings.TrimSpace(partner)
	if o.status != StatusRoutingBlocked {
		return Change{}, ErrNotRoutingBlocked
	}
	if partner == "" {
		return Change{}, ErrRoutingPartnerRequired
	}

	previousPartner := o.partner
	previousBlockCode := o.routingBlockCode
	previousBlockReason := o.routingBlockReason
	o.status = StatusQueued
	o.partner = partner
	o.routingBlockCode = ""
	o.routingBlockReason = ""
	o.updatedAt = now
	o.record(CustomerOrderRoutingResolved{
		OrderID:                  o.id,
		StoreID:                  o.storeID,
		Partner:                  o.partner,
		PreviousPartner:          previousPartner,
		PreviousRoutingBlockCode: previousBlockCode,
		OccurredAt:               now.UTC(),
	})

	return Change{
		Message: fmt.Sprintf("Routing unblocked: manually rerouted to %s", partner),
		Details: details(
			"status", StatusQueued,
			"previous_partner", previousPartner,
			"partner", partner,
			"previous_routing_block_code", previousBlockCode,
			"previous_routing_block_reason", previousBlockReason,
			"manual_reroute", "true",
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
