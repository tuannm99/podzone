package routing

import (
	"fmt"
	"strings"
	"time"
)

type RoutedOrderSnapshot struct {
	ID                     string
	StoreID                string
	CandidateID            string
	ProductTitle           string
	Partner                string
	Quantity               int
	Total                  string
	CustomerName           string
	Status                 string
	Timeline               []string
	ActivityLog            []RoutedOrderActivity
	ExceptionType          string
	ExceptionStatus        string
	ShipmentStatus         string
	ShipmentCarrier        string
	ShipmentTrackingNumber string
	ShipmentTrackingURL    string
	ShipmentNotes          string
	OperatorAssignee       string
	ShipmentSlaDueAt       *time.Time
	IssueSlaDueAt          *time.Time
	RoutingBlockCode       string
	RoutingBlockReason     string
	BaseCostSnapshot       string
	FulfillmentCost        string
	ShippingCost           string
	IssueCost              string
	IssueResolution        string
	IssueNotes             string
	RealizedMargin         string
	SettlementStatus       string
	SettlementNotes        string
	ShippedAt              *time.Time
	DeliveredAt            *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func RehydrateRoutedOrder(snapshot RoutedOrderSnapshot) (*RoutedOrder, error) {
	if strings.TrimSpace(snapshot.ID) == "" {
		return nil, fmt.Errorf("routed order id is required")
	}
	if strings.TrimSpace(snapshot.StoreID) == "" {
		return nil, fmt.Errorf("routed order store id is required")
	}
	if snapshot.Quantity < 0 {
		return nil, fmt.Errorf("routed order quantity is invalid")
	}

	return &RoutedOrder{
		ID:                     snapshot.ID,
		StoreID:                snapshot.StoreID,
		CandidateID:            snapshot.CandidateID,
		ProductTitle:           snapshot.ProductTitle,
		Partner:                snapshot.Partner,
		Quantity:               snapshot.Quantity,
		Total:                  snapshot.Total,
		CustomerName:           snapshot.CustomerName,
		Status:                 snapshot.Status,
		Timeline:               append([]string(nil), snapshot.Timeline...),
		ActivityLog:            append([]RoutedOrderActivity(nil), snapshot.ActivityLog...),
		ExceptionType:          snapshot.ExceptionType,
		ExceptionStatus:        snapshot.ExceptionStatus,
		ShipmentStatus:         snapshot.ShipmentStatus,
		ShipmentCarrier:        snapshot.ShipmentCarrier,
		ShipmentTrackingNumber: snapshot.ShipmentTrackingNumber,
		ShipmentTrackingURL:    snapshot.ShipmentTrackingURL,
		ShipmentNotes:          snapshot.ShipmentNotes,
		OperatorAssignee:       snapshot.OperatorAssignee,
		ShipmentSlaDueAt:       snapshot.ShipmentSlaDueAt,
		IssueSlaDueAt:          snapshot.IssueSlaDueAt,
		RoutingBlockCode:       snapshot.RoutingBlockCode,
		RoutingBlockReason:     snapshot.RoutingBlockReason,
		BaseCostSnapshot:       snapshot.BaseCostSnapshot,
		FulfillmentCost:        snapshot.FulfillmentCost,
		ShippingCost:           snapshot.ShippingCost,
		IssueCost:              snapshot.IssueCost,
		IssueResolution:        snapshot.IssueResolution,
		IssueNotes:             snapshot.IssueNotes,
		RealizedMargin:         snapshot.RealizedMargin,
		SettlementStatus:       snapshot.SettlementStatus,
		SettlementNotes:        snapshot.SettlementNotes,
		ShippedAt:              snapshot.ShippedAt,
		DeliveredAt:            snapshot.DeliveredAt,
		CreatedAt:              snapshot.CreatedAt,
		UpdatedAt:              snapshot.UpdatedAt,
	}, nil
}
