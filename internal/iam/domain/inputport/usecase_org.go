package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type OrganizationUsecase interface {
	CreateOrganization(ctx context.Context, name string, slug string) (*entity.Organization, error)
	ListOrganizations(ctx context.Context) ([]entity.Organization, error)
	AttachTenantToOrganization(ctx context.Context, tenantID string, orgID string) error
	DetachTenantFromOrganization(ctx context.Context, tenantID string) error
	AttachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error
	DetachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error
	ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error)
}
