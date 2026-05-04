package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	PartnerStatusActive   = "active"
	PartnerStatusInactive = "inactive"

	PartnerTypePrintOnDemand = "print_on_demand"
	PartnerTypeFulfillment   = "fulfillment"
	PartnerTypeDropship      = "dropship_supplier"
)

var (
	ErrPartnerNotFound       = errors.New("partner not found")
	ErrPartnerCodeTaken      = errors.New("partner code already exists in store")
	ErrInvalidPartnerID      = errors.New("invalid partner id")
	ErrInvalidPartnerCode    = errors.New("invalid partner code")
	ErrInvalidPartnerName    = errors.New("invalid partner name")
	ErrInvalidTenantID       = errors.New("invalid tenant id")
	ErrInvalidPartnerType    = errors.New("invalid partner type")
	ErrInvalidPartnerStatus  = errors.New("invalid partner status")
)

type Partner struct {
	ID           string
	TenantID     string
	Code         string
	Name         string
	ContactName  string
	ContactEmail string
	Notes        string
	PartnerType  string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreatePartnerCmd struct {
	TenantID     string
	Code         string
	Name         string
	ContactName  string
	ContactEmail string
	Notes        string
	PartnerType  string
}

type UpdatePartnerCmd struct {
	ID           string
	Name         string
	ContactName  string
	ContactEmail string
	Notes        string
	PartnerType  string
}

type ListPartnersQuery struct {
	TenantID    string
	Status      string
	PartnerType string
}

func NormalizePartnerStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", PartnerStatusActive:
		return PartnerStatusActive
	case PartnerStatusInactive:
		return PartnerStatusInactive
	default:
		return ""
	}
}

func NormalizePartnerType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", PartnerTypePrintOnDemand:
		return PartnerTypePrintOnDemand
	case PartnerTypeFulfillment:
		return PartnerTypeFulfillment
	case PartnerTypeDropship:
		return PartnerTypeDropship
	default:
		return ""
	}
}
