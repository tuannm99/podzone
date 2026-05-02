package repository

import (
	"time"

	supplierdomain "github.com/tuannm99/podzone/internal/partner/domain"
)

type supplierModel struct {
	ID           string    `db:"id"`
	TenantID     string    `db:"tenant_id"`
	Code         string    `db:"code"`
	Name         string    `db:"name"`
	ContactName  string    `db:"contact_name"`
	ContactEmail string    `db:"contact_email"`
	Notes        string    `db:"notes"`
	PartnerType  string    `db:"partner_type"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (m supplierModel) toEntity() *supplierdomain.Supplier {
	return &supplierdomain.Supplier{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Code:         m.Code,
		Name:         m.Name,
		ContactName:  m.ContactName,
		ContactEmail: m.ContactEmail,
		Notes:        m.Notes,
		PartnerType:  m.PartnerType,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
