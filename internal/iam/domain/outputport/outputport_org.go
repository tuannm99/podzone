package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type OrganizationCommandRepository interface {
	Create(ctx context.Context, org entity.Organization) (*entity.Organization, error)
	AttachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
	DetachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
}

type OrganizationQueryRepository interface {
	List(ctx context.Context) ([]entity.Organization, error)
	GetByID(ctx context.Context, orgID string) (*entity.Organization, error)
	ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error)
	ListServiceControlPolicyStatements(ctx context.Context, orgID string) ([]entity.PolicyStatement, error)
}

type OrganizationRepository interface {
	OrganizationCommandRepository
	OrganizationQueryRepository
}
