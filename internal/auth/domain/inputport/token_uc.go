package inputport

import "github.com/tuannm99/podzone/internal/auth/domain/entity"

type TokenUsecase interface {
	CreateJwtToken(user entity.User) (string, error)
	CreateJwtTokenForTenant(user entity.User, activeTenantID string) (string, error)
	CreateJwtTokenForSession(user entity.User, activeTenantID, sessionID string) (string, error)
}
