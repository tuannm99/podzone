package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/tuannm99/podzone/pkg/pdauthn"
)

const (
	SessionStatusActive  = "active"
	SessionStatusRevoked = "revoked"
)

type Session struct {
	ID                          string                   `json:"id"`
	UserID                      uint                     `json:"user_id"`
	ActiveTenantID              string                   `json:"active_tenant_id"`
	SessionPolicy               []SessionPolicyStatement `json:"session_policy,omitempty"`
	SessionTags                 map[string]string        `json:"session_tags,omitempty"`
	AssumedRoleID               uint64                   `json:"assumed_role_id,omitempty"`
	AssumedRoleScope            string                   `json:"assumed_role_scope,omitempty"`
	AssumedRoleName             string                   `json:"assumed_role_name,omitempty"`
	AssumedRoleTenantID         string                   `json:"assumed_role_tenant_id,omitempty"`
	AssumedRoleServicePrincipal string                   `json:"assumed_role_service_principal,omitempty"`
	AssumedRoleSessionName      string                   `json:"assumed_role_session_name,omitempty"`
	AssumedRoleSourceIdentity   string                   `json:"assumed_role_source_identity,omitempty"`
	AssumedRoleExpiresAt        *time.Time               `json:"assumed_role_expires_at,omitempty"`
	Status                      string                   `json:"status"`
	CreatedAt                   time.Time                `json:"created_at"`
	UpdatedAt                   time.Time                `json:"updated_at"`
	ExpiresAt                   time.Time                `json:"expires_at"`
	RevokedAt                   *time.Time               `json:"revoked_at"`
}

type SessionPolicyStatement = pdauthn.PolicyStatement

type SessionPolicyCondition = pdauthn.PolicyCondition

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
