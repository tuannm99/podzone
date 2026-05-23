package domain

import (
	"context"
	"strings"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
)

func (u *authInteractorImpl) AssumeSessionPolicy(
	ctx context.Context,
	userID uint,
	accessToken string,
	statements []entity.SessionPolicyStatement,
) (*inputport.AuthResult, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	if len(statements) == 0 {
		return nil, entity.ErrInvalidSessionPolicy
	}
	session, user, now, err := u.loadOwnedActiveSession(ctx, userID, accessToken)
	if err != nil {
		return nil, err
	}
	normalized := normalizeSessionPolicyStatements(statements)
	if len(normalized) == 0 {
		return nil, entity.ErrInvalidSessionPolicy
	}
	if err := u.sessionRepository.UpdateSessionPolicy(ctx, session.ID, normalized, now); err != nil {
		return nil, err
	}
	session.SessionPolicy = normalized
	token, err := u.issueSessionAccessToken(*user, *session)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken: token,
		UserInfo: *user,
	}, nil
}

func (u *authInteractorImpl) ClearSessionPolicy(
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
	if err := u.sessionRepository.UpdateSessionPolicy(ctx, session.ID, nil, now); err != nil {
		return nil, err
	}
	session.SessionPolicy = nil
	token, err := u.issueSessionAccessToken(*user, *session)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken: token,
		UserInfo: *user,
	}, nil
}

func normalizeSessionPolicyStatements(
	statements []entity.SessionPolicyStatement,
) []entity.SessionPolicyStatement {
	out := make([]entity.SessionPolicyStatement, 0, len(statements))
	for _, statement := range statements {
		effect := strings.ToLower(strings.TrimSpace(statement.Effect))
		if effect == "" {
			effect = "allow"
		}
		action := strings.TrimSpace(statement.ActionPattern)
		if action == "" {
			continue
		}
		resource := strings.TrimSpace(statement.ResourcePattern)
		if resource == "" {
			resource = "*"
		}
		out = append(out, entity.SessionPolicyStatement{
			Effect:          effect,
			ActionPattern:   action,
			ResourcePattern: resource,
		})
	}
	return out
}
