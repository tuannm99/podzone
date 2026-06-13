package routing

import "github.com/tuannm99/podzone/pkg/ddd"

var (
	ErrStoreScopeRequired    = ddd.NewDomainError("STORE_SCOPE_REQUIRED", "store scope is required")
	ErrRoutedOrderNotFound   = ddd.NewDomainError("ROUTED_ORDER_NOT_FOUND", "routed order not found")
	ErrRoutedOrderIDRequired = ddd.NewDomainError(
		"ROUTED_ORDER_ID_REQUIRED",
		"routed order id is required",
	)
	ErrRoutedOrderStoreRequired = ddd.NewDomainError(
		"ROUTED_ORDER_STORE_REQUIRED",
		"routed order store id is required",
	)
	ErrRoutedOrderQuantityInvalid = ddd.NewDomainError(
		"ROUTED_ORDER_QUANTITY_INVALID",
		"routed order quantity is invalid",
	)
	ErrRoutingCandidateRequired = ddd.NewDomainError(
		"ROUTING_CANDIDATE_REQUIRED",
		"routing candidate id is required",
	)
)
