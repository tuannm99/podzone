package messaging

import "errors"

var (
	ErrNoMessages  = errors.New("messaging: no messages")
	ErrNilHandler  = errors.New("messaging: handler is nil")
	ErrNilRegistry = errors.New("messaging: registry is nil")
)
