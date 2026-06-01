package grpchandler

import (
	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	iaminputport "github.com/tuannm99/podzone/internal/iam/domain/inputport"
	iamoutputport "github.com/tuannm99/podzone/internal/iam/domain/outputport"
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

var _ pbiamv1.IAMQueryServiceServer = (*IAMQueryServer)(nil)

type IAMQueryServer struct {
	pbiamv1.UnimplementedIAMQueryServiceServer
	*iamHandlerBase
	queries iaminputport.IAMQueryUsecase
}

func NewIAMQueryServer(
	queries iaminputport.IAMQueryUsecase,
	auditRep iamoutputport.AuditLogRepository,
	userDirectory iamoutputport.UserDirectory,
	cfg iamconfig.ServerConfig,
) *IAMQueryServer {
	return &IAMQueryServer{
		iamHandlerBase: newIAMHandlerBase(auditRep, userDirectory, cfg),
		queries:        queries,
	}
}
