package ddd

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("aggregate not found")
	ErrVersionConflict     = errors.New("aggregate version conflict")
	ErrMissingEventApplier = errors.New("event applier is required")
)

type DomainError struct {
	Code    string
	Message string
}

var _ error = (*DomainError)(nil)

func NewDomainError(code string, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func (e *DomainError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return e.Message
	}
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func IsDomainError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr)
}
