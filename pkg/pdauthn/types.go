package pdauthn

import "github.com/golang-jwt/jwt/v5"

type PolicyStatement struct {
	Effect          string            `json:"effect"`
	ActionPattern   string            `json:"action_pattern"`
	ResourcePattern string            `json:"resource_pattern"`
	Conditions      []PolicyCondition `json:"conditions,omitempty"`
}

type PolicyCondition struct {
	Operator string `json:"operator"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type Claims struct {
	UserID                      uint              `json:"user_id"`
	Email                       string            `json:"email"`
	Username                    string            `json:"user_name"`
	IdentitySource              string            `json:"identity_source,omitempty"`
	ActiveTenantID              string            `json:"active_tenant_id,omitempty"`
	SessionID                   string            `json:"session_id,omitempty"`
	SessionPolicy               []PolicyStatement `json:"session_policy,omitempty"`
	SessionTags                 map[string]string `json:"session_tags,omitempty"`
	AssumedRoleID               uint64            `json:"assumed_role_id,omitempty"`
	AssumedRoleScope            string            `json:"assumed_role_scope,omitempty"`
	AssumedRoleName             string            `json:"assumed_role_name,omitempty"`
	AssumedRoleTenantID         string            `json:"assumed_role_tenant_id,omitempty"`
	AssumedRoleServicePrincipal string            `json:"assumed_role_service_principal,omitempty"`
	AssumedRoleSessionName      string            `json:"assumed_role_session_name,omitempty"`
	AssumedRoleSourceIdentity   string            `json:"assumed_role_source_identity,omitempty"`
	Key                         string            `json:"key"`
	jwt.RegisteredClaims
}

type Config struct {
	JWTSecret string
	JWTKey    string
}
