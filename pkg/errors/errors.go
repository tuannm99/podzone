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

type AppError struct {
	Type    ErrorType
	Message string
	Err     error
	Fields  map[string]string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

func (e *AppError) WithField(key, value string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[key] = value
	return e
}

func (e *AppError) WithFields(fields map[string]string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	maps.Copy(e.Fields, fields)
	return e
}

func NewInternal(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

func NewValidation(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

func NewNotFound(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Err:     err,
	}
}

func NewUnauthorized(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Err:     err,
	}
}

func NewForbidden(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
		Err:     err,
	}
}

func NewConflict(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
		Err:     err,
	}
}

func NewTimeout(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Err:     err,
	}
}

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

func IsValidation(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeValidation
	}
	return false
}

func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeNotFound
	}
	return false
}

func IsUnauthorized(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeUnauthorized
	}
	return false
}

func IsForbidden(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeForbidden
	}
	return false
}

func IsConflict(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeConflict
	}
	return false
}

func IsTimeout(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeTimeout
	}
	return false
}

func IsBadRequest(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeBadRequest
	}
	return false
}

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

func WrapIfErr(message string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

type ValidationError struct {
	Errors map[string]string
}

func NewValidationErrors() *ValidationError {
	return &ValidationError{
		Errors: make(map[string]string),
	}
}

func (v *ValidationError) Add(field, message string) {
	v.Errors[field] = message
}

func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

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

func (v *ValidationError) ToAppError() *AppError {
	return NewValidation("validation errors", v).WithFields(v.Errors)
}
