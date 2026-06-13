package store

import "github.com/tuannm99/podzone/pkg/ddd"

var (
	ErrStoreIDRequired    = ddd.NewDomainError("STORE_ID_REQUIRED", "store id is required")
	ErrStoreNameRequired  = ddd.NewDomainError("STORE_NAME_REQUIRED", "store name is required")
	ErrStoreOwnerRequired = ddd.NewDomainError("STORE_OWNER_REQUIRED", "store owner id is required")
	ErrStoreTimeRequired  = ddd.NewDomainError("STORE_TIME_REQUIRED", "store time is required")
)
