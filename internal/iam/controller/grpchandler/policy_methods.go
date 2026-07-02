package grpchandler

import (
	"context"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

func (s *IAMCommandServer) CreatePolicy(
	ctx context.Context,
	req *pbiamv1.CreatePolicyRequest,
) (*pbiamv1.CreatePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	policy, statements, err := s.commands.CreatePolicy(ctx, iamdomain.CreatePolicyInput{
		Scope:       req.Scope,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.created", "iam_policy", policy.Name, "", map[string]any{
		"scope":      policy.Scope,
		"statements": len(statements),
	})
	return &pbiamv1.CreatePolicyResponse{
		Policy:     iammapper.ToPBPolicy(policy),
		Statements: iammapper.ToPBPolicyStatements(statements),
	}, nil
}

func (s *IAMCommandServer) CreatePolicyVersion(
	ctx context.Context,
	req *pbiamv1.CreatePolicyVersionRequest,
) (*pbiamv1.CreatePolicyVersionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	version, statements, err := s.commands.CreatePolicyVersion(ctx, iamdomain.CreatePolicyVersionInput{
		PolicyName:   req.Name,
		Statements:   iammapper.FromPBPolicyStatements(req.Statements),
		SetAsDefault: req.SetAsDefault,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.version.created", "iam_policy_version", req.Name, "", map[string]any{
		"version":        version.Version,
		"set_as_default": req.SetAsDefault,
		"statements":     len(statements),
	})
	return &pbiamv1.CreatePolicyVersionResponse{
		PolicyVersion: iammapper.ToPBPolicyVersion(version),
		Statements:    iammapper.ToPBPolicyStatements(statements),
	}, nil
}

func (s *IAMCommandServer) SetDefaultPolicyVersion(
	ctx context.Context,
	req *pbiamv1.SetDefaultPolicyVersionRequest,
) (*pbiamv1.SetDefaultPolicyVersionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.SetDefaultPolicyVersion(ctx, req.Name, req.Version); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.policy.version.set_default",
		"iam_policy_version",
		req.Name,
		"",
		map[string]any{
			"version": req.Version,
		},
	)
	return &pbiamv1.SetDefaultPolicyVersionResponse{}, nil
}

func (s *IAMCommandServer) DeletePolicyVersion(
	ctx context.Context,
	req *pbiamv1.DeletePolicyVersionRequest,
) (*pbiamv1.DeletePolicyVersionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeletePolicyVersion(ctx, req.Name, req.Version); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.version.deleted", "iam_policy_version", req.Name, "", map[string]any{
		"version": req.Version,
	})
	return &pbiamv1.DeletePolicyVersionResponse{}, nil
}

func (s *IAMCommandServer) DeletePolicy(
	ctx context.Context,
	req *pbiamv1.DeletePolicyRequest,
) (*pbiamv1.DeletePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeletePolicy(ctx, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.deleted", "iam_policy", req.Name, "", map[string]any{
		"policy_name": req.Name,
	})
	return &pbiamv1.DeletePolicyResponse{}, nil
}

func (s *IAMQueryServer) GetPolicy(
	ctx context.Context,
	req *pbiamv1.GetPolicyRequest,
) (*pbiamv1.GetPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	policy, statements, err := s.queries.GetPolicy(ctx, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetPolicyResponse{
		Policy:     iammapper.ToPBPolicy(policy),
		Statements: iammapper.ToPBPolicyStatements(statements),
	}, nil
}

func (s *IAMQueryServer) ListPolicyVersions(
	ctx context.Context,
	req *pbiamv1.ListPolicyVersionsRequest,
) (*pbiamv1.ListPolicyVersionsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListPolicyVersions(ctx, req.Name, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.PolicyVersion, 0, len(page.Items))
	for i := range page.Items {
		out = append(out, iammapper.ToPBPolicyVersion(&page.Items[i]))
	}
	return &pbiamv1.ListPolicyVersionsResponse{
		Versions: out,
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) ListPolicies(
	ctx context.Context,
	req *pbiamv1.ListPoliciesRequest,
) (*pbiamv1.ListPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListPolicies(ctx, req.Scope, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.ListPoliciesResponse{
		Policies: iammapper.ToPBPolicies(page.Items),
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) ListPolicyAttachments(
	ctx context.Context,
	req *pbiamv1.ListPolicyAttachmentsRequest,
) (*pbiamv1.ListPolicyAttachmentsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListPolicyAttachments(ctx, req.Name, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.ListPolicyAttachmentsResponse{
		Attachments: iammapper.ToPBPolicyAttachments(page.Items),
		PageInfo:    iammapper.ToPBPageInfo(page),
	}, nil
}
