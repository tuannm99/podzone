package iamprojection

import (
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
)

func NewHandler(repo outputport.IAMProjectionRepository) (messaging.Handler, error) {
	return messaging.NewRegistry(
		NewTenantCreatedHandler(repo),
		NewTenantMemberAddedHandler(repo),
	)
}
