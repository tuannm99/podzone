package model

import (
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
)

type Session struct {
	ID             string     `db:"id"`
	UserID         uint       `db:"user_id"`
	ActiveTenantID string     `db:"active_tenant_id"`
	Status         string     `db:"status"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	ExpiresAt      time.Time  `db:"expires_at"`
	RevokedAt      *time.Time `db:"revoked_at"`
}

func (s Session) ToEntity() *entity.Session {
	return &entity.Session{
		ID:             s.ID,
		UserID:         s.UserID,
		ActiveTenantID: s.ActiveTenantID,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
		ExpiresAt:      s.ExpiresAt,
		RevokedAt:      s.RevokedAt,
	}
}

type RefreshToken struct {
	ID                string     `db:"id"`
	SessionID         string     `db:"session_id"`
	TokenHash         string     `db:"token_hash"`
	ExpiresAt         time.Time  `db:"expires_at"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
	RevokedAt         *time.Time `db:"revoked_at"`
	ReplacedByTokenID *string    `db:"replaced_by_token_id"`
}

func (t RefreshToken) ToEntity() *entity.RefreshToken {
	return &entity.RefreshToken{
		ID:                t.ID,
		SessionID:         t.SessionID,
		TokenHash:         t.TokenHash,
		ExpiresAt:         t.ExpiresAt,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
		RevokedAt:         t.RevokedAt,
		ReplacedByTokenID: t.ReplacedByTokenID,
	}
}
