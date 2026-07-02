package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type GroupCommandRepository interface {
	CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupID uint64) error
	PutInlinePolicy(ctx context.Context, input entity.PutGroupInlinePolicyInput) error
	DeleteInlinePolicy(ctx context.Context, groupID uint64, name string) error
	AddMember(ctx context.Context, groupID uint64, userID uint) error
	RemoveMember(ctx context.Context, groupID uint64, userID uint) error
	AttachPolicy(ctx context.Context, groupID uint64, policyID uint64) error
	DetachPolicy(ctx context.Context, groupID uint64, policyID uint64) error
}

type GroupQueryRepository interface {
	GetByID(ctx context.Context, groupID uint64) (*entity.Group, error)
	ListGroups(
		ctx context.Context,
		scope string,
		tenantID string,
		query collection.Query,
	) (collection.Page[entity.Group], error)
	GetInlinePolicy(ctx context.Context, groupID uint64, name string) (*entity.GroupInlinePolicy, error)
	ListInlinePolicies(
		ctx context.Context,
		groupID uint64,
		query collection.Query,
	) (collection.Page[entity.GroupInlinePolicy], error)
	ListMembers(ctx context.Context, groupID uint64, query collection.Query) (collection.Page[uint], error)
	ListPolicies(
		ctx context.Context,
		groupID uint64,
		query collection.Query,
	) (collection.Page[entity.Policy], error)
}

type GroupRepository interface {
	GroupCommandRepository
	GroupQueryRepository
}
