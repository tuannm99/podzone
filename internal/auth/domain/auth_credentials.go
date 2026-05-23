package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
)

func (u *authInteractorImpl) Login(ctx context.Context, username, password string) (*inputport.AuthResult, error) {
	user, err := u.userRepository.GetByUsernameOrEmail(username)
	if err != nil {
		return nil, err
	}

	err = entity.CheckPassword(user.Password, password)
	if err != nil {
		return nil, err
	}

	return u.newSessionAuthResult(ctx, user, "")
}

func (u *authInteractorImpl) Register(ctx context.Context, req inputport.RegisterCmd) (*inputport.AuthResult, error) {
	user, err := u.userRepository.Create(
		entity.User{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
		},
	)
	if err != nil {
		return nil, err
	}

	err = u.userRepository.UpdateById(user.Id, entity.User{InitialFrom: "podzone"})
	if err != nil {
		return nil, err
	}

	return u.newSessionAuthResult(ctx, user, "")
}

func (u *authInteractorImpl) SwitchActiveTenant(
	ctx context.Context,
	userID uint,
	tenantID, accessToken string,
) (*inputport.AuthResult, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}

	if err := u.tenantAccessChecker.EnsureActiveMembership(ctx, tenantID, userID); err != nil {
		return nil, err
	}

	user, err := u.userRepository.GetByID(fmt.Sprintf("%d", userID))
	if err != nil {
		return nil, err
	}

	session, err := u.sessionFromAccessToken(accessToken)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, entity.ErrSessionRevoked
	}
	now := time.Now().UTC()
	if session.Status != entity.SessionStatusActive || session.RevokedAt != nil || now.After(session.ExpiresAt) {
		return nil, entity.ErrSessionRevoked
	}
	if err := u.sessionRepository.UpdateActiveTenant(ctx, session.ID, tenantID, now); err != nil {
		return nil, err
	}

	session.ActiveTenantID = tenantID
	token, err := u.issueSessionAccessToken(*user, *session)
	if err != nil {
		return nil, err
	}

	return &inputport.AuthResult{
		JwtToken: token,
		UserInfo: *user,
	}, nil
}

func (u *authInteractorImpl) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (*inputport.AuthResult, error) {
	hashed := entity.HashToken(refreshToken)
	stored, err := u.refreshTokenRepo.GetByTokenHash(ctx, hashed)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if stored.RevokedAt != nil {
		return nil, entity.ErrRefreshTokenInvalid
	}
	if now.After(stored.ExpiresAt) {
		return nil, entity.ErrRefreshTokenExpired
	}

	session, err := u.sessionRepository.GetByID(ctx, stored.SessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.SessionStatusActive || session.RevokedAt != nil || now.After(session.ExpiresAt) {
		return nil, entity.ErrSessionRevoked
	}

	user, err := u.userRepository.GetByID(fmt.Sprintf("%d", session.UserID))
	if err != nil {
		return nil, err
	}

	newRefreshToken, refreshEntity, err := u.newRefreshToken(session.ID, now)
	if err != nil {
		return nil, err
	}
	if err := u.refreshTokenRepo.Revoke(ctx, stored.ID, now, &refreshEntity.ID); err != nil {
		return nil, err
	}
	if err := u.refreshTokenRepo.Create(ctx, refreshEntity); err != nil {
		return nil, err
	}
	accessToken, err := u.issueSessionAccessToken(*user, *session)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken:     accessToken,
		RefreshToken: newRefreshToken,
		UserInfo:     *user,
	}, nil
}

func (u *authInteractorImpl) Logout(ctx context.Context, accessToken string) (string, error) {
	session, err := u.sessionFromAccessToken(accessToken)
	if err != nil {
		return "/", err
	}
	now := time.Now().UTC()
	if err := u.sessionRepository.Revoke(ctx, session.ID, now); err != nil {
		return "/", err
	}
	if err := u.refreshTokenRepo.RevokeBySession(ctx, session.ID, now); err != nil {
		return "/", err
	}
	return "/", nil
}
