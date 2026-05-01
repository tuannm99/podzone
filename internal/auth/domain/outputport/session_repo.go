package outputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
)

type SessionRepository interface {
	Create(ctx context.Context, session entity.Session) error
	GetByID(ctx context.Context, id string) (*entity.Session, error)
	ListByUser(ctx context.Context, userID uint) ([]entity.Session, error)
	UpdateActiveTenant(ctx context.Context, id, tenantID string, updatedAt time.Time) error
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token entity.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time, replacedByTokenID *string) error
	RevokeBySession(ctx context.Context, sessionID string, revokedAt time.Time) error
}
