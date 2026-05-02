package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type supplierService struct {
	repo SupplierRepository
}

func NewSupplierUsecase(repo SupplierRepository) SupplierUsecase {
	return &supplierService{repo: repo}
}

func (s *supplierService) CreateSupplier(ctx context.Context, cmd CreateSupplierCmd) (*Supplier, error) {
	tenantID := strings.TrimSpace(cmd.TenantID)
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ErrInvalidSupplierName
	}

	code := normalizeSupplierCode(cmd.Code)
	if code == "" {
		code = normalizeSupplierCode(name)
	}
	if code == "" {
		return nil, ErrInvalidSupplierCode
	}
	partnerType := NormalizePartnerType(cmd.PartnerType)
	if partnerType == "" {
		return nil, ErrInvalidPartnerType
	}

	now := time.Now().UTC()
	return s.repo.Create(ctx, Supplier{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
		Code:         code,
		Name:         name,
		ContactName:  strings.TrimSpace(cmd.ContactName),
		ContactEmail: strings.TrimSpace(strings.ToLower(cmd.ContactEmail)),
		Notes:        strings.TrimSpace(cmd.Notes),
		PartnerType:  partnerType,
		Status:       SupplierStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *supplierService) GetSupplier(ctx context.Context, id string) (*Supplier, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidSupplierID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *supplierService) ListSuppliers(ctx context.Context, query ListSuppliersQuery) ([]Supplier, error) {
	query.TenantID = strings.TrimSpace(query.TenantID)
	if query.TenantID == "" {
		return nil, ErrInvalidTenantID
	}
	if query.Status != "" {
		query.Status = NormalizeSupplierStatus(query.Status)
		if query.Status == "" {
			return nil, ErrInvalidSupplierStatus
		}
	}
	if query.PartnerType != "" {
		query.PartnerType = NormalizePartnerType(query.PartnerType)
		if query.PartnerType == "" {
			return nil, ErrInvalidPartnerType
		}
	}
	return s.repo.List(ctx, query)
}

func (s *supplierService) UpdateSupplier(ctx context.Context, cmd UpdateSupplierCmd) (*Supplier, error) {
	id := strings.TrimSpace(cmd.ID)
	if id == "" {
		return nil, ErrInvalidSupplierID
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ErrInvalidSupplierName
	}
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cmd.PartnerType) != "" {
		partnerType := NormalizePartnerType(cmd.PartnerType)
		if partnerType == "" {
			return nil, ErrInvalidPartnerType
		}
		current.PartnerType = partnerType
	}
	current.Name = name
	current.ContactName = strings.TrimSpace(cmd.ContactName)
	current.ContactEmail = strings.TrimSpace(strings.ToLower(cmd.ContactEmail))
	current.Notes = strings.TrimSpace(cmd.Notes)
	current.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, *current)
}

func (s *supplierService) UpdateSupplierStatus(ctx context.Context, id, status string) (*Supplier, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidSupplierID
	}
	status = NormalizeSupplierStatus(status)
	if status == "" {
		return nil, ErrInvalidSupplierStatus
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func normalizeSupplierCode(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		return ""
	}
	return out
}
