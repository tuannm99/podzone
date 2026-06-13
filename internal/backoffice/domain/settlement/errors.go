package settlement

import "github.com/tuannm99/podzone/pkg/ddd"

var (
	ErrOrderIDRequired        = ddd.NewDomainError("SETTLEMENT_ORDER_ID_REQUIRED", "settlement order id is required")
	ErrStatusInvalid          = ddd.NewDomainError("SETTLEMENT_STATUS_INVALID", "invalid settlement status")
	ErrFulfillmentCostInvalid = ddd.NewDomainError(
		"SETTLEMENT_FULFILLMENT_COST_INVALID",
		"invalid fulfillment cost",
	)
	ErrShippingCostInvalid = ddd.NewDomainError(
		"SETTLEMENT_SHIPPING_COST_INVALID",
		"invalid shipping cost",
	)
	ErrIssueContextRequired = ddd.NewDomainError(
		"SETTLEMENT_ISSUE_CONTEXT_REQUIRED",
		"issue cost handling requires an active exception or delivery issue",
	)
	ErrIssueCostInvalid       = ddd.NewDomainError("SETTLEMENT_ISSUE_COST_INVALID", "invalid issue cost")
	ErrIssueResolutionInvalid = ddd.NewDomainError(
		"SETTLEMENT_ISSUE_RESOLUTION_INVALID",
		"invalid issue resolution",
	)
)
