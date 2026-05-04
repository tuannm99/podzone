package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type partnerService struct {
	repo PartnerRepository
}

func NewPartnerUsecase(repo PartnerRepository) PartnerUsecase {
	return &partnerService{repo: repo}
}

func (s *partnerService) CreatePartner(ctx context.Context, cmd CreatePartnerCmd) (*Partner, error) {
	tenantID := strings.TrimSpace(cmd.TenantID)
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ErrInvalidPartnerName
	}

	code := normalizePartnerCode(cmd.Code)
	if code == "" {
		code = normalizePartnerCode(name)
	}
	if code == "" {
		return nil, ErrInvalidPartnerCode
	}
	partnerType := NormalizePartnerType(cmd.PartnerType)
	if partnerType == "" {
		return nil, ErrInvalidPartnerType
	}

	now := time.Now().UTC()
	return s.repo.Create(ctx, Partner{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
		Code:         code,
		Name:         name,
		ContactName:  strings.TrimSpace(cmd.ContactName),
		ContactEmail: strings.TrimSpace(strings.ToLower(cmd.ContactEmail)),
		Notes:        strings.TrimSpace(cmd.Notes),
		PartnerType:  partnerType,
		Status:       PartnerStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *partnerService) GetPartner(ctx context.Context, id string) (*Partner, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidPartnerID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *partnerService) ListPartners(ctx context.Context, query ListPartnersQuery) ([]Partner, error) {
	query.TenantID = strings.TrimSpace(query.TenantID)
	if query.TenantID == "" {
		return nil, ErrInvalidTenantID
	}
	if query.Status != "" {
		query.Status = NormalizePartnerStatus(query.Status)
		if query.Status == "" {
			return nil, ErrInvalidPartnerStatus
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

func (s *partnerService) UpdatePartner(ctx context.Context, cmd UpdatePartnerCmd) (*Partner, error) {
	id := strings.TrimSpace(cmd.ID)
	if id == "" {
		return nil, ErrInvalidPartnerID
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ErrInvalidPartnerName
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

func (s *partnerService) UpdatePartnerStatus(ctx context.Context, id, status string) (*Partner, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidPartnerID
	}
	status = NormalizePartnerStatus(status)
	if status == "" {
		return nil, ErrInvalidPartnerStatus
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func normalizePartnerCode(raw string) string {
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
