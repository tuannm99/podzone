package grpchandler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	inputport "github.com/tuannm99/podzone/internal/auth/domain/inputport"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

func TestGoogleLogin_OK(t *testing.T) {
	srv, authUC, _, _ := newAuthServer(t)
	authUC.EXPECT().GenerateOAuthURL(mock.Anything).Return("https://accounts.google.com/auth?state=xyz", nil)

	res, err := srv.GoogleLogin(context.Background(), &pbauthv1.GoogleLoginRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "https://accounts.google.com/auth?state=xyz", res.RedirectUrl)
}

func TestLogin_OK(t *testing.T) {
	srv, authUC, _, _ := newAuthServer(t)
	authUC.EXPECT().Login(mock.Anything, "neo", "TheOne!").Return(&inputport.AuthResult{
		JwtToken: "jwt-token",
		UserInfo: entity.User{Id: 7, Username: "neo", Email: "neo@mx.io"},
	}, nil)

	res, err := srv.Login(context.Background(), &pbauthv1.LoginRequest{
		Username: "neo",
		Password: "TheOne!",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-token", res.JwtToken)
	assert.Equal(t, "neo@mx.io", res.UserInfo.Email)
}

func TestSwitchActiveTenant_OK(t *testing.T) {
	srv, authUC, _, auditRepo := newAuthServer(t)
	expectAuditMaybe(auditRepo)
	authUC.EXPECT().
		SwitchActiveTenant(mock.Anything, uint(7), "tenant-1", "access-token").
		Return(&inputport.AuthResult{
			JwtToken: "jwt-switched",
			UserInfo: entity.User{Id: 7, Username: "neo", Email: "neo@mx.io"},
		}, nil)

	res, err := srv.SwitchActiveTenant(authContextForUser(t, 7), &pbauthv1.SwitchActiveTenantRequest{
		UserId:      7,
		TenantId:    "tenant-1",
		AccessToken: "access-token",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-switched", res.JwtToken)
}

func TestAssumeSessionPolicy_OK(t *testing.T) {
	srv, authUC, sessionRepo, _ := newAuthServer(t)
	sessionRepo.EXPECT().GetByID(mock.Anything, "session-1").Return(sessionWithID("session-1", 7), nil)
	accessToken := accessTokenForSession(t, entity.User{Id: 7}, "", "session-1")
	authUC.EXPECT().
		AssumeSessionPolicy(mock.Anything, uint(7), accessToken, mock.Anything).
		Return(&inputport.AuthResult{JwtToken: "jwt-scoped"}, nil)

	res, err := srv.AssumeSessionPolicy(authContextForUser(t, 7), &pbauthv1.AssumeSessionPolicyRequest{
		AccessToken: accessToken,
		Statements: []*pbauthv1.PolicyStatement{{
			Effect:          "allow",
			ActionPattern:   "order:read",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-scoped", res.JwtToken)
}

func TestRefreshToken_OK(t *testing.T) {
	srv, authUC, _, _ := newAuthServer(t)
	authUC.EXPECT().RefreshAccessToken(mock.Anything, "refresh-token").Return(&inputport.AuthResult{
		JwtToken:     "jwt-refreshed",
		RefreshToken: "refresh-next",
		UserInfo:     entity.User{Id: 7, Username: "neo", Email: "neo@mx.io"},
	}, nil)

	res, err := srv.RefreshToken(context.Background(), &pbauthv1.RefreshTokenRequest{RefreshToken: "refresh-token"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-refreshed", res.JwtToken)
	assert.Equal(t, "refresh-next", res.RefreshToken)
}

func TestGetSession_OK(t *testing.T) {
	srv, _, sessionRepo, _ := newAuthServer(t)
	session := sessionWithID("session-1", 7)
	sessionRepo.EXPECT().GetByID(mock.Anything, "session-1").Return(session, nil)

	res, err := srv.GetSession(context.Background(), &pbauthv1.GetSessionRequest{SessionId: "session-1"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "session-1", res.Session.Id)
	assert.Equal(t, uint64(7), res.Session.UserId)
}
