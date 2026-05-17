package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

const (
	SessionStatusActive  = "active"
	SessionStatusRevoked = "revoked"
)

type Session struct {
	ID                  string                   `json:"id"`
	UserID              uint                     `json:"user_id"`
	ActiveTenantID      string                   `json:"active_tenant_id"`
	SessionPolicy       []SessionPolicyStatement `json:"session_policy,omitempty"`
	AssumedRoleID       uint64                   `json:"assumed_role_id,omitempty"`
	AssumedRoleScope    string                   `json:"assumed_role_scope,omitempty"`
	AssumedRoleName     string                   `json:"assumed_role_name,omitempty"`
	AssumedRoleTenantID string                   `json:"assumed_role_tenant_id,omitempty"`
	Status              string                   `json:"status"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
	ExpiresAt           time.Time                `json:"expires_at"`
	RevokedAt           *time.Time               `json:"revoked_at"`
}

type SessionPolicyStatement struct {
	Effect          string `json:"effect"`
	ActionPattern   string `json:"action_pattern"`
	ResourcePattern string `json:"resource_pattern"`
}

type RefreshToken struct {
	ID                string     `json:"id"`
	SessionID         string     `json:"session_id"`
	TokenHash         string     `json:"token_hash"`
	ExpiresAt         time.Time  `json:"expires_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	RevokedAt         *time.Time `json:"revoked_at"`
	ReplacedByTokenID *string    `json:"replaced_by_token_id"`
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrRefreshTokenInvalid  = errors.New("refresh token invalid")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrInvalidSessionPolicy = errors.New("session policy must include at least one statement")
)
