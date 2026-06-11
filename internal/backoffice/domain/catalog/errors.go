package catalog

import "github.com/tuannm99/podzone/pkg/ddd"

var (
	ErrDraftIDRequired = ddd.NewDomainError("PRODUCT_DRAFT_ID_REQUIRED", "product setup draft id is required")
	ErrCandidateIDRequired = ddd.NewDomainError(
		"PRODUCT_CANDIDATE_ID_REQUIRED",
		"product setup candidate id is required",
	)
	ErrVariantIDRequired = ddd.NewDomainError(
		"PRODUCT_VARIANT_ID_REQUIRED",
		"product setup variant id is required",
	)
	ErrStoreIDRequired = ddd.NewDomainError("STORE_ID_REQUIRED", "product setup store id is required")
	ErrProductNameRequired = ddd.NewDomainError("PRODUCT_NAME_REQUIRED", "product setup name is required")
	ErrDraftStatusInvalid = ddd.NewDomainError(
		"PRODUCT_DRAFT_STATUS_INVALID",
		"invalid product setup draft status",
	)
)
