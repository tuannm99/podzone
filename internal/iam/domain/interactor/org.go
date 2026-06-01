package interactor

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

func (s *interactor) CreateOrganization(ctx context.Context, name string, slug string) (*entity.Organization, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, entity.ErrInvalidOrganizationName
	}
	if slug == "" {
		return nil, entity.ErrInvalidOrganizationSlug
	}
	now := time.Now().UTC()
	return s.orgCommands.Create(ctx, entity.Organization{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *interactor) ListOrganizations(ctx context.Context) ([]entity.Organization, error) {
	return s.orgQueries.List(ctx)
}

func (s *interactor) AttachTenantToOrganization(ctx context.Context, tenantID string, orgID string) error {
	tenantID = strings.TrimSpace(tenantID)
	orgID = strings.TrimSpace(orgID)
	if tenantID == "" {
		return entity.ErrTenantNotFound
	}
	if orgID == "" {
		return entity.ErrOrganizationNotFound
	}
	if _, err := s.tenantQueries.GetByID(ctx, tenantID); err != nil {
		return err
	}
	if _, err := s.orgQueries.GetByID(ctx, orgID); err != nil {
		return err
	}
	return s.tenantCommands.AttachOrganization(ctx, tenantID, orgID)
}

func (s *interactor) DetachTenantFromOrganization(ctx context.Context, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return entity.ErrTenantNotFound
	}
	return s.tenantCommands.DetachOrganization(ctx, tenantID)
}

func (s *interactor) AttachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return entity.ErrOrganizationNotFound
	}
	policy, err := s.policyQueries.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.orgCommands.AttachServiceControlPolicy(ctx, orgID, policy.ID)
}

func (s *interactor) DetachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return entity.ErrOrganizationNotFound
	}
	policy, err := s.policyQueries.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.orgCommands.DetachServiceControlPolicy(ctx, orgID, policy.ID)
}

func (s *interactor) ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error) {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return nil, entity.ErrOrganizationNotFound
	}
	return s.orgQueries.ListServiceControlPolicies(ctx, orgID)
}
