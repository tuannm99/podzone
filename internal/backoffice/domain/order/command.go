package order

import (
	"time"
)

type CreateCustomerOrderCmd struct {
	StoreID          string
	CandidateID      string
	CustomerName     string
	Quantity         int
	ProductType      string
	ShipRegion       string
	PreferredPartner string
}

type AdvanceCustomerOrderCmd struct {
	StoreID string
	OrderID string
}

type UpdateOrderQueueControlCmd struct {
	StoreID          string
	OrderID          string
	OperatorAssignee string
	ShipmentSlaDueAt *time.Time
	IssueSlaDueAt    *time.Time
}

type BulkUpdateOrdersCmd struct {
	StoreID          string
	OrderIDs         []string
	OperatorAssignee *string
	ShipmentSlaDueAt *time.Time
	SettlementStatus *string
}
