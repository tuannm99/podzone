package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/entity"
)

type AuthzUsecase interface {
	AssumeRole(ctx context.Context, input entity.AssumeRoleInput) (*entity.AssumedRole, error)
	CheckPermission(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
	RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error
	SimulateAccess(ctx context.Context, input entity.SimulateAccessInput) (*entity.SimulateAccessResult, error)
}
