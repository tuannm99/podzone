package settlement

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/backoffice/domain/shared"
)

const (
	StatusPending    = "pending"
	StatusReconciled = "reconciled"
	StatusPaid       = "paid"
	StatusDisputed   = "disputed"

	IssueResolutionMonitor      = "monitor"
	IssueResolutionReprint      = "reprint"
	IssueResolutionRefund       = "refund"
	IssueResolutionCarrierClaim = "carrier_claim"
	IssueResolutionAddressRetry = "address_retry"

	ShipmentStatusDeliveryIssue = "delivery_issue"
)

type ActivityDetail struct {
	Key   string
	Value string
}

type Change struct {
	Message string
	Details []ActivityDetail
}

type SettlementRecordSnapshot struct {
	OrderID         string
	Total           string
	FulfillmentCost string
	ShippingCost    string
	IssueCost       string
	IssueResolution string
	IssueNotes      string
	RealizedMargin  string
	Status          string
	Notes           string
	ExceptionType   string
	ShipmentStatus  string
}

type SettlementRecord struct {
	orderID         string
	total           shared.Money
	fulfillmentCost shared.Money
	shippingCost    shared.Money
	issueCost       shared.Money
	issueResolution string
	issueNotes      string
	realizedMargin  shared.Money
	status          string
	notes           string
	exceptionType   string
	shipmentStatus  string
	pendingEvents   []DomainEvent
}

var _ shared.AggregateRoot = (*SettlementRecord)(nil)

func RehydrateSettlementRecord(snapshot SettlementRecordSnapshot) (*SettlementRecord, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, fmt.Errorf("settlement order id is required")
	}
	total := parseMoneyOrZero(snapshot.Total)
	fulfillmentCost := parseMoneyOrZero(snapshot.FulfillmentCost)
	shippingCost := parseMoneyOrZero(snapshot.ShippingCost)
	issueCost := parseMoneyOrZero(snapshot.IssueCost)
	realizedMargin := parseMoneyOrZero(snapshot.RealizedMargin)
	status := normalizeStatus(snapshot.Status)
	if strings.TrimSpace(snapshot.Status) != "" && status == "" {
		return nil, fmt.Errorf("invalid settlement status")
	}
	return &SettlementRecord{
		orderID:         snapshot.OrderID,
		total:           total,
		fulfillmentCost: fulfillmentCost,
		shippingCost:    shippingCost,
		issueCost:       issueCost,
		issueResolution: snapshot.IssueResolution,
		issueNotes:      snapshot.IssueNotes,
		realizedMargin:  realizedMargin,
		status:          status,
		notes:           snapshot.Notes,
		exceptionType:   snapshot.ExceptionType,
		shipmentStatus:  snapshot.ShipmentStatus,
	}, nil
}

func (r *SettlementRecord) AggregateID() string {
	return r.orderID
}

func (r *SettlementRecord) Snapshot() SettlementRecordSnapshot {
	return SettlementRecordSnapshot{
		OrderID:         r.orderID,
		Total:           r.total.Format(),
		FulfillmentCost: r.fulfillmentCost.Format(),
		ShippingCost:    r.shippingCost.Format(),
		IssueCost:       r.issueCost.Format(),
		IssueResolution: r.issueResolution,
		IssueNotes:      r.issueNotes,
		RealizedMargin:  r.realizedMargin.Format(),
		Status:          r.status,
		Notes:           r.notes,
		ExceptionType:   r.exceptionType,
		ShipmentStatus:  r.shipmentStatus,
	}
}

func (r *SettlementRecord) PullEvents() []DomainEvent {
	events := r.pendingEvents
	r.pendingEvents = nil
	return events
}

func (r *SettlementRecord) UpdateSettlement(
	fulfillmentCost string,
	shippingCost string,
	status string,
	notes string,
	now time.Time,
) (Change, *Change, error) {
	normalizedFulfillmentCost, err := shared.ParseMoney(fulfillmentCost)
	if err != nil {
		return Change{}, nil, fmt.Errorf("invalid fulfillment cost")
	}
	normalizedShippingCost, err := shared.ParseMoney(shippingCost)
	if err != nil {
		return Change{}, nil, fmt.Errorf("invalid shipping cost")
	}
	normalizedStatus := normalizeStatus(status)
	if normalizedStatus == "" {
		return Change{}, nil, fmt.Errorf("invalid settlement status")
	}

	r.fulfillmentCost = normalizedFulfillmentCost
	r.shippingCost = normalizedShippingCost
	r.realizedMargin = calculateMargin(r.total, normalizedFulfillmentCost, normalizedShippingCost, r.issueCost)
	r.status = normalizedStatus
	r.notes = strings.TrimSpace(notes)
	r.record(SettlementUpdated{
		OrderID:        r.orderID,
		Status:         r.status,
		RealizedMargin: r.realizedMargin.Format(),
		OccurredAt:     now.UTC(),
	})

	systemChange := Change{
		Message: settlementTimelineEntry(r),
		Details: details(
			"settlement_status", r.status,
			"fulfillment_cost", r.fulfillmentCost.Format(),
			"shipping_cost", r.shippingCost.Format(),
			"realized_margin", r.realizedMargin.Format(),
		),
	}
	var noteChange *Change
	if r.notes != "" {
		noteChange = &Change{
			Message: r.notes,
			Details: details(
				"settlement_status", r.status,
				"realized_margin", r.realizedMargin.Format(),
			),
		}
	}
	return systemChange, noteChange, nil
}

func (r *SettlementRecord) UpdateIssueHandling(
	issueCost string,
	issueResolution string,
	notes string,
	now time.Time,
) (Change, *Change, error) {
	if r.exceptionType == "" && r.shipmentStatus != ShipmentStatusDeliveryIssue {
		return Change{}, nil, fmt.Errorf("issue cost handling requires an active exception or delivery issue")
	}

	normalizedIssueCost, err := shared.ParseMoney(issueCost)
	if err != nil {
		return Change{}, nil, fmt.Errorf("invalid issue cost")
	}
	issueResolution = normalizeIssueResolution(issueResolution)
	if issueResolution == "" {
		return Change{}, nil, fmt.Errorf("invalid issue resolution")
	}

	r.issueCost = normalizedIssueCost
	r.issueResolution = issueResolution
	r.issueNotes = strings.TrimSpace(notes)
	r.realizedMargin = calculateMargin(r.total, r.fulfillmentCost, r.shippingCost, r.issueCost)
	r.record(IssueHandlingUpdated{
		OrderID:         r.orderID,
		IssueResolution: r.issueResolution,
		IssueCost:       r.issueCost.Format(),
		RealizedMargin:  r.realizedMargin.Format(),
		OccurredAt:      now.UTC(),
	})

	systemChange := Change{
		Message: issueHandlingTimelineEntry(r),
		Details: details(
			"issue_resolution", r.issueResolution,
			"issue_cost", r.issueCost.Format(),
			"realized_margin", r.realizedMargin.Format(),
		),
	}
	var noteChange *Change
	if r.issueNotes != "" {
		noteChange = &Change{
			Message: r.issueNotes,
			Details: details("issue_resolution", r.issueResolution, "issue_cost", r.issueCost.Format()),
		}
	}
	return systemChange, noteChange, nil
}

func (r *SettlementRecord) record(event DomainEvent) {
	r.pendingEvents = append(r.pendingEvents, event)
}

func MultiplyMoney(raw string, qty int) string {
	value, err := shared.ParseMoney(raw)
	if err != nil {
		return "TBD"
	}
	return value.MulInt(qty).Format()
}

func CalculateMargin(total, fulfillmentCost, shippingCost, issueCost string) string {
	totalValue, err := shared.ParseMoney(total)
	if err != nil {
		return "TBD"
	}
	fulfillmentValue, err := shared.ParseMoney(fulfillmentCost)
	if err != nil {
		return "TBD"
	}
	shippingValue, err := shared.ParseMoney(shippingCost)
	if err != nil {
		return "TBD"
	}
	issueValue, err := shared.ParseMoney(issueCost)
	if err != nil {
		return "TBD"
	}
	return calculateMargin(totalValue, fulfillmentValue, shippingValue, issueValue).Format()
}

func NormalizeMoney(raw string) (string, error) {
	value, err := shared.ParseMoney(raw)
	if err != nil {
		return "", err
	}
	return value.Format(), nil
}

func normalizeStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case StatusPending, StatusReconciled, StatusPaid, StatusDisputed:
		return raw
	default:
		return ""
	}
}

func normalizeIssueResolution(raw string) string {
	switch strings.TrimSpace(raw) {
	case IssueResolutionMonitor,
		IssueResolutionReprint,
		IssueResolutionRefund,
		IssueResolutionCarrierClaim,
		IssueResolutionAddressRetry:
		return raw
	default:
		return ""
	}
}

func settlementTimelineEntry(record *SettlementRecord) string {
	switch record.status {
	case StatusPaid:
		return fmt.Sprintf("Settlement marked paid with realized margin %s", record.realizedMargin.Format())
	case StatusDisputed:
		return "Settlement flagged for manual dispute follow-up"
	case StatusReconciled:
		return fmt.Sprintf("Settlement reconciled with realized margin %s", record.realizedMargin.Format())
	default:
		return fmt.Sprintf("Settlement remains pending with current margin %s", record.realizedMargin.Format())
	}
}

func issueHandlingTimelineEntry(record *SettlementRecord) string {
	return fmt.Sprintf(
		"Issue handling updated: %s with impact %s",
		strings.ReplaceAll(record.issueResolution, "_", " "),
		record.issueCost.Format(),
	)
}

func calculateMargin(
	total shared.Money,
	fulfillmentCost shared.Money,
	shippingCost shared.Money,
	issueCost shared.Money,
) shared.Money {
	margin, err := total.Sub(fulfillmentCost)
	if err != nil {
		return parseMoneyOrZero("")
	}
	margin, err = margin.Sub(shippingCost)
	if err != nil {
		return parseMoneyOrZero("")
	}
	margin, err = margin.Sub(issueCost)
	if err != nil {
		return parseMoneyOrZero("")
	}
	return margin
}

func parseMoneyOrZero(raw string) shared.Money {
	value, err := shared.ParseMoney(raw)
	if err == nil {
		return value
	}
	zero, _ := shared.NewMoney("USD", 0)
	return zero
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
