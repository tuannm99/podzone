package domain

import (
	"context"
	"strings"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
)

func (u *authInteractorImpl) AssumeRole(
	ctx context.Context,
	userID uint,
	accessToken string,
	roleName string,
	tenantID string,
	sessionPolicy []entity.SessionPolicyStatement,
	externalID string,
	sessionName string,
	sourceIdentity string,
	durationSeconds uint32,
	servicePrincipal string,
	sessionTags map[string]string,
) (*inputport.AuthResult, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	session, user, now, err := u.loadOwnedActiveSession(ctx, userID, accessToken)
	if err != nil {
		return nil, err
	}
	assumedRole, err := u.roleAssumer.AssumeRole(ctx, outputport.AssumeRoleInput{
		AccessToken:      accessToken,
		UserID:           userID,
		RoleName:         roleName,
		TenantID:         tenantID,
		ExternalID:       externalID,
		SessionName:      sessionName,
		SourceIdentity:   sourceIdentity,
		DurationSeconds:  durationSeconds,
		ServicePrincipal: strings.TrimSpace(servicePrincipal),
		SessionTags: func() map[string]string {
			if len(sessionTags) > 0 {
				return cloneStringMap(sessionTags)
			}
			return cloneStringMap(session.SessionTags)
		}(),
	})
	if err != nil {
		return nil, err
	}
	normalized := normalizeSessionPolicyStatements(sessionPolicy)
	session.AssumedRoleID = assumedRole.RoleID
	session.AssumedRoleScope = assumedRole.RoleScope
	session.AssumedRoleName = assumedRole.RoleName
	session.AssumedRoleTenantID = assumedRole.TenantID
	session.AssumedRoleServicePrincipal = assumedRole.ServicePrincipal
	session.AssumedRoleSessionName = assumedRole.SessionName
	session.AssumedRoleSourceIdentity = assumedRole.SourceIdentity
	session.SessionTags = cloneStringMap(assumedRole.SessionTags)
	session.AssumedRoleExpiresAt = &assumedRole.ExpiresAt
	session.SessionPolicy = normalized
	if assumedRole.RoleScope == outputport.PolicyScopeTenant {
		session.ActiveTenantID = assumedRole.TenantID
	}
	if err := u.sessionRepository.UpdateSessionPolicy(ctx, session.ID, normalized, now); err != nil {
		return nil, err
	}
	if err := u.sessionRepository.UpdateAssumedRole(ctx, *session, now); err != nil {
		return nil, err
	}
	token, err := u.issueSessionAccessToken(*user, *session)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken: token,
		UserInfo: *user,
	}, nil
}

func (u *authInteractorImpl) ClearAssumedRole(
	ctx context.Context,
	userID uint,
	accessToken string,
) (*inputport.AuthResult, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	session, user, now, err := u.loadOwnedActiveSession(ctx, userID, accessToken)
	if err != nil {
		return nil, err
	}
	session.AssumedRoleID = 0
	session.AssumedRoleScope = ""
	session.AssumedRoleName = ""
	session.AssumedRoleTenantID = ""
	session.AssumedRoleServicePrincipal = ""
	session.AssumedRoleSessionName = ""
	session.AssumedRoleSourceIdentity = ""
	session.AssumedRoleExpiresAt = nil
	session.SessionTags = nil
	if err := u.sessionRepository.UpdateAssumedRole(ctx, *session, now); err != nil {
		return nil, err
	}
	token, err := u.issueSessionAccessToken(*user, *session)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken: token,
		UserInfo: *user,
	}, nil
}
