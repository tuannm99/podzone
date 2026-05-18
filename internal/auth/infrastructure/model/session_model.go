package model

import (
	"encoding/json"
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
)

type Session struct {
	ID                          string     `db:"id"`
	UserID                      uint       `db:"user_id"`
	ActiveTenantID              string     `db:"active_tenant_id"`
	SessionPolicyJSON           string     `db:"session_policy_json"`
	SessionTagsJSON             string     `db:"session_tags_json"`
	AssumedRoleID               uint64     `db:"assumed_role_id"`
	AssumedRoleScope            string     `db:"assumed_role_scope"`
	AssumedRoleName             string     `db:"assumed_role_name"`
	AssumedRoleTenantID         string     `db:"assumed_role_tenant_id"`
	AssumedRoleServicePrincipal string     `db:"assumed_role_service_principal"`
	AssumedRoleSessionName      string     `db:"assumed_role_session_name"`
	AssumedRoleSourceIdentity   string     `db:"assumed_role_source_identity"`
	AssumedRoleExpiresAt        *time.Time `db:"assumed_role_expires_at"`
	Status                      string     `db:"status"`
	CreatedAt                   time.Time  `db:"created_at"`
	UpdatedAt                   time.Time  `db:"updated_at"`
	ExpiresAt                   time.Time  `db:"expires_at"`
	RevokedAt                   *time.Time `db:"revoked_at"`
}

func (s Session) ToEntity() *entity.Session {
	statements := make([]entity.SessionPolicyStatement, 0)
	if s.SessionPolicyJSON != "" {
		_ = json.Unmarshal([]byte(s.SessionPolicyJSON), &statements)
	}
	tags := map[string]string{}
	if s.SessionTagsJSON != "" {
		_ = json.Unmarshal([]byte(s.SessionTagsJSON), &tags)
	}
	return &entity.Session{
		ID:                          s.ID,
		UserID:                      s.UserID,
		ActiveTenantID:              s.ActiveTenantID,
		SessionPolicy:               statements,
		SessionTags:                 tags,
		AssumedRoleID:               s.AssumedRoleID,
		AssumedRoleScope:            s.AssumedRoleScope,
		AssumedRoleName:             s.AssumedRoleName,
		AssumedRoleTenantID:         s.AssumedRoleTenantID,
		AssumedRoleServicePrincipal: s.AssumedRoleServicePrincipal,
		AssumedRoleSessionName:      s.AssumedRoleSessionName,
		AssumedRoleSourceIdentity:   s.AssumedRoleSourceIdentity,
		AssumedRoleExpiresAt:        s.AssumedRoleExpiresAt,
		Status:                      s.Status,
		CreatedAt:                   s.CreatedAt,
		UpdatedAt:                   s.UpdatedAt,
		ExpiresAt:                   s.ExpiresAt,
		RevokedAt:                   s.RevokedAt,
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
