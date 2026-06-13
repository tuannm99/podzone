package order

import "github.com/tuannm99/podzone/pkg/ddd"

var (
	ErrOrderIDRequired     = ddd.NewDomainError("ORDER_ID_REQUIRED", "customer order id is required")
	ErrStoreIDRequired     = ddd.NewDomainError("STORE_ID_REQUIRED", "customer order store id is required")
	ErrCandidateIDRequired = ddd.NewDomainError(
		"ORDER_CANDIDATE_ID_REQUIRED",
		"customer order candidate id is required",
	)
	ErrProductTitleRequired = ddd.NewDomainError(
		"ORDER_PRODUCT_TITLE_REQUIRED",
		"customer order product title is required",
	)
	ErrQuantityInvalid       = ddd.NewDomainError("ORDER_QUANTITY_INVALID", "customer order quantity is invalid")
	ErrRoutingReasonRequired = ddd.NewDomainError(
		"ORDER_ROUTING_REASON_REQUIRED",
		"routing block reason is required when partner is empty",
	)
	ErrActiveException = ddd.NewDomainError(
		"ORDER_ACTIVE_EXCEPTION",
		"resolve the active exception before advancing the routed order",
	)
	ErrRoutingBlocked = ddd.NewDomainError(
		"ORDER_ROUTING_BLOCKED",
		"resolve the routing block before advancing the routed order",
	)
	ErrStatusInvalid           = ddd.NewDomainError("ORDER_STATUS_INVALID", "invalid customer order status")
	ErrSettlementStatusInvalid = ddd.NewDomainError(
		"ORDER_SETTLEMENT_STATUS_INVALID",
		"invalid settlement status",
	)
	ErrNotRoutingBlocked = ddd.NewDomainError(
		"ORDER_NOT_ROUTING_BLOCKED",
		"customer order is not routing blocked",
	)
	ErrRoutingPartnerRequired = ddd.NewDomainError(
		"ORDER_ROUTING_PARTNER_REQUIRED",
		"selected routing partner is required",
	)
)
