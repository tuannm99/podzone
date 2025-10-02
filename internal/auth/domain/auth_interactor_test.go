package domain

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	// "os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/dto"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
)

type memStateRepo struct {
	m map[string]string
}

func newMemStateRepo() *memStateRepo { return &memStateRepo{m: map[string]string{}} }

func (r *memStateRepo) Set(key string, ttl time.Duration) error {
	// TTL không dùng trong test; lưu 1 marker
	r.m[key] = "ok"
	return nil
}

func (r *memStateRepo) Get(key string) (string, error) {
	v, ok := r.m[key]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return v, nil
}

func (r *memStateRepo) Del(key string) error {
	delete(r.m, key)
	return nil
}

// Ensure it matches your interface
var _ outputport.OauthStateRepository = (*memStateRepo)(nil)

// --------- helpers ---------

func newInteractorForTests(
	t *testing.T,
	uuc *inputmocks.MockUserUsecase,
	tuc *inputmocks.MockTokenUsecase,
	ext *outputmocks.MockGoogleOauthExternal,
	ur *outputmocks.MockUserRepository,
	sr outputport.OauthStateRepository,
) *authInteractorImpl {
	t.Helper()
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	return NewAuthUsecase(uuc, tuc, ext, sr, ur, cfg)
}

func initMock() (
	*inputmocks.MockUserUsecase,
	*inputmocks.MockTokenUsecase,
	*outputmocks.MockGoogleOauthExternal,
	*outputmocks.MockUserRepository,
	*memStateRepo,
) {
	return &inputmocks.MockUserUsecase{},
		&inputmocks.MockTokenUsecase{},
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockUserRepository{},
		newMemStateRepo()
}

// --------- tests ---------

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	// hash password to pass CheckPassword
	hashed, err := entity.GeneratePasswordHash("pass123")
	require.NoError(t, err)

	// mocks
	uuc, tuc, ext, ur, sr := initMock()
	user := &entity.User{
		Id:       1,
		Username: "jdoe",
		Email:    "jdoe@example.com",
		Password: hashed,
	}

	ur.On("GetByUsernameOrEmail", "jdoe").Return(user, nil)
	tuc.On("CreateJwtToken", *user).Return("jwt-token", nil)

	uc := newInteractorForTests(t, uuc, tuc, ext, ur, sr)

	resp, err := uc.Login(ctx, "jdoe", "pass123")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-token", resp.JwtToken)
	assert.Equal(t, user.Email, resp.UserInfo.Email)

	ur.AssertExpectations(t)
	tuc.AssertExpectations(t)
}

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	created := &entity.User{
		Id:       10,
		Username: "neo",
		Email:    "neo@mx.io",
	}

	// Create returns user
	ur.On("Create", entity.User{
		Username: "neo",
		Password: "TheOne!",
		Email:    "neo@mx.io",
	}).Return(created, nil)

	// UpdateById called to set InitialFrom
	ur.On("UpdateById", uint(10), entity.User{InitialFrom: "podzone"}).Return(nil)

	// token
	tuc.On("CreateJwtToken", *created).Return("jwt-register", nil)

	uc := newInteractorForTests(t, uuc, tuc, ext, ur, sr)

	req := dto.RegisterReq{
		Username: "neo",
		Password: "TheOne!",
		Email:    "neo@mx.io",
	}
	resp, err := uc.Register(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "jwt-register", resp.JwtToken)
	assert.Equal(t, created.Email, resp.UserInfo.Email)

	ur.AssertExpectations(t)
	tuc.AssertExpectations(t)
}

func TestGenerateOAuthURL_SetsStateAndReturnsURL(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	// Google external mock: provide a Config with a dummy AuthURL
	cfg := &oauth2.Config{
		ClientID:     "cid",
		ClientSecret: "sec",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.example/token",
		},
		RedirectURL: "https://app.example.com/callback",
		Scopes:      []string{"openid", "email", "profile"},
	}
	ext.On("GetConfig").Return(cfg)

	uc := newInteractorForTests(t, uuc, tuc, ext, ur, sr)

	urlStr, err := uc.GenerateOAuthURL(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, urlStr)

	// parse state param and ensure it was stored
	parsed, err := url.Parse(urlStr)
	require.NoError(t, err)
	state := parsed.Query().Get("state")
	require.NotEmpty(t, state)

	// state should exist in mem repo
	_, getErr := sr.Get("oauth:google:" + state)
	require.NoError(t, getErr)

	ext.AssertExpectations(t)
}

func TestHandleOAuthCallback_HappyPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
        "access_token": "tok123",
        "token_type": "Bearer",
        "expires_in": 3600
    }`))
	}))
	defer ts.Close()

	ext := &outputmocks.MockGoogleOauthExternal{}
	cfg := &oauth2.Config{
		ClientID:     "cid",
		ClientSecret: "sec",
		Endpoint: oauth2.Endpoint{
			AuthURL:   ts.URL + "/auth",
			TokenURL:  ts.URL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: "https://app.example.com/callback",
	}
	ext.On("GetConfig").Return(cfg)
}

func TestHandleOAuthCallback_InvalidState(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	uc := newInteractorForTests(t, uuc, tuc, ext, ur, sr)

	resp, err := uc.HandleOAuthCallback(ctx, "CODE", "BAD_STATE")
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestLogout(t *testing.T) {
	uc := newInteractorForTests(
		t,
		&inputmocks.MockUserUsecase{},
		&inputmocks.MockTokenUsecase{},
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockUserRepository{},
		newMemStateRepo(),
	)
	loc, err := uc.Logout(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "/", loc)
}
