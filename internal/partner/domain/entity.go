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
	ErrPartnerNotFound      = errors.New("partner not found")
	ErrPartnerCodeTaken     = errors.New("partner code already exists in store")
	ErrInvalidPartnerID     = errors.New("invalid partner id")
	ErrInvalidPartnerCode   = errors.New("invalid partner code")
	ErrInvalidPartnerName   = errors.New("invalid partner name")
	ErrInvalidTenantID      = errors.New("invalid tenant id")
	ErrInvalidPartnerType   = errors.New("invalid partner type")
	ErrInvalidPartnerStatus = errors.New("invalid partner status")
)

type Partner struct {
	ID                    string
	TenantID              string
	Code                  string
	Name                  string
	ContactName           string
	ContactEmail          string
	Notes                 string
	PartnerType           string
	Status                string
	SupportedProductTypes []string
	SupportedRegions      []string
	SLADays               int32
	RoutingPriority       int32
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CreatePartnerCmd struct {
	TenantID              string
	Code                  string
	Name                  string
	ContactName           string
	ContactEmail          string
	Notes                 string
	PartnerType           string
	SupportedProductTypes []string
	SupportedRegions      []string
	SLADays               int32
	RoutingPriority       int32
}

type UpdatePartnerCmd struct {
	ID                    string
	Name                  string
	ContactName           string
	ContactEmail          string
	Notes                 string
	PartnerType           string
	SupportedProductTypes []string
	SupportedRegions      []string
	SLADays               int32
	RoutingPriority       int32
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

func NormalizeCapabilityList(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(strings.ToLower(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}
