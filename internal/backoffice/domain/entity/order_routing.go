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

	RoutedOrderIssueResolutionMonitor     = "monitor"
	RoutedOrderIssueResolutionReprint     = "reprint"
	RoutedOrderIssueResolutionRefund      = "refund"
	RoutedOrderIssueResolutionCarrierClaim = "carrier_claim"
	RoutedOrderIssueResolutionAddressRetry = "address_retry"
)

type RoutedOrder struct {
	ID                     string     `json:"id"`
	CandidateID            string     `json:"candidateId"`
	ProductTitle           string     `json:"productTitle"`
	Partner                string     `json:"partner"`
	Quantity               int        `json:"quantity"`
	Total                  string     `json:"total"`
	CustomerName           string     `json:"customerName"`
	Status                 string     `json:"status"`
	Timeline               []string   `json:"timeline"`
	ExceptionType          string     `json:"exceptionType"`
	ExceptionStatus        string     `json:"exceptionStatus"`
	ShipmentStatus         string     `json:"shipmentStatus"`
	ShipmentCarrier        string     `json:"shipmentCarrier"`
	ShipmentTrackingNumber string     `json:"shipmentTrackingNumber"`
	ShipmentTrackingURL    string     `json:"shipmentTrackingUrl"`
	ShipmentNotes          string     `json:"shipmentNotes"`
	BaseCostSnapshot       string     `json:"baseCostSnapshot"`
	FulfillmentCost        string     `json:"fulfillmentCost"`
	ShippingCost           string     `json:"shippingCost"`
	IssueCost              string     `json:"issueCost"`
	IssueResolution        string     `json:"issueResolution"`
	IssueNotes             string     `json:"issueNotes"`
	RealizedMargin         string     `json:"realizedMargin"`
	SettlementStatus       string     `json:"settlementStatus"`
	SettlementNotes        string     `json:"settlementNotes"`
	ShippedAt              *time.Time `json:"shippedAt,omitempty"`
	DeliveredAt            *time.Time `json:"deliveredAt,omitempty"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}
