package entity

import (
	"fmt"
	"strings"
	"time"

	sharedentity "github.com/tuannm99/podzone/internal/backoffice/domain/shared/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/shared/valueobject"
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
	total           valueobject.Money
	fulfillmentCost valueobject.Money
	shippingCost    valueobject.Money
	issueCost       valueobject.Money
	issueResolution string
	issueNotes      string
	realizedMargin  valueobject.Money
	status          string
	notes           string
	exceptionType   string
	shipmentStatus  string
}

var _ sharedentity.AggregateRoot = (*SettlementRecord)(nil)

func RehydrateSettlementRecord(snapshot SettlementRecordSnapshot) (*SettlementRecord, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, fmt.Errorf("settlement order id is required")
	}
	total := parseMoneyOrZero(snapshot.Total)
	fulfillmentCost := parseMoneyOrZero(snapshot.FulfillmentCost)
	shippingCost := parseMoneyOrZero(snapshot.ShippingCost)
	issueCost := parseMoneyOrZero(snapshot.IssueCost)
	realizedMargin := parseMoneyOrZero(snapshot.RealizedMargin)
	return &SettlementRecord{
		orderID:         snapshot.OrderID,
		total:           total,
		fulfillmentCost: fulfillmentCost,
		shippingCost:    shippingCost,
		issueCost:       issueCost,
		issueResolution: snapshot.IssueResolution,
		issueNotes:      snapshot.IssueNotes,
		realizedMargin:  realizedMargin,
		status:          snapshot.Status,
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

func (r *SettlementRecord) UpdateSettlement(
	fulfillmentCost string,
	shippingCost string,
	status string,
	notes string,
	_ time.Time,
) (Change, *Change, error) {
	normalizedFulfillmentCost, err := valueobject.ParseMoney(fulfillmentCost)
	if err != nil {
		return Change{}, nil, fmt.Errorf("invalid fulfillment cost")
	}
	normalizedShippingCost, err := valueobject.ParseMoney(shippingCost)
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
	_ time.Time,
) (Change, *Change, error) {
	if r.exceptionType == "" && r.shipmentStatus != ShipmentStatusDeliveryIssue {
		return Change{}, nil, fmt.Errorf("issue cost handling requires an active exception or delivery issue")
	}

	normalizedIssueCost, err := valueobject.ParseMoney(issueCost)
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

func MultiplyMoney(raw string, qty int) string {
	value, err := valueobject.ParseMoney(raw)
	if err != nil {
		return "TBD"
	}
	return value.MulInt(qty).Format()
}

func CalculateMargin(total, fulfillmentCost, shippingCost, issueCost string) string {
	totalValue, err := valueobject.ParseMoney(total)
	if err != nil {
		return "TBD"
	}
	fulfillmentValue, err := valueobject.ParseMoney(fulfillmentCost)
	if err != nil {
		return "TBD"
	}
	shippingValue, err := valueobject.ParseMoney(shippingCost)
	if err != nil {
		return "TBD"
	}
	issueValue, err := valueobject.ParseMoney(issueCost)
	if err != nil {
		return "TBD"
	}
	return calculateMargin(totalValue, fulfillmentValue, shippingValue, issueValue).Format()
}

func NormalizeMoney(raw string) (string, error) {
	value, err := valueobject.ParseMoney(raw)
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
	total valueobject.Money,
	fulfillmentCost valueobject.Money,
	shippingCost valueobject.Money,
	issueCost valueobject.Money,
) valueobject.Money {
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

func parseMoneyOrZero(raw string) valueobject.Money {
	value, err := valueobject.ParseMoney(raw)
	if err == nil {
		return value
	}
	zero, _ := valueobject.NewMoney("USD", 0)
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
