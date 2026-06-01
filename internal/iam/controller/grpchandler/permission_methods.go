package grpchandler

import (
	"context"
	"errors"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

func (s *IAMQueryServer) CheckPermission(
	ctx context.Context,
	req *pbiamv1.CheckPermissionRequest,
) (*pbiamv1.CheckPermissionResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	allowed, err := s.queries.CheckPermissionForResource(
		ctx,
		req.TenantId,
		userID,
		req.Permission,
		req.Resource,
	)
	if err != nil {
		if errors.Is(err, iamdomain.ErrPermissionDenied) || errors.Is(err, iamdomain.ErrInactiveMembership) {
			return &pbiamv1.CheckPermissionResponse{Allowed: false}, nil
		}
		return nil, iamStatusError(err)
	}
	return &pbiamv1.CheckPermissionResponse{Allowed: allowed}, nil
}

func (s *IAMQueryServer) CheckPlatformPermission(
	ctx context.Context,
	req *pbiamv1.CheckPlatformPermissionRequest,
) (*pbiamv1.CheckPermissionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	allowed, err := s.queries.CheckPlatformPermission(ctx, actorUserID, req.Permission)
	if err != nil {
		if errors.Is(err, iamdomain.ErrPermissionDenied) {
			return &pbiamv1.CheckPermissionResponse{Allowed: false}, nil
		}
		return nil, iamStatusError(err)
	}
	return &pbiamv1.CheckPermissionResponse{Allowed: allowed}, nil
}

func (s *IAMQueryServer) SimulateAccess(
	ctx context.Context,
	req *pbiamv1.SimulateAccessRequest,
) (*pbiamv1.SimulateAccessResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	var assumedRole *iamdomain.AssumedRole
	if req.UseAssumedRole && req.AssumedRoleSession != nil && req.AssumedRoleSession.AssumedRoleId != 0 {
		assumedRole = &iamdomain.AssumedRole{
			RoleID:           req.AssumedRoleSession.AssumedRoleId,
			RoleScope:        req.AssumedRoleSession.AssumedRoleScope,
			RoleName:         req.AssumedRoleSession.AssumedRoleName,
			TenantID:         req.AssumedRoleSession.AssumedRoleTenantId,
			ServicePrincipal: req.AssumedRoleSession.AssumedRoleServicePrincipal,
			SessionName:      req.AssumedRoleSession.AssumedRoleSessionName,
			SourceIdentity:   req.AssumedRoleSession.AssumedRoleSourceIdentity,
			SessionTags:      req.AssumedRoleSession.SessionTags,
		}
	}
	result, err := s.queries.SimulateAccess(ctx, iamdomain.SimulateAccessInput{
		Scope:            req.Scope,
		TenantID:         req.TenantId,
		UserID:           userID,
		Action:           req.Action,
		Resource:         req.Resource,
		UseAssumedRole:   req.UseAssumedRole,
		AssumedRole:      assumedRole,
		SessionPolicy:    iammapper.FromPBPolicyStatements(req.SessionPolicy),
		Attributes:       mergeStringMaps(req.Attributes, toPrincipalTagAttributes(req.SessionTags)),
		ServicePrincipal: req.ServicePrincipal,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	return iammapper.ToPBSimulateAccessResponse(result), nil
}
