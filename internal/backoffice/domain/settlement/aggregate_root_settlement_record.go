package settlement

import (
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/backoffice/domain/shared"
	"github.com/tuannm99/podzone/pkg/ddd"
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
	aggregate       ddd.AggregateBase
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
}

var _ ddd.AggregateRoot = (*SettlementRecord)(nil)

func RehydrateSettlementRecord(snapshot SettlementRecordSnapshot) (*SettlementRecord, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, ErrOrderIDRequired
	}
	total := parseMoneyOrZero(snapshot.Total)
	fulfillmentCost := parseMoneyOrZero(snapshot.FulfillmentCost)
	shippingCost := parseMoneyOrZero(snapshot.ShippingCost)
	issueCost := parseMoneyOrZero(snapshot.IssueCost)
	realizedMargin := parseMoneyOrZero(snapshot.RealizedMargin)
	status := normalizeStatus(snapshot.Status)
	if strings.TrimSpace(snapshot.Status) != "" && status == "" {
		return nil, ErrStatusInvalid
	}
	aggregate, err := newAggregate(snapshot.OrderID)
	if err != nil {
		return nil, err
	}
	return &SettlementRecord{
		aggregate:       aggregate,
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

func (r *SettlementRecord) AggregateID() ddd.ID {
	if r == nil {
		return ""
	}
	return r.aggregate.AggregateID()
}

func (r *SettlementRecord) AggregateVersion() ddd.Version {
	if r == nil {
		return 0
	}
	return r.aggregate.AggregateVersion()
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
	if r == nil {
		return nil
	}
	return r.aggregate.PullEvents()
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
		return Change{}, nil, ErrFulfillmentCostInvalid
	}
	normalizedShippingCost, err := shared.ParseMoney(shippingCost)
	if err != nil {
		return Change{}, nil, ErrShippingCostInvalid
	}
	normalizedStatus := normalizeStatus(status)
	if normalizedStatus == "" {
		return Change{}, nil, ErrStatusInvalid
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
		return Change{}, nil, ErrIssueContextRequired
	}

	normalizedIssueCost, err := shared.ParseMoney(issueCost)
	if err != nil {
		return Change{}, nil, ErrIssueCostInvalid
	}
	issueResolution = normalizeIssueResolution(issueResolution)
	if issueResolution == "" {
		return Change{}, nil, ErrIssueResolutionInvalid
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
	r.aggregate.RecordEvent(event)
}

func newAggregate(rawID string) (ddd.AggregateBase, error) {
	id, err := ddd.ParseID(rawID)
	if err != nil {
		return ddd.AggregateBase{}, err
	}
	return ddd.NewAggregateBase(id, 0)
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
