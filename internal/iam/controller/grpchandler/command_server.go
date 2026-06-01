package grpchandler

import (
	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	iaminputport "github.com/tuannm99/podzone/internal/iam/domain/inputport"
	iamoutputport "github.com/tuannm99/podzone/internal/iam/domain/outputport"
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

var _ pbiamv1.IAMCommandServiceServer = (*IAMCommandServer)(nil)

type IAMCommandServer struct {
	pbiamv1.UnimplementedIAMCommandServiceServer
	*iamHandlerBase
	commands iaminputport.IAMCommandUsecase
	queries  iaminputport.IAMQueryUsecase
}

func NewIAMCommandServer(
	commands iaminputport.IAMCommandUsecase,
	queries iaminputport.IAMQueryUsecase,
	auditRep iamoutputport.AuditLogRepository,
	userDirectory iamoutputport.UserDirectory,
	cfg iamconfig.ServerConfig,
) *IAMCommandServer {
	return &IAMCommandServer{
		iamHandlerBase: newIAMHandlerBase(auditRep, userDirectory, cfg),
		commands:       commands,
		queries:        queries,
	}
}
