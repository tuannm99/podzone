package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type iamService struct {
	tenants     TenantRepository
	roles       RoleRepository
	memberships MembershipRepository
}

func NewIAMUsecase(
	tenants TenantRepository,
	roles RoleRepository,
	memberships MembershipRepository,
) IAMUsecase {
	return &iamService{
		tenants:     tenants,
		roles:       roles,
		memberships: memberships,
	}
}

func (s *iamService) CreateTenant(ctx context.Context, ownerUserID uint, cmd CreateTenantCmd) (*Tenant, error) {
	if ownerUserID == 0 {
		return nil, ErrInvalidUserID
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ErrInvalidTenantName
	}
	slug := strings.TrimSpace(cmd.Slug)
	if slug == "" {
		return nil, ErrInvalidTenantSlug
	}

	now := time.Now().UTC()
	tenant, err := s.tenants.Create(ctx, Tenant{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, err
	}

	role, err := s.roles.GetByName(ctx, RoleTenantOwner)
	if err != nil {
		return nil, err
	}

	if err := s.memberships.Upsert(ctx, Membership{
		TenantID:  tenant.ID,
		UserID:    ownerUserID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *iamService) AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error {
	if strings.TrimSpace(tenantID) == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return ErrInvalidRoleName
	}

	if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
		return err
	}
	role, err := s.roles.GetByName(ctx, roleName)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return s.memberships.Upsert(ctx, Membership{
		TenantID:  tenantID,
		UserID:    userID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *iamService) GetMembership(ctx context.Context, tenantID string, userID uint) (*Membership, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.memberships.GetByTenantAndUser(ctx, tenantID, userID)
}

func (s *iamService) ListUserTenants(ctx context.Context, userID uint) ([]Membership, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.memberships.ListByUser(ctx, userID)
}

func (s *iamService) ListTenantMembers(ctx context.Context, tenantID string) ([]Membership, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantNotFound
	}
	return s.memberships.ListByTenant(ctx, tenantID)
}

func (s *iamService) RemoveMember(ctx context.Context, tenantID string, userID uint) error {
	if strings.TrimSpace(tenantID) == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.memberships.Delete(ctx, tenantID, userID)
}

func (s *iamService) CheckPermission(
	ctx context.Context,
	tenantID string,
	userID uint,
	permission string,
) (bool, error) {
	membership, err := s.GetMembership(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	if membership.Status != MembershipStatusActive {
		return false, ErrInactiveMembership
	}
	return s.roles.RoleHasPermission(ctx, membership.RoleID, permission)
}

func (s *iamService) RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error {
	allowed, err := s.CheckPermission(ctx, tenantID, userID, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}
