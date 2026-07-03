package entity

import "errors"

var (
	ErrInvalidInput         = errors.New("infrasmanager: invalid input")
	ErrUnsupportedInfraType = errors.New("infrasmanager: unsupported infra type")
	ErrConnectionNotFound   = errors.New("infrasmanager: connection not found")
	ErrAccessDenied         = errors.New("infrasmanager: access denied")
)
