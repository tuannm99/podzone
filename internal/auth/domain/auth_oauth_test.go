package domain

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
)

func TestGenerateOAuthURL_SetsStateAndReturnsURL(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	cfg := makeOAuthCfg("https://oauth2.example/token")
	ext.On("GetConfig").Return(cfg)

	var capturedKey string
	sr.On("Set", mock.MatchedBy(func(key string) bool {
		ok := strings.HasPrefix(key, "oauth:google:")
		if ok {
			capturedKey = key
		}
		return ok
	}), mock.AnythingOfType("time.Duration")).Return(nil)

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

	ginfo := &outputport.GoogleUserInfo{
		Email: "jdoe@example.com",
		Name:  "John Doe",
	}
	ext.On("FetchUserInfo", "tok123").Return(ginfo, nil)

	uuc.On("CreateNewAfterAuthCallback", mock.MatchedBy(func(e entity.User) bool { return e.Email == "jdoe@example.com" })).
		Return(&entity.User{Id: 7, Email: "jdoe@example.com"}, nil)

	tuc.On("CreateJwtTokenForSession", mock.MatchedBy(func(e entity.User) bool { return e.Email == "jdoe@example.com" }), "", mock.AnythingOfType("string")).
		Return("jwt-ok", nil)

	state := "ST"
	key := "oauth:google:" + state
	sr.On("Get", key).Return("ok", nil)
	sr.On("Del", key).Return(nil)
	sr.On("SetValue", mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "oauth:google:exchange:")
	}), mock.AnythingOfType("string"), 2*time.Minute).Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)

	resp, err := uc.HandleOAuthCallback(context.Background(), "CODE", state)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.ExchangeCode)
	assert.Equal(t, "jdoe@example.com", resp.UserInfo.Email)
	assert.Contains(t, resp.RedirectUrl, "exchange_code=")
}

func TestHandleOAuthCallback_InvalidState(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	sr.On("Get", "oauth:google:BAD_STATE").Return("", fmt.Errorf("not found"))

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.HandleOAuthCallback(ctx, "CODE", "BAD_STATE")
	require.Error(t, err)
	assert.Nil(t, resp)
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
}

func TestHandleOAuthCallback_CreateUserError(t *testing.T) {
	ts := makeTokenServerOK(t)
	defer ts.Close()

	uuc, tuc, _, ur, sr := initMock()
	ext := &outputmocks.MockGoogleOauthExternal{}
	ext.On("GetConfig").Return(makeOAuthCfg(ts.URL))
	ext.On("FetchUserInfo", "tok123").Return(&outputport.GoogleUserInfo{Email: "e@x.io"}, nil)

	uuc.On("CreateNewAfterAuthCallback", mock.MatchedBy(func(e entity.User) bool { return e.Email == "e@x.io" })).
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
}

func TestHandleOAuthCallback_PersistExchangeError(t *testing.T) {
	ts := makeTokenServerOK(t)
	defer ts.Close()

	uuc, tuc, _, ur, sr := initMock()
	ext := &outputmocks.MockGoogleOauthExternal{}
	ext.On("GetConfig").Return(makeOAuthCfg(ts.URL))
	ext.On("FetchUserInfo", "tok123").Return(&outputport.GoogleUserInfo{Email: "a@b.com"}, nil)

	user := entity.User{Id: 1, Email: "a@b.com"}
	uuc.On("CreateNewAfterAuthCallback", mock.MatchedBy(func(e entity.User) bool { return e.Email == "a@b.com" })).
		Return(&user, nil)
	tuc.On("CreateJwtTokenForSession", user, "", mock.AnythingOfType("string")).Return("jwt-ok", nil)

	state := "ST"
	key := "oauth:google:" + state
	sr.On("Get", key).Return("ok", nil)
	sr.On("Del", key).Return(nil)
	sr.On("SetValue", mock.AnythingOfType("string"), mock.AnythingOfType("string"), 2*time.Minute).
		Return(fmt.Errorf("redis fail"))

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.HandleOAuthCallback(context.Background(), "CODE", state)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to persist exchange code")
}

func TestExchangeOAuthLogin_Success(t *testing.T) {
	uuc, tuc, ext, ur, sr := initMock()
	payload := `{"jwt_token":"jwt-oauth","refresh_token":"refresh-oauth","user_info":{"id":7,"email":"neo@mx.io","username":"neo"}}`
	sr.On("Get", "oauth:google:exchange:code-1").Return(payload, nil)
	sr.On("Del", "oauth:google:exchange:code-1").Return(nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)
	resp, err := uc.ExchangeOAuthLogin(context.Background(), "code-1")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-oauth", resp.JwtToken)
	assert.Equal(t, "refresh-oauth", resp.RefreshToken)
	assert.Equal(t, uint(7), resp.UserInfo.Id)
}
