package domain

import "context"

type OrganizationRepository interface {
	Create(ctx context.Context, org Organization) (*Organization, error)
	List(ctx context.Context) ([]Organization, error)
	GetByID(ctx context.Context, orgID string) (*Organization, error)
	AttachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
	DetachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error
	ListServiceControlPolicies(ctx context.Context, orgID string) ([]Policy, error)
	ListServiceControlPolicyStatements(ctx context.Context, orgID string) ([]PolicyStatement, error)
}
