package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	// mocks
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"

	// domain
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/dto"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
)

//
// ---------------- Helpers ----------------
//

func initMock() (
	*inputmocks.MockUserUsecase,
	*inputmocks.MockTokenUsecase,
	*outputmocks.MockGoogleOauthExternal,
	*outputmocks.MockUserRepository,
	*outputmocks.MockOauthStateRepository,
) {
	return &inputmocks.MockUserUsecase{},
		&inputmocks.MockTokenUsecase{},
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockUserRepository{},
		&outputmocks.MockOauthStateRepository{}
}

func newUC(
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

func makeTokenServerOK(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// minimal OAuth token JSON
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "tok123",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
}

func makeTokenServerErr(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadRequest)
	}))
}

func makeOAuthCfg(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "cid",
		ClientSecret: "sec",
		Endpoint: oauth2.Endpoint{
			AuthURL:   tokenURL + "/auth",
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams, // creds in params (safe for our fake)
		},
		RedirectURL: "https://app.example.com/callback",
		Scopes:      []string{"openid", "email", "profile"},
	}
}

//
// ---------------- Tests ----------------
//

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	hashed, err := entity.GeneratePasswordHash("pass123")
	require.NoError(t, err)

	user := &entity.User{
		Id:       1,
		Username: "jdoe",
		Email:    "jdoe@example.com",
		Password: hashed,
	}

	ur.On("GetByUsernameOrEmail", "jdoe").Return(user, nil)
	tuc.On("CreateJwtToken", *user).Return("jwt-token", nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)

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

	created := &entity.User{Id: 10, Username: "neo", Email: "neo@mx.io"}

	ur.
		On("Create", entity.User{Username: "neo", Password: "TheOne!", Email: "neo@mx.io"}).
		Return(created, nil)

	ur.
		On("UpdateById", uint(10), entity.User{InitialFrom: "podzone"}).
		Return(nil)

	tuc.
		On("CreateJwtToken", *created).
		Return("jwt-register", nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)

	resp, err := uc.Register(ctx, dto.RegisterReq{
		Username: "neo", Password: "TheOne!", Email: "neo@mx.io",
	})
	require.NoError(t, err)
	assert.Equal(t, "jwt-register", resp.JwtToken)
	assert.Equal(t, "neo@mx.io", resp.UserInfo.Email)

	ur.AssertExpectations(t)
	tuc.AssertExpectations(t)
}

func TestGenerateOAuthURL_SetsStateAndReturnsURL(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

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

	var capturedKey string
	sr.
		On("Set",
			mock.MatchedBy(func(key string) bool {
				ok := strings.HasPrefix(key, "oauth:google:")
				if ok {
					capturedKey = key
				}
				return ok
			}),
			mock.AnythingOfType("time.Duration"),
		).
		Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)

	urlStr, err := uc.GenerateOAuthURL(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, urlStr)

	parsed, err := url.Parse(urlStr)
	require.NoError(t, err)
	state := parsed.Query().Get("state")
	require.NotEmpty(t, state)
	assert.Equal(t, "oauth:google:"+state, capturedKey)

	ext.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestHandleOAuthCallback_HappyPath(t *testing.T) {
	ts := makeTokenServerOK(t)
	defer ts.Close()

	uuc, tuc, _, ur, sr := initMock()
	ext := &outputmocks.MockGoogleOauthExternal{}
	ext.On("GetConfig").Return(makeOAuthCfg(ts.URL))

	// Google profile
	ginfo := &outputport.GoogleUserInfo{
		Email: "jdoe@example.com",
		Name:  "John Doe",
	}
	ext.On("FetchUserInfo", "tok123").Return(ginfo, nil)

	// After mapping, match by email only (MapStruct details are opaque here)
	uuc.
		On("CreateNewAfterAuthCallback",
			mock.MatchedBy(func(e entity.User) bool { return e.Email == "jdoe@example.com" }),
		).
		Return(&entity.User{Id: 7, Email: "jdoe@example.com"}, nil)

	tuc.
		On("CreateJwtToken",
			mock.MatchedBy(func(e entity.User) bool { return e.Email == "jdoe@example.com" }),
		).
		Return("jwt-ok", nil)

	state := "ST"
	key := "oauth:google:" + state
	sr.On("Get", key).Return("ok", nil)
	sr.On("Del", key).Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)

	resp, err := uc.HandleOAuthCallback(context.Background(), "CODE", state)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-ok", resp.JwtToken)
	assert.Equal(t, "jdoe@example.com", resp.UserInfo.Email)
	assert.Contains(t, resp.RedirectUrl, "token=jwt-ok")

	ext.AssertExpectations(t)
	uuc.AssertExpectations(t)
	tuc.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestHandleOAuthCallback_InvalidState(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	sr.On("Get", "oauth:google:BAD_STATE").Return("", fmt.Errorf("not found"))

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.HandleOAuthCallback(ctx, "CODE", "BAD_STATE")
	require.Error(t, err)
	assert.Nil(t, resp)

	sr.AssertExpectations(t)
}

func TestHandleOAuthCallback_ExchangeError(t *testing.T) {
	ts := makeTokenServerErr(t)
	defer ts.Close()

	uuc, tuc, _, ur, sr := initMock()
	ext := &outputmocks.MockGoogleOauthExternal{}
	ext.On("GetConfig").Return(makeOAuthCfg(ts.URL))

	state := "ST"
	key := "oauth:google:" + state
	sr.On("Get", key).Return("ok", nil)
	sr.On("Del", key).Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.HandleOAuthCallback(context.Background(), "CODE", state)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to exchange code")

	ext.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestHandleOAuthCallback_FetchUserInfoError(t *testing.T) {
	ts := makeTokenServerOK(t)
	defer ts.Close()

	uuc, tuc, _, ur, sr := initMock()
	ext := &outputmocks.MockGoogleOauthExternal{}
	ext.On("GetConfig").Return(makeOAuthCfg(ts.URL))

	ext.On("FetchUserInfo", "tok123").Return(nil, fmt.Errorf("fetch err"))

	state := "ST"
	key := "oauth:google:" + state
	sr.On("Get", key).Return("ok", nil)
	sr.On("Del", key).Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.HandleOAuthCallback(context.Background(), "CODE", state)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to fetch user info")

	ext.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestHandleOAuthCallback_CreateUserError(t *testing.T) {
	ts := makeTokenServerOK(t)
	defer ts.Close()

	uuc, tuc, _, ur, sr := initMock()
	ext := &outputmocks.MockGoogleOauthExternal{}
	ext.On("GetConfig").Return(makeOAuthCfg(ts.URL))

	ext.On("FetchUserInfo", "tok123").Return(&outputport.GoogleUserInfo{Email: "e@x.io"}, nil)

	uuc.
		On("CreateNewAfterAuthCallback",
			mock.MatchedBy(func(e entity.User) bool { return e.Email == "e@x.io" }),
		).
		Return((*entity.User)(nil), fmt.Errorf("persist fail"))

	state := "ST"
	key := "oauth:google:" + state
	sr.On("Get", key).Return("ok", nil)
	sr.On("Del", key).Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.HandleOAuthCallback(context.Background(), "CODE", state)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create new user")

	ext.AssertExpectations(t)
	uuc.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestHandleOAuthCallback_CreateJWTError(t *testing.T) {
	ts := makeTokenServerOK(t)
	defer ts.Close()

	uuc, tuc, _, ur, sr := initMock()
	ext := &outputmocks.MockGoogleOauthExternal{}
	ext.On("GetConfig").Return(makeOAuthCfg(ts.URL))

	ext.On("FetchUserInfo", "tok123").Return(&outputport.GoogleUserInfo{Email: "a@b.com"}, nil)

	user := entity.User{Id: 1, Email: "a@b.com"}
	uuc.
		On("CreateNewAfterAuthCallback",
			mock.MatchedBy(func(e entity.User) bool { return e.Email == "a@b.com" }),
		).
		Return(&user, nil)

	tuc.
		On("CreateJwtToken", user).
		Return("", fmt.Errorf("jwt fail"))

	state := "ST"
	key := "oauth:google:" + state
	sr.On("Get", key).Return("ok", nil)
	sr.On("Del", key).Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.HandleOAuthCallback(context.Background(), "CODE", state)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create JWT")

	ext.AssertExpectations(t)
	tuc.AssertExpectations(t)
	uuc.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestLogout(t *testing.T) {
	uc := newUC(
		t,
		&inputmocks.MockUserUsecase{},
		&inputmocks.MockTokenUsecase{},
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockUserRepository{},
		&outputmocks.MockOauthStateRepository{},
	)
	loc, err := uc.Logout(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "/", loc)
}
