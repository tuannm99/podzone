package grpchandler

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/inputport"
)

func requirePolicyAccess(
	ctx context.Context,
	queries inputport.IAMQueryUsecase,
	actorUserID uint,
	scope string,
	orgID string,
	mutate bool,
) error {
	if scope != entity.PolicyScopeOrganization {
		return queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles")
	}
	permission := "organization:read"
	if mutate {
		permission = "organization:manage_iam"
	}
	return queries.RequireOrganizationPermission(ctx, orgID, actorUserID, permission)
}

func requireGroupAccess(
	ctx context.Context,
	queries inputport.IAMQueryUsecase,
	actorUserID uint,
	groupID uint64,
	mutate bool,
) (*entity.Group, error) {
	group, err := queries.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	switch group.Scope {
	case entity.PolicyScopeOrganization:
		permission := "organization:read"
		if mutate {
			permission = "organization:manage_iam"
		}
		err = queries.RequireOrganizationPermission(ctx, group.OrgID, actorUserID, permission)
	case entity.PolicyScopeTenant:
		err = queries.RequirePermission(ctx, group.TenantID, actorUserID, "tenant:manage_members")
	default:
		err = queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles")
	}
	if err != nil {
		return nil, err
	}
	return group, nil
}
