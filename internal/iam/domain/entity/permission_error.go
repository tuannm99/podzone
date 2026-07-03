package entity

import (
	"fmt"
	"strings"
)

type PermissionDeniedError struct {
	Permission string
	Resource   string
}

func NewPermissionDeniedError(permission string, resource string) *PermissionDeniedError {
	return &PermissionDeniedError{
		Permission: strings.TrimSpace(permission),
		Resource:   strings.TrimSpace(resource),
	}
}

func (e *PermissionDeniedError) Error() string {
	if e == nil || e.Permission == "" {
		return ErrPermissionDenied.Error()
	}
	if e.Resource == "" || e.Resource == "*" {
		return fmt.Sprintf("iam: missing permission %q", e.Permission)
	}
	return fmt.Sprintf("iam: missing permission %q on %q", e.Permission, e.Resource)
}

func (e *PermissionDeniedError) Unwrap() error {
	return ErrPermissionDenied
}
