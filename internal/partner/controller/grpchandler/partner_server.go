package grpchandler

import (
	"context"
	"errors"

	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
	pbpartnerv1 "github.com/tuannm99/podzone/pkg/api/proto/partner/v1"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PartnerServer struct {
	pbpartnerv1.UnimplementedPartnerServiceServer
	uc         partnerdomain.PartnerUsecase
	authorizer TenantAuthorizer
}

func NewPartnerServer(uc partnerdomain.PartnerUsecase, authorizer TenantAuthorizer) *PartnerServer {
	return &PartnerServer{uc: uc, authorizer: authorizer}
}

func (s *PartnerServer) CreatePartner(
	ctx context.Context,
	req *pbpartnerv1.CreatePartnerRequest,
) (*pbpartnerv1.Partner, error) {
	if _, err := s.authorizer.AuthorizeTenant(ctx, req.TenantId, "partner:manage"); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	out, err := s.uc.CreatePartner(ctx, partnerdomain.CreatePartnerCmd{
		TenantID:     req.TenantId,
		Code:         req.Code,
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		Notes:        req.Notes,
		PartnerType:  req.PartnerType,
	})
	if err != nil {
		return nil, partnerStatusError(err)
	}
	return toProtoPartner(out)
}

func (s *PartnerServer) GetPartner(
	ctx context.Context,
	req *pbpartnerv1.GetPartnerRequest,
) (*pbpartnerv1.Partner, error) {
	out, err := s.uc.GetPartner(ctx, req.Id)
	if err != nil {
		return nil, partnerStatusError(err)
	}
	if _, err := s.authorizer.AuthorizeTenant(ctx, out.TenantID, "partner:read"); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return toProtoPartner(out)
}

func (s *PartnerServer) ListPartners(
	ctx context.Context,
	req *pbpartnerv1.ListPartnersRequest,
) (*pbpartnerv1.ListPartnersResponse, error) {
	if _, err := s.authorizer.AuthorizeTenant(ctx, req.TenantId, "partner:read"); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	items, err := s.uc.ListPartners(ctx, partnerdomain.ListPartnersQuery{
		TenantID:    req.TenantId,
		Status:      req.Status,
		PartnerType: req.PartnerType,
	})
	if err != nil {
		return nil, partnerStatusError(err)
	}
	partners := make([]*pbpartnerv1.Partner, 0, len(items))
	for i := range items {
		mapped, err := toProtoPartner(&items[i])
		if err != nil {
			return nil, err
		}
		partners = append(partners, mapped)
	}
	return &pbpartnerv1.ListPartnersResponse{Partners: partners}, nil
}

func (s *PartnerServer) UpdatePartner(
	ctx context.Context,
	req *pbpartnerv1.UpdatePartnerRequest,
) (*pbpartnerv1.Partner, error) {
	current, err := s.uc.GetPartner(ctx, req.Id)
	if err != nil {
		return nil, partnerStatusError(err)
	}
	if _, err := s.authorizer.AuthorizeTenant(ctx, current.TenantID, "partner:manage"); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	out, err := s.uc.UpdatePartner(ctx, partnerdomain.UpdatePartnerCmd{
		ID:           req.Id,
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		Notes:        req.Notes,
		PartnerType:  req.PartnerType,
	})
	if err != nil {
		return nil, partnerStatusError(err)
	}
	return toProtoPartner(out)
}

func (s *PartnerServer) UpdatePartnerStatus(
	ctx context.Context,
	req *pbpartnerv1.UpdatePartnerStatusRequest,
) (*pbpartnerv1.Partner, error) {
	current, err := s.uc.GetPartner(ctx, req.Id)
	if err != nil {
		return nil, partnerStatusError(err)
	}
	if _, err := s.authorizer.AuthorizeTenant(ctx, current.TenantID, "partner:manage"); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	out, err := s.uc.UpdatePartnerStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, partnerStatusError(err)
	}
	return toProtoPartner(out)
}

func partnerStatusError(err error) error {
	switch {
	case errors.Is(err, partnerdomain.ErrPartnerNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, partnerdomain.ErrPartnerCodeTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, partnerdomain.ErrInvalidPartnerID),
		errors.Is(err, partnerdomain.ErrInvalidPartnerCode),
		errors.Is(err, partnerdomain.ErrInvalidPartnerName),
		errors.Is(err, partnerdomain.ErrInvalidTenantID),
		errors.Is(err, partnerdomain.ErrInvalidPartnerType),
		errors.Is(err, partnerdomain.ErrInvalidPartnerStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProtoPartner(in *partnerdomain.Partner) (*pbpartnerv1.Partner, error) {
	out, err := toolkit.MapStruct[partnerdomain.Partner, pbpartnerv1.Partner](*in)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return out, nil
}
