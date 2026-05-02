package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	SupplierStatusActive   = "active"
	SupplierStatusInactive = "inactive"

	PartnerTypePrintOnDemand = "print_on_demand"
	PartnerTypeFulfillment   = "fulfillment"
	PartnerTypeDropship      = "dropship_supplier"
)

var (
	ErrSupplierNotFound      = errors.New("supplier not found")
	ErrSupplierCodeTaken     = errors.New("supplier code already exists in store")
	ErrInvalidSupplierID     = errors.New("invalid supplier id")
	ErrInvalidSupplierCode   = errors.New("invalid supplier code")
	ErrInvalidSupplierName   = errors.New("invalid supplier name")
	ErrInvalidTenantID       = errors.New("invalid tenant id")
	ErrInvalidPartnerType    = errors.New("invalid partner type")
	ErrInvalidSupplierStatus = errors.New("invalid supplier status")
)

type Supplier struct {
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

type CreateSupplierCmd struct {
	TenantID     string
	Code         string
	Name         string
	ContactName  string
	ContactEmail string
	Notes        string
	PartnerType  string
}

type UpdateSupplierCmd struct {
	ID           string
	Name         string
	ContactName  string
	ContactEmail string
	Notes        string
	PartnerType  string
}

type ListSuppliersQuery struct {
	TenantID    string
	Status      string
	PartnerType string
}

func NormalizeSupplierStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", SupplierStatusActive:
		return SupplierStatusActive
	case SupplierStatusInactive:
		return SupplierStatusInactive
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
