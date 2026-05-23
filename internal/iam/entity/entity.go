package entity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

const (
	RolePlatformOwner = "platform_owner"
	RolePlatformAdmin = "platform_admin"

	RoleTenantOwner  = "tenant_owner"
	RoleTenantAdmin  = "tenant_admin"
	RoleTenantEditor = "tenant_editor"
	RoleTenantViewer = "tenant_viewer"

	PolicyScopePlatform = "platform"
	PolicyScopeTenant   = "tenant"

	PolicyEffectAllow = "allow"
	PolicyEffectDeny  = "deny"

	ConditionStringEquals             = "StringEquals"
	ConditionStringLike               = "StringLike"
	ConditionBool                     = "Bool"
	ConditionStringNotEquals          = "StringNotEquals"
	ConditionStringNotLike            = "StringNotLike"
	ConditionNumericEquals            = "NumericEquals"
	ConditionNumericGreaterThanEquals = "NumericGreaterThanEquals"
	ConditionNumericLessThanEquals    = "NumericLessThanEquals"
	ConditionDateGreaterThan          = "DateGreaterThan"
	ConditionDateLessThan             = "DateLessThan"
	ConditionIpAddress                = "IpAddress"
	ConditionNull                     = "Null"

	TrustPrincipalUser         = "user"
	TrustPrincipalPlatformRole = "platform_role"
	TrustPrincipalTenantRole   = "tenant_role"
	TrustPrincipalService      = "service"

	MembershipStatusActive = "active"

	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRevoked  = "revoked"
)

var (
	ErrTenantNotFound          = errors.New("iam: tenant not found")
	ErrRoleNotFound            = errors.New("iam: role not found")
	ErrPolicyNotFound          = errors.New("iam: policy not found")
	ErrGroupNotFound           = errors.New("iam: group not found")
	ErrMembershipNotFound      = errors.New("iam: membership not found")
	ErrPermissionDenied        = errors.New("iam: permission denied")
	ErrTenantSlugTaken         = errors.New("iam: tenant slug already exists")
	ErrInvalidTenantName       = errors.New("iam: tenant name is required")
	ErrInvalidTenantSlug       = errors.New("iam: tenant slug is required")
	ErrInvalidUserID           = errors.New("iam: user id is required")
	ErrInvalidRoleName         = errors.New("iam: role name is required")
	ErrInvalidInviteEmail      = errors.New("iam: invite email is required")
	ErrInvalidInviteToken      = errors.New("iam: invite token is required")
	ErrInactiveMembership      = errors.New("iam: membership is inactive")
	ErrInviteNotFound          = errors.New("iam: invite not found")
	ErrInviteExpired           = errors.New("iam: invite expired")
	ErrInviteRevoked           = errors.New("iam: invite revoked")
	ErrInviteAccepted          = errors.New("iam: invite already accepted")
	ErrInviteEmailMismatch     = errors.New("iam: invite email does not match authenticated user")
	ErrImmutablePolicy         = errors.New("iam: managed/system policy cannot be deleted")
	ErrImmutableGroup          = errors.New("iam: system group cannot be deleted")
	ErrPolicyInUse             = errors.New("iam: policy is still attached")
	ErrInvalidPolicyName       = errors.New("iam: policy name is required")
	ErrPolicyVersionNotFound   = errors.New("iam: policy version not found")
	ErrDefaultPolicyVersion    = errors.New("iam: default policy version cannot be deleted")
	ErrInvalidAssumeRole       = errors.New("iam: invalid assume role target")
	ErrAssumeRoleDenied        = errors.New("iam: assume role denied")
	ErrInvalidServicePrincipal = errors.New("iam: invalid service principal")
	ErrInvalidPolicyStatement  = errors.New("iam: invalid policy statement")
)

func ResourceTenant(tenantID string) string {
	return "podzone:tenant/" + strings.TrimSpace(tenantID)
}

func ResourceStore(tenantID, storeID string) string {
	return ResourceTenant(tenantID) + "/store/" + strings.TrimSpace(storeID)
}

func ResourceOrder(tenantID, storeID, orderID string) string {
	return ResourceStore(tenantID, storeID) + "/order/" + strings.TrimSpace(orderID)
}

func ResourcePartner(partnerID string) string {
	return "podzone:partner/" + strings.TrimSpace(partnerID)
}

func NormalizeInviteEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func HashInviteToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func NewInviteToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
