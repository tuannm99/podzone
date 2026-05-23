package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/entity"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org entity.Organization) (*entity.Organization, error)
	List(ctx context.Context) ([]entity.Organization, error)
	GetByID(ctx context.Context, orgID string) (*entity.Organization, error)
	AttachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
	DetachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
	ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error)
	ListServiceControlPolicyStatements(ctx context.Context, orgID string) ([]entity.PolicyStatement, error)
}
