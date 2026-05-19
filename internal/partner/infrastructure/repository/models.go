package repository

import (
	"time"

	"github.com/lib/pq"
	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
)

type partnerModel struct {
	ID                    string         `db:"id"`
	TenantID              string         `db:"tenant_id"`
	Code                  string         `db:"code"`
	Name                  string         `db:"name"`
	ContactName           string         `db:"contact_name"`
	ContactEmail          string         `db:"contact_email"`
	Notes                 string         `db:"notes"`
	PartnerType           string         `db:"partner_type"`
	Status                string         `db:"status"`
	SupportedProductTypes pq.StringArray `db:"supported_product_types"`
	SupportedRegions      pq.StringArray `db:"supported_regions"`
	SLADays               int32          `db:"sla_days"`
	RoutingPriority       int32          `db:"routing_priority"`
	CreatedAt             time.Time      `db:"created_at"`
	UpdatedAt             time.Time      `db:"updated_at"`
}

func (m partnerModel) toEntity() *partnerdomain.Partner {
	return &partnerdomain.Partner{
		ID:                    m.ID,
		TenantID:              m.TenantID,
		Code:                  m.Code,
		Name:                  m.Name,
		ContactName:           m.ContactName,
		ContactEmail:          m.ContactEmail,
		Notes:                 m.Notes,
		PartnerType:           m.PartnerType,
		Status:                m.Status,
		SupportedProductTypes: []string(m.SupportedProductTypes),
		SupportedRegions:      []string(m.SupportedRegions),
		SLADays:               m.SLADays,
		RoutingPriority:       m.RoutingPriority,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}
