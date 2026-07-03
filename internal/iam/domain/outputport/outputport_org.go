package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type OrganizationCommandRepository interface {
	Create(ctx context.Context, org entity.Organization) (*entity.Organization, error)
	EnsureRoot(ctx context.Context, org entity.Organization) (*entity.Organization, error)
	AttachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
	DetachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
}

type OrganizationQueryRepository interface {
	List(ctx context.Context, query collection.Query) (collection.Page[entity.Organization], error)
	GetByID(ctx context.Context, orgID string) (*entity.Organization, error)
	GetByRootUserID(ctx context.Context, userID uint) (*entity.Organization, error)
	IsRoot(ctx context.Context, orgID string, userID uint) (bool, error)
	ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error)
	ListServiceControlPolicyStatements(ctx context.Context, orgID string) ([]entity.PolicyStatement, error)
}

type OrganizationRepository interface {
	OrganizationCommandRepository
	OrganizationQueryRepository
}
