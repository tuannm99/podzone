package grpchandler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/dto"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

func newServerWithMock() (*AuthServer, *inputmocks.MockAuthUsecase) {
	uc := &inputmocks.MockAuthUsecase{}
	srv := NewAuthServer(uc)
	return srv, uc
}

func TestGoogleLogin_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("GenerateOAuthURL", mock.Anything).
		Return("https://accounts.google.com/auth?state=xyz", nil)

	res, err := srv.GoogleLogin(ctx, &pbauthv1.GoogleLoginRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "https://accounts.google.com/auth?state=xyz", res.RedirectUrl)

	uc.AssertExpectations(t)
}

func TestGoogleLogin_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("GenerateOAuthURL", mock.Anything).
		Return("", assert.AnError)

	res, err := srv.GoogleLogin(ctx, &pbauthv1.GoogleLoginRequest{})
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestGoogleCallback_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	cb := &dto.GoogleCallbackResp{
		JwtToken:    "jwt-abc",
		RedirectUrl: "https://app.example.com?token=jwt-abc",
		UserInfo: dto.UserInfoResp{
			Id:    "gid-1",
			Email: "neo@mx.io",
			Name:  "Neo",
		},
	}
	uc.On("HandleOAuthCallback", mock.Anything, "CODE", "STATE").
		Return(cb, nil)

	res, err := srv.GoogleCallback(ctx, &pbauthv1.GoogleCallbackRequest{
		Code:  "CODE",
		State: "STATE",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-abc", res.JwtToken)
	assert.Equal(t, "https://app.example.com?token=jwt-abc", res.RedirectUrl)
	assert.Equal(t, "neo@mx.io", res.UserInfo.Email)
	assert.Equal(t, "Neo", res.UserInfo.Name)

	uc.AssertExpectations(t)
}

func TestGoogleCallback_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("HandleOAuthCallback", mock.Anything, "BAD", "STATE").
		Return((*dto.GoogleCallbackResp)(nil), assert.AnError)

	res, err := srv.GoogleCallback(ctx, &pbauthv1.GoogleCallbackRequest{
		Code:  "BAD",
		State: "STATE",
	})
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestLogout_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("Logout", mock.Anything).
		Return("/", nil)

	res, err := srv.Logout(ctx, &pbauthv1.LogoutRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Success)
	assert.Equal(t, "/", res.RedirectUrl)

	uc.AssertExpectations(t)
}

func TestLogin_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	user := entity.User{
		Id:       1,
		Email:    "jdoe@example.com",
		Username: "jdoe",
	}
	uc.On("Login", mock.Anything, "jdoe", "pass").
		Return(&dto.LoginResp{
			JwtToken: "jwt-login",
			UserInfo: user,
		}, nil)

	res, err := srv.Login(ctx, &pbauthv1.LoginRequest{
		Username: "jdoe",
		Password: "pass",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-login", res.JwtToken)
	assert.Equal(t, "jdoe@example.com", res.UserInfo.Email)
	assert.Equal(t, "jdoe", res.UserInfo.Username)

	uc.AssertExpectations(t)
}

func TestLogin_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("Login", mock.Anything, "jdoe", "bad").
		Return((*dto.LoginResp)(nil), assert.AnError)

	res, err := srv.Login(ctx, &pbauthv1.LoginRequest{
		Username: "jdoe",
		Password: "bad",
	})
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestRegister_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	inReq := &pbauthv1.RegisterRequest{
		Username: "neo",
		Password: "TheOne!",
		Email:    "neo@mx.io",
	}
	out := &dto.RegisterResp{
		JwtToken: "jwt-reg",
		UserInfo: entity.User{
			Id:       9,
			Email:    "neo@mx.io",
			Username: "neo",
		},
	}

	// Dùng MatchedBy để không phụ thuộc chi tiết mapping (toolkit.MapStruct)
	uc.On("Register", mock.Anything, mock.MatchedBy(func(r dto.RegisterReq) bool {
		return r.Username == "neo" && r.Password == "TheOne!" && r.Email == "neo@mx.io"
	})).Return(out, nil)

	res, err := srv.Register(ctx, inReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-reg", res.JwtToken)
	assert.Equal(t, "neo@mx.io", res.UserInfo.Email)
	assert.Equal(t, "neo", res.UserInfo.Username)

	uc.AssertExpectations(t)
}

func TestRegister_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	inReq := &pbauthv1.RegisterRequest{
		Username: "neo",
		Password: "x",
		Email:    "neo@mx.io",
	}
	uc.On("Register", mock.Anything, mock.AnythingOfType("dto.RegisterReq")).
		Return((*dto.RegisterResp)(nil), assert.AnError)

	res, err := srv.Register(ctx, inReq)
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}
