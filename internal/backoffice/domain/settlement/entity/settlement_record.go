package entity

import (
	"fmt"
	"strings"
	"time"
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
	total           string
	fulfillmentCost string
	shippingCost    string
	issueCost       string
	issueResolution string
	issueNotes      string
	realizedMargin  string
	status          string
	notes           string
	exceptionType   string
	shipmentStatus  string
}

func RehydrateSettlementRecord(snapshot SettlementRecordSnapshot) (*SettlementRecord, error) {
	if strings.TrimSpace(snapshot.OrderID) == "" {
		return nil, fmt.Errorf("settlement order id is required")
	}
	return &SettlementRecord{
		orderID:         snapshot.OrderID,
		total:           snapshot.Total,
		fulfillmentCost: snapshot.FulfillmentCost,
		shippingCost:    snapshot.ShippingCost,
		issueCost:       snapshot.IssueCost,
		issueResolution: snapshot.IssueResolution,
		issueNotes:      snapshot.IssueNotes,
		realizedMargin:  snapshot.RealizedMargin,
		status:          snapshot.Status,
		notes:           snapshot.Notes,
		exceptionType:   snapshot.ExceptionType,
		shipmentStatus:  snapshot.ShipmentStatus,
	}, nil
}

func (r *SettlementRecord) Snapshot() SettlementRecordSnapshot {
	return SettlementRecordSnapshot{
		OrderID:         r.orderID,
		Total:           r.total,
		FulfillmentCost: r.fulfillmentCost,
		ShippingCost:    r.shippingCost,
		IssueCost:       r.issueCost,
		IssueResolution: r.issueResolution,
		IssueNotes:      r.issueNotes,
		RealizedMargin:  r.realizedMargin,
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
	normalizedFulfillmentCost, err := NormalizeMoney(fulfillmentCost)
	if err != nil {
		return Change{}, nil, fmt.Errorf("invalid fulfillment cost")
	}
	normalizedShippingCost, err := NormalizeMoney(shippingCost)
	if err != nil {
		return Change{}, nil, fmt.Errorf("invalid shipping cost")
	}
	normalizedStatus := normalizeStatus(status)
	if normalizedStatus == "" {
		return Change{}, nil, fmt.Errorf("invalid settlement status")
	}

	r.fulfillmentCost = normalizedFulfillmentCost
	r.shippingCost = normalizedShippingCost
	r.realizedMargin = CalculateMargin(r.total, normalizedFulfillmentCost, normalizedShippingCost, r.issueCost)
	r.status = normalizedStatus
	r.notes = strings.TrimSpace(notes)

	systemChange := Change{
		Message: settlementTimelineEntry(r),
		Details: details(
			"settlement_status", r.status,
			"fulfillment_cost", r.fulfillmentCost,
			"shipping_cost", r.shippingCost,
			"realized_margin", r.realizedMargin,
		),
	}
	var noteChange *Change
	if r.notes != "" {
		noteChange = &Change{
			Message: r.notes,
			Details: details(
				"settlement_status", r.status,
				"realized_margin", r.realizedMargin,
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

	normalizedIssueCost, err := NormalizeMoney(issueCost)
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
	r.realizedMargin = CalculateMargin(r.total, r.fulfillmentCost, r.shippingCost, r.issueCost)

	systemChange := Change{
		Message: issueHandlingTimelineEntry(r),
		Details: details(
			"issue_resolution", r.issueResolution,
			"issue_cost", r.issueCost,
			"realized_margin", r.realizedMargin,
		),
	}
	var noteChange *Change
	if r.issueNotes != "" {
		noteChange = &Change{
			Message: r.issueNotes,
			Details: details("issue_resolution", r.issueResolution, "issue_cost", r.issueCost),
		}
	}
	return systemChange, noteChange, nil
}

func MultiplyMoney(raw string, qty int) string {
	value, ok := parseMoney(raw)
	if !ok {
		return "TBD"
	}
	return formatMoney(value * float64(qty))
}

func CalculateMargin(total, fulfillmentCost, shippingCost, issueCost string) string {
	totalValue, ok := parseMoney(total)
	if !ok {
		return "TBD"
	}
	fulfillmentValue, ok := parseMoney(fulfillmentCost)
	if !ok {
		return "TBD"
	}
	shippingValue, ok := parseMoney(shippingCost)
	if !ok {
		return "TBD"
	}
	issueValue, ok := parseMoney(issueCost)
	if !ok {
		return "TBD"
	}
	return formatMoney(totalValue - fulfillmentValue - shippingValue - issueValue)
}

func NormalizeMoney(raw string) (string, error) {
	value, ok := parseMoney(raw)
	if !ok {
		return "", fmt.Errorf("invalid money")
	}
	return formatMoney(value), nil
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
		return fmt.Sprintf("Settlement marked paid with realized margin %s", record.realizedMargin)
	case StatusDisputed:
		return "Settlement flagged for manual dispute follow-up"
	case StatusReconciled:
		return fmt.Sprintf("Settlement reconciled with realized margin %s", record.realizedMargin)
	default:
		return fmt.Sprintf("Settlement remains pending with current margin %s", record.realizedMargin)
	}
}

func issueHandlingTimelineEntry(record *SettlementRecord) string {
	return fmt.Sprintf(
		"Issue handling updated: %s with impact %s",
		strings.ReplaceAll(record.issueResolution, "_", " "),
		record.issueCost,
	)
}

func parseMoney(raw string) (float64, bool) {
	negative := strings.Contains(raw, "-")
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' {
			return r
		}
		return -1
	}, raw)
	if cleaned == "" {
		return 0, false
	}
	var value float64
	if _, err := fmt.Sscanf(cleaned, "%f", &value); err != nil {
		return 0, false
	}
	if negative {
		value = -value
	}
	return value, true
}

func formatMoney(value float64) string {
	return fmt.Sprintf("$%.2f", value)
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
