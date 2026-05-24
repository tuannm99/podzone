package interactor

import entity "github.com/tuannm99/podzone/internal/iam/entity"

type (
	AccessRequest                    = entity.AccessRequest
	AssumedRole                      = entity.AssumedRole
	CreateGroupInput                 = entity.CreateGroupInput
	CreatePolicyInput                = entity.CreatePolicyInput
	CreatePolicyVersionInput         = entity.CreatePolicyVersionInput
	Group                            = entity.Group
	GroupInlinePolicy                = entity.GroupInlinePolicy
	Membership                       = entity.Membership
	Organization                     = entity.Organization
	PermissionBoundary               = entity.PermissionBoundary
	PlatformMembership               = entity.PlatformMembership
	Policy                           = entity.Policy
	PolicyAttachment                 = entity.PolicyAttachment
	PolicyCondition                  = entity.PolicyCondition
	PolicyStatement                  = entity.PolicyStatement
	PolicyVersion                    = entity.PolicyVersion
	PutGroupInlinePolicyInput        = entity.PutGroupInlinePolicyInput
	PutPlatformUserInlinePolicyInput = entity.PutPlatformUserInlinePolicyInput
	PutRoleTrustPolicyInput          = entity.PutRoleTrustPolicyInput
	PutTenantUserInlinePolicyInput   = entity.PutTenantUserInlinePolicyInput
	RolePermissionBoundary           = entity.RolePermissionBoundary
	RoleTrustStatement               = entity.RoleTrustStatement
	SimulateAccessInput              = entity.SimulateAccessInput
	SimulateDecisionLayer            = entity.SimulateDecisionLayer
	SimulateAccessResult             = entity.SimulateAccessResult
	SimulateMatchedStatement         = entity.SimulateMatchedStatement
	TenantInvite                     = entity.TenantInvite
	UserInlinePolicy                 = entity.UserInlinePolicy
)

const (
	ConditionBool                     = entity.ConditionBool
	ConditionDateGreaterThan          = entity.ConditionDateGreaterThan
	ConditionDateLessThan             = entity.ConditionDateLessThan
	ConditionIpAddress                = entity.ConditionIpAddress
	ConditionNull                     = entity.ConditionNull
	ConditionNumericEquals            = entity.ConditionNumericEquals
	ConditionNumericGreaterThanEquals = entity.ConditionNumericGreaterThanEquals
	ConditionNumericLessThanEquals    = entity.ConditionNumericLessThanEquals
	ConditionStringEquals             = entity.ConditionStringEquals
	ConditionStringLike               = entity.ConditionStringLike
	ConditionStringNotEquals          = entity.ConditionStringNotEquals
	ConditionStringNotLike            = entity.ConditionStringNotLike
	InviteStatusAccepted              = entity.InviteStatusAccepted
	InviteStatusPending               = entity.InviteStatusPending
	InviteStatusRevoked               = entity.InviteStatusRevoked
	MembershipStatusActive            = entity.MembershipStatusActive
	PolicyEffectAllow                 = entity.PolicyEffectAllow
	PolicyEffectDeny                  = entity.PolicyEffectDeny
	PolicyScopePlatform               = entity.PolicyScopePlatform
	PolicyScopeTenant                 = entity.PolicyScopeTenant
	TrustPrincipalPlatformRole        = entity.TrustPrincipalPlatformRole
	TrustPrincipalService             = entity.TrustPrincipalService
	TrustPrincipalTenantRole          = entity.TrustPrincipalTenantRole
	TrustPrincipalUser                = entity.TrustPrincipalUser
)

var (
	ErrAssumeRoleDenied        = entity.ErrAssumeRoleDenied
	ErrGroupNotFound           = entity.ErrGroupNotFound
	ErrImmutableGroup          = entity.ErrImmutableGroup
	ErrInactiveMembership      = entity.ErrInactiveMembership
	ErrInvalidAssumeRole       = entity.ErrInvalidAssumeRole
	ErrInvalidInviteEmail      = entity.ErrInvalidInviteEmail
	ErrInvalidInviteToken      = entity.ErrInvalidInviteToken
	ErrInviteAccepted          = entity.ErrInviteAccepted
	ErrInvalidOrganizationName = entity.ErrInvalidOrganizationName
	ErrInvalidOrganizationSlug = entity.ErrInvalidOrganizationSlug
	ErrInvalidPolicyName       = entity.ErrInvalidPolicyName
	ErrInvalidRoleName         = entity.ErrInvalidRoleName
	ErrInvalidServicePrincipal = entity.ErrInvalidServicePrincipal
	ErrInvalidTenantName       = entity.ErrInvalidTenantName
	ErrInvalidTenantSlug       = entity.ErrInvalidTenantSlug
	ErrInvalidUserID           = entity.ErrInvalidUserID
	ErrInviteEmailMismatch     = entity.ErrInviteEmailMismatch
	ErrInviteExpired           = entity.ErrInviteExpired
	ErrInviteNotFound          = entity.ErrInviteNotFound
	ErrInviteRevoked           = entity.ErrInviteRevoked
	ErrOrganizationNotFound    = entity.ErrOrganizationNotFound
	ErrPermissionDenied        = entity.ErrPermissionDenied
	ErrPolicyNotFound          = entity.ErrPolicyNotFound
	ErrPolicyVersionNotFound   = entity.ErrPolicyVersionNotFound
	ErrRoleNotFound            = entity.ErrRoleNotFound
	ErrTenantNotFound          = entity.ErrTenantNotFound
	GetAssumedRole             = entity.GetAssumedRole
	GetSessionPolicyStatements = entity.GetSessionPolicyStatements
	HashInviteToken            = entity.HashInviteToken
	NewInviteToken             = entity.NewInviteToken
	NormalizeInviteEmail       = entity.NormalizeInviteEmail
)
