package grpchandler

import (
	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	iamoutputport "github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/pkg/pdauthn"
)

type iamHandlerBase struct {
	auditRep       iamoutputport.AuditLogRepository
	userDirectory  iamoutputport.UserDirectory
	appRedirectURL string
	verifier       *pdauthn.Verifier
}

func newIAMHandlerBase(
	auditRep iamoutputport.AuditLogRepository,
	userDirectory iamoutputport.UserDirectory,
	cfg iamconfig.ServerConfig,
) *iamHandlerBase {
	return &iamHandlerBase{
		auditRep:       auditRep,
		userDirectory:  userDirectory,
		appRedirectURL: cfg.AppRedirectURL,
		verifier:       pdauthn.NewVerifier(cfg.Authn),
	}
}
