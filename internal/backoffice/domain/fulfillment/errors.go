package fulfillment

import "github.com/tuannm99/podzone/pkg/ddd"

var (
	ErrOrderIDRequired = ddd.NewDomainError("FULFILLMENT_ORDER_ID_REQUIRED", "fulfillment order id is required")
	ErrShipmentStatusInvalid = ddd.NewDomainError(
		"FULFILLMENT_STATUS_INVALID",
		"invalid shipment status",
	)
	ErrPartnerRequired = ddd.NewDomainError(
		"FULFILLMENT_PARTNER_REQUIRED",
		"fulfillment partner is required before shipping",
	)
	ErrCarrierRequired = ddd.NewDomainError("FULFILLMENT_CARRIER_REQUIRED", "shipment carrier is required")
)
