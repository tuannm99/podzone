package errors

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorType is the type of error
type ErrorType string

const (
	// ErrorTypeInternal represents an internal server error
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeValidation represents a validation error
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound represents a not found error
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeUnauthorized represents an unauthorized error
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	// ErrorTypeForbidden represents a forbidden error
	ErrorTypeForbidden ErrorType = "forbidden"
	// ErrorTypeConflict represents a conflict error
	ErrorTypeConflict ErrorType = "conflict"
	// ErrorTypeTimeout represents a timeout error
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeBadRequest represents a bad request error
	ErrorTypeBadRequest ErrorType = "bad_request"
)

// AppError represents an application error
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
	Fields  map[string]string
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap implements the unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// WithField adds a field to the error
func (e *AppError) WithField(key, value string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[key] = value
	return e
}

// WithFields adds multiple fields to the error
func (e *AppError) WithFields(fields map[string]string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	maps.Copy(e.Fields, fields)
	return e
}

// NewInternal creates a new internal error
func NewInternal(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

// NewValidation creates a new validation error
func NewValidation(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

// NewNotFound creates a new not found error
func NewNotFound(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Err:     err,
	}
}

// NewUnauthorized creates a new unauthorized error
func NewUnauthorized(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Err:     err,
	}
}

// NewForbidden creates a new forbidden error
func NewForbidden(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
		Err:     err,
	}
}

// NewConflict creates a new conflict error
func NewConflict(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
		Err:     err,
	}
}

// NewTimeout creates a new timeout error
func NewTimeout(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Err:     err,
	}
}

// NewBadRequest creates a new bad request error
func NewBadRequest(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeBadRequest,
		Message: message,
		Err:     err,
	}
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeInternal
	}
	return false
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeValidation
	}
	return false
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeNotFound
	}
	return false
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeUnauthorized
	}
	return false
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeForbidden
	}
	return false
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeConflict
	}
	return false
}

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeTimeout
	}
	return false
}

// IsBadRequest checks if an error is a bad request error
func IsBadRequest(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeBadRequest
	}
	return false
}

// HTTPStatusCode returns the HTTP status code for an error
func HTTPStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case ErrorTypeInternal:
			return http.StatusInternalServerError
		case ErrorTypeValidation:
			return http.StatusBadRequest
		case ErrorTypeNotFound:
			return http.StatusNotFound
		case ErrorTypeUnauthorized:
			return http.StatusUnauthorized
		case ErrorTypeForbidden:
			return http.StatusForbidden
		case ErrorTypeConflict:
			return http.StatusConflict
		case ErrorTypeTimeout:
			return http.StatusRequestTimeout
		case ErrorTypeBadRequest:
			return http.StatusBadRequest
		}
	}
	return http.StatusInternalServerError
}

// GRPCStatusCode returns the gRPC status code for an error
func GRPCStatusCode(err error) codes.Code {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case ErrorTypeInternal:
			return codes.Internal
		case ErrorTypeValidation:
			return codes.InvalidArgument
		case ErrorTypeNotFound:
			return codes.NotFound
		case ErrorTypeUnauthorized:
			return codes.Unauthenticated
		case ErrorTypeForbidden:
			return codes.PermissionDenied
		case ErrorTypeConflict:
			return codes.AlreadyExists
		case ErrorTypeTimeout:
			return codes.DeadlineExceeded
		case ErrorTypeBadRequest:
			return codes.InvalidArgument
		}
	}
	return codes.Internal
}

// ToGRPCStatus converts an error to a gRPC status
func ToGRPCStatus(err error) *status.Status {
	var appErr *AppError
	if errors.As(err, &appErr) {
		code := GRPCStatusCode(err)
		msg := appErr.Message
		if appErr.Err != nil && msg == "" {
			msg = appErr.Err.Error()
		}
		s := status.New(code, msg)
		// Add details if needed
		return s
	}
	return status.New(codes.Internal, err.Error())
}

// FromGRPCStatus converts a gRPC status to an error
func FromGRPCStatus(s *status.Status) error {
	code := s.Code()
	msg := s.Message()

	switch code {
	case codes.Internal:
		return NewInternal(msg, nil)
	case codes.InvalidArgument:
		return NewValidation(msg, nil)
	case codes.NotFound:
		return NewNotFound(msg, nil)
	case codes.Unauthenticated:
		return NewUnauthorized(msg, nil)
	case codes.PermissionDenied:
		return NewForbidden(msg, nil)
	case codes.AlreadyExists:
		return NewConflict(msg, nil)
	case codes.DeadlineExceeded:
		return NewTimeout(msg, nil)
	default:
		return NewInternal(msg, nil)
	}
}

// WrapIfErr wraps an error if it's not nil
func WrapIfErr(message string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// ValidationError represents multiple validation errors
type ValidationError struct {
	Errors map[string]string
}

// NewValidationErrors creates a new validation errors
func NewValidationErrors() *ValidationError {
	return &ValidationError{
		Errors: make(map[string]string),
	}
}

// Add adds a validation error
func (v *ValidationError) Add(field, message string) {
	v.Errors[field] = message
}

// HasErrors checks if there are any errors
func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString("validation errors: ")
	i := 0
	for field, message := range v.Errors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(field)
		sb.WriteString(": ")
		sb.WriteString(message)
		i++
	}
	return sb.String()
}

// ToAppError converts validation errors to an app error
func (v *ValidationError) ToAppError() *AppError {
	return NewValidation("validation errors", v).WithFields(v.Errors)
}
