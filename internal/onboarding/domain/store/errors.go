package store

import "errors"

var (
	ErrStoreNotFound       = errors.New("store request not found")
	ErrSubdomainTaken      = errors.New("subdomain is already taken")
	ErrInvalidStatus       = errors.New("invalid store request status")
	ErrStoreNotActive      = errors.New("store is not active")
	ErrStoreNotCompleted   = errors.New("store onboarding is not completed")
	ErrNameRequired        = errors.New("name is required")
	ErrSubdomainRequired   = errors.New("subdomain is required")
	ErrWorkspaceIDRequired = errors.New("workspace_id is required")
	ErrRequestedByRequired = errors.New("requested_by is required")
)
