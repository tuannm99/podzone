package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type GroupRepository interface {
	CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error)
	GetByID(ctx context.Context, groupID uint64) (*entity.Group, error)
	ListGroups(ctx context.Context, scope string, tenantID string) ([]entity.Group, error)
	DeleteGroup(ctx context.Context, groupID uint64) error
	PutInlinePolicy(ctx context.Context, input entity.PutGroupInlinePolicyInput) error
	GetInlinePolicy(ctx context.Context, groupID uint64, name string) (*entity.GroupInlinePolicy, error)
	ListInlinePolicies(ctx context.Context, groupID uint64) ([]entity.GroupInlinePolicy, error)
	DeleteInlinePolicy(ctx context.Context, groupID uint64, name string) error
	AddMember(ctx context.Context, groupID uint64, userID uint) error
	RemoveMember(ctx context.Context, groupID uint64, userID uint) error
	ListMembers(ctx context.Context, groupID uint64) ([]uint, error)
	AttachPolicy(ctx context.Context, groupID uint64, policyID uint64) error
	DetachPolicy(ctx context.Context, groupID uint64, policyID uint64) error
	ListPolicies(ctx context.Context, groupID uint64) ([]entity.Policy, error)
}
