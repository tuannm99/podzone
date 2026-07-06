package entity

import "errors"

var (
	ErrInvalidInput         = errors.New("infrasmanager: invalid input")
	ErrUnsupportedInfraType = errors.New("infrasmanager: unsupported infra type")
	ErrConnectionNotFound   = errors.New("infrasmanager: connection not found")
	ErrResourceNotFound     = errors.New("infrasmanager: resource not found")
	ErrAccessDenied         = errors.New("infrasmanager: access denied")
)
