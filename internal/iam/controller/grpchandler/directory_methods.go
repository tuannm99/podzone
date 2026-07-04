package grpchandler

import (
	"context"
	"fmt"
	"strings"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
	"github.com/tuannm99/podzone/pkg/collection"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *IAMQueryServer) ListDirectoryUsers(
	ctx context.Context,
	req *pbiamv1.ListDirectoryUsersRequest,
) (*pbiamv1.ListDirectoryUsersResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.requireDirectoryAccess(
		ctx,
		actorUserID,
		req.GetScope(),
		req.GetOrgId(),
		req.GetTenantId(),
	); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListDirectoryUsers(ctx, iammapper.ToCollectionQuery(req.GetCollection()))
	if err != nil {
		return nil, iamStatusError(err)
	}
	users := make([]*pbiamv1.DirectoryUser, 0, len(page.Items))
	for _, user := range page.Items {
		users = append(users, iammapper.ToPBDirectoryUser(user))
	}
	return &pbiamv1.ListDirectoryUsersResponse{
		Users:    users,
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) ListPermissions(
	ctx context.Context,
	req *pbiamv1.ListPermissionsRequest,
) (*pbiamv1.ListPermissionsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.requireDirectoryAccess(
		ctx,
		actorUserID,
		req.GetScope(),
		req.GetOrgId(),
		req.GetTenantId(),
	); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListPermissions(ctx, iammapper.ToCollectionQuery(req.GetCollection()))
	if err != nil {
		return nil, iamStatusError(err)
	}
	permissions := make([]*pbiamv1.Permission, 0, len(page.Items))
	for _, permission := range page.Items {
		permissions = append(permissions, iammapper.ToPBPermission(permission))
	}
	return &pbiamv1.ListPermissionsResponse{
		Permissions: permissions,
		PageInfo:    iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) requireDirectoryAccess(
	ctx context.Context,
	actorUserID uint,
	scope string,
	orgID string,
	tenantID string,
) error {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		switch {
		case strings.TrimSpace(orgID) != "":
			scope = entity.PolicyScopeOrganization
		case strings.TrimSpace(tenantID) != "":
			scope = entity.PolicyScopeTenant
		default:
			scope = entity.PolicyScopePlatform
		}
	}
	switch scope {
	case entity.PolicyScopePlatform:
		return s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles")
	case entity.PolicyScopeOrganization:
		if strings.TrimSpace(orgID) == "" {
			return fmt.Errorf("%w: organization id is required", collection.ErrInvalidQuery)
		}
		return s.queries.RequireOrganizationPermission(
			ctx,
			strings.TrimSpace(orgID),
			actorUserID,
			"organization:manage_iam",
		)
	case entity.PolicyScopeTenant:
		if strings.TrimSpace(tenantID) == "" {
			return fmt.Errorf("%w: tenant id is required", collection.ErrInvalidQuery)
		}
		return s.queries.RequirePermission(
			ctx,
			strings.TrimSpace(tenantID),
			actorUserID,
			"tenant:manage_members",
		)
	default:
		return fmt.Errorf("%w: unsupported directory scope %q", collection.ErrInvalidQuery, scope)
	}
}
