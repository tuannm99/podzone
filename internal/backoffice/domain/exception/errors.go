package exception

import "github.com/tuannm99/podzone/pkg/ddd"

var (
	ErrOrderIDRequired = ddd.NewDomainError("EXCEPTION_ORDER_ID_REQUIRED", "order id is required")
	ErrStatusInvalid = ddd.NewDomainError("EXCEPTION_STATUS_INVALID", "invalid exception status")
	ErrTypeInvalid = ddd.NewDomainError("EXCEPTION_TYPE_INVALID", "invalid exception type")
	ErrNoActiveType = ddd.NewDomainError(
		"EXCEPTION_TYPE_MISSING",
		"routed order has no active exception type",
	)
	ErrResolvedCannotReopen = ddd.NewDomainError(
		"EXCEPTION_RESOLVED",
		"resolved exception cannot be reopened by status update",
	)
)
