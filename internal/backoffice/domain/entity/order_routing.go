package entity

import "time"

const (
	RoutedOrderStatusQueued       = "queued"
	RoutedOrderStatusInProduction = "in_production"
	RoutedOrderStatusShipped      = "shipped"

	RoutedOrderExceptionStatusOpen      = "open"
	RoutedOrderExceptionStatusEscalated = "escalated"
	RoutedOrderExceptionStatusResolved  = "resolved"

	RoutedOrderShipmentStatusAwaitingLabel = "awaiting_label"
	RoutedOrderShipmentStatusLabelReady    = "label_ready"
	RoutedOrderShipmentStatusInTransit     = "in_transit"
	RoutedOrderShipmentStatusDelivered     = "delivered"
	RoutedOrderShipmentStatusDeliveryIssue = "delivery_issue"

	RoutedOrderSettlementStatusPending    = "pending"
	RoutedOrderSettlementStatusReconciled = "reconciled"
	RoutedOrderSettlementStatusPaid       = "paid"
	RoutedOrderSettlementStatusDisputed   = "disputed"

	RoutedOrderIssueResolutionMonitor      = "monitor"
	RoutedOrderIssueResolutionReprint      = "reprint"
	RoutedOrderIssueResolutionRefund       = "refund"
	RoutedOrderIssueResolutionCarrierClaim = "carrier_claim"
	RoutedOrderIssueResolutionAddressRetry = "address_retry"

	RoutedOrderActivityTypeSystem         = "system"
	RoutedOrderActivityTypeShipmentNote   = "shipment_note"
	RoutedOrderActivityTypeSettlementNote = "settlement_note"
	RoutedOrderActivityTypeIssueNote      = "issue_note"
)

type RoutedOrderActivity struct {
	Type      string                      `json:"type"`
	Actor     string                      `json:"actor"`
	Message   string                      `json:"message"`
	Details   []RoutedOrderActivityDetail `json:"details"`
	CreatedAt time.Time                   `json:"createdAt"`
}

type RoutedOrderActivityDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RoutedOrderActivityFeedEntry struct {
	OrderID          string              `json:"orderId"`
	ProductTitle     string              `json:"productTitle"`
	Partner          string              `json:"partner"`
	OperatorAssignee string              `json:"operatorAssignee"`
	Activity         RoutedOrderActivity `json:"activity"`
}

type RoutedOrderActivityFeedPage struct {
	Entries    []RoutedOrderActivityFeedEntry `json:"entries"`
	Total      int                            `json:"total"`
	NextCursor *string                        `json:"nextCursor,omitempty"`
}

type RoutedOrder struct {
	ID                     string                `json:"id"`
	CandidateID            string                `json:"candidateId"`
	ProductTitle           string                `json:"productTitle"`
	Partner                string                `json:"partner"`
	Quantity               int                   `json:"quantity"`
	Total                  string                `json:"total"`
	CustomerName           string                `json:"customerName"`
	Status                 string                `json:"status"`
	Timeline               []string              `json:"timeline"`
	ActivityLog            []RoutedOrderActivity `json:"activityLog"`
	ExceptionType          string                `json:"exceptionType"`
	ExceptionStatus        string                `json:"exceptionStatus"`
	ShipmentStatus         string                `json:"shipmentStatus"`
	ShipmentCarrier        string                `json:"shipmentCarrier"`
	ShipmentTrackingNumber string                `json:"shipmentTrackingNumber"`
	ShipmentTrackingURL    string                `json:"shipmentTrackingUrl"`
	ShipmentNotes          string                `json:"shipmentNotes"`
	OperatorAssignee       string                `json:"operatorAssignee"`
	ShipmentSlaDueAt       *time.Time            `json:"shipmentSlaDueAt,omitempty"`
	IssueSlaDueAt          *time.Time            `json:"issueSlaDueAt,omitempty"`
	BaseCostSnapshot       string                `json:"baseCostSnapshot"`
	FulfillmentCost        string                `json:"fulfillmentCost"`
	ShippingCost           string                `json:"shippingCost"`
	IssueCost              string                `json:"issueCost"`
	IssueResolution        string                `json:"issueResolution"`
	IssueNotes             string                `json:"issueNotes"`
	RealizedMargin         string                `json:"realizedMargin"`
	SettlementStatus       string                `json:"settlementStatus"`
	SettlementNotes        string                `json:"settlementNotes"`
	ShippedAt              *time.Time            `json:"shippedAt,omitempty"`
	DeliveredAt            *time.Time            `json:"deliveredAt,omitempty"`
	CreatedAt              time.Time             `json:"createdAt"`
	UpdatedAt              time.Time             `json:"updatedAt"`
}
