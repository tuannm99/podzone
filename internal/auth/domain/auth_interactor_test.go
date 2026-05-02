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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	// mocks
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"

	// domain
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
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
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	tenantAccessChecker := &tenantAccessCheckerFake{
		ensureActiveMembershipFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return iamdomain.ErrMembershipNotFound
		},
	}
	return NewAuthUsecase(uuc, tuc, ext, sr, ur, &sessionRepoFake{}, &refreshTokenRepoFake{}, tenantAccessChecker, cfg)
}

type sessionRepoFake struct {
	items map[string]entity.Session
}

func (r *sessionRepoFake) Create(ctx context.Context, session entity.Session) error {
	if r.items == nil {
		r.items = map[string]entity.Session{}
	}
	r.items[session.ID] = session
	return nil
}

func (r *sessionRepoFake) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, entity.ErrSessionNotFound
	}
	copyItem := item
	return &copyItem, nil
}

func (r *sessionRepoFake) ListByUser(ctx context.Context, userID uint) ([]entity.Session, error) {
	out := make([]entity.Session, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *sessionRepoFake) UpdateActiveTenant(ctx context.Context, id, tenantID string, updatedAt time.Time) error {
	item, ok := r.items[id]
	if !ok {
		return entity.ErrSessionNotFound
	}
	item.ActiveTenantID = tenantID
	item.UpdatedAt = updatedAt
	r.items[id] = item
	return nil
}

func (r *sessionRepoFake) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	item, ok := r.items[id]
	if !ok {
		return entity.ErrSessionNotFound
	}
	item.Status = entity.SessionStatusRevoked
	item.RevokedAt = &revokedAt
	r.items[id] = item
	return nil
}

type refreshTokenRepoFake struct {
	items map[string]entity.RefreshToken
}

func (r *refreshTokenRepoFake) Create(ctx context.Context, token entity.RefreshToken) error {
	if r.items == nil {
		r.items = map[string]entity.RefreshToken{}
	}
	r.items[token.TokenHash] = token
	return nil
}

func (r *refreshTokenRepoFake) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	item, ok := r.items[tokenHash]
	if !ok {
		return nil, entity.ErrRefreshTokenInvalid
	}
	copyItem := item
	return &copyItem, nil
}

func (r *refreshTokenRepoFake) Revoke(
	ctx context.Context,
	id string,
	revokedAt time.Time,
	replacedByTokenID *string,
) error {
	for key, item := range r.items {
		if item.ID == id {
			item.RevokedAt = &revokedAt
			item.ReplacedByTokenID = replacedByTokenID
			r.items[key] = item
			return nil
		}
	}
	return entity.ErrRefreshTokenInvalid
}

func (r *refreshTokenRepoFake) RevokeBySession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	for key, item := range r.items {
		if item.SessionID == sessionID {
			item.RevokedAt = &revokedAt
			r.items[key] = item
		}
	}
	return nil
}

type tenantAccessCheckerFake struct {
	ensureActiveMembershipFunc func(ctx context.Context, tenantID string, userID uint) error
}

func (f *tenantAccessCheckerFake) EnsureActiveMembership(ctx context.Context, tenantID string, userID uint) error {
	return f.ensureActiveMembershipFunc(ctx, tenantID, userID)
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
	tuc.On("CreateJwtTokenForSession", *user, "", mock.AnythingOfType("string")).Return("jwt-token", nil)

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
		On("CreateJwtTokenForSession", *created, "", mock.AnythingOfType("string")).
		Return("jwt-register", nil)

	uc := newUC(t, uuc, tuc, ext, ur, sr)

	resp, err := uc.Register(ctx, inputport.RegisterCmd{
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
		On("CreateJwtTokenForSession",
			mock.MatchedBy(func(e entity.User) bool { return e.Email == "jdoe@example.com" }),
			"",
			mock.AnythingOfType("string"),
		).
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

func TestSwitchActiveTenant_Success(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()
	tenantAccessChecker := &tenantAccessCheckerFake{
		ensureActiveMembershipFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
	}

	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	sessionRepo := &sessionRepoFake{
		items: map[string]entity.Session{
			"session-1": {
				ID:        "session-1",
				UserID:    5,
				Status:    entity.SessionStatusActive,
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
	}
	refreshRepo := &refreshTokenRepoFake{}
	uc := NewAuthUsecase(uuc, tuc, ext, sr, ur, sessionRepo, refreshRepo, tenantAccessChecker, cfg)

	user := &entity.User{Id: 5, Username: "neo", Email: "neo@mx.io"}
	ur.On("GetByID", "5").Return(user, nil)
	tuc.On("CreateJwtTokenForSession", *user, "tenant-1", "session-1").Return("jwt-tenant-1", nil)

	tokenUC := NewTokenUsecase(cfg)
	accessToken, err := tokenUC.CreateJwtTokenForSession(
		entity.User{Id: 5, Email: "neo@mx.io", Username: "neo"},
		"",
		"session-1",
	)
	require.NoError(t, err)

	resp, err := uc.SwitchActiveTenant(ctx, 5, "tenant-1", accessToken)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-tenant-1", resp.JwtToken)
	assert.Equal(t, "neo@mx.io", resp.UserInfo.Email)
}

func TestSwitchActiveTenant_InactiveMembership(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()
	tenantAccessChecker := &tenantAccessCheckerFake{
		ensureActiveMembershipFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return iamdomain.ErrInactiveMembership
		},
	}

	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	uc := NewAuthUsecase(uuc, tuc, ext, sr, ur, &sessionRepoFake{}, &refreshTokenRepoFake{}, tenantAccessChecker, cfg)

	resp, err := uc.SwitchActiveTenant(ctx, 5, "tenant-1", "access-token")
	require.ErrorIs(t, err, iamdomain.ErrInactiveMembership)
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

func TestHandleOAuthCallback_PersistExchangeError(t *testing.T) {
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
		On("CreateJwtTokenForSession", user, "", mock.AnythingOfType("string")).
		Return("jwt-ok", nil)

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

	ext.AssertExpectations(t)
	tuc.AssertExpectations(t)
	uuc.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestLogout(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	sessionRepo := &sessionRepoFake{
		items: map[string]entity.Session{
			"session-1": {
				ID:        "session-1",
				UserID:    9,
				Status:    entity.SessionStatusActive,
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
	}
	refreshRepo := &refreshTokenRepoFake{
		items: map[string]entity.RefreshToken{
			"h1": {ID: "refresh-1", SessionID: "session-1"},
		},
	}
	uc := NewAuthUsecase(
		&inputmocks.MockUserUsecase{},
		&inputmocks.MockTokenUsecase{},
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		&outputmocks.MockUserRepository{},
		sessionRepo,
		refreshRepo,
		&tenantAccessCheckerFake{
			ensureActiveMembershipFunc: func(ctx context.Context, tenantID string, userID uint) error {
				return nil
			},
		},
		cfg,
	)
	tokenUC := NewTokenUsecase(cfg)
	accessToken, err := tokenUC.CreateJwtTokenForSession(entity.User{Id: 9}, "", "session-1")
	require.NoError(t, err)

	loc, err := uc.Logout(context.Background(), accessToken)
	require.NoError(t, err)
	assert.Equal(t, "/", loc)
	assert.Equal(t, entity.SessionStatusRevoked, sessionRepo.items["session-1"].Status)
}

func TestRefreshAccessToken_Success(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	now := time.Now().UTC()
	sessionRepo := &sessionRepoFake{
		items: map[string]entity.Session{
			"session-1": {
				ID:             "session-1",
				UserID:         9,
				ActiveTenantID: "tenant-1",
				Status:         entity.SessionStatusActive,
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiresAt:      now.Add(time.Hour),
			},
		},
	}
	rawRefresh := "refresh-raw-token"
	refreshRepo := &refreshTokenRepoFake{
		items: map[string]entity.RefreshToken{
			entity.HashToken(rawRefresh): {
				ID:        "refresh-1",
				SessionID: "session-1",
				TokenHash: entity.HashToken(rawRefresh),
				ExpiresAt: now.Add(time.Hour),
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByID", "9").Return(&entity.User{
		Id:       9,
		Email:    "neo@mx.io",
		Username: "neo",
	}, nil)
	uc := NewAuthUsecase(
		&inputmocks.MockUserUsecase{},
		NewTokenUsecase(cfg),
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		userRepo,
		sessionRepo,
		refreshRepo,
		&tenantAccessCheckerFake{
			ensureActiveMembershipFunc: func(ctx context.Context, tenantID string, userID uint) error {
				return nil
			},
		},
		cfg,
	)

	resp, err := uc.RefreshAccessToken(context.Background(), rawRefresh)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.JwtToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, rawRefresh, resp.RefreshToken)
	assert.Equal(t, uint(9), resp.UserInfo.Id)

	storedOld, err := refreshRepo.GetByTokenHash(context.Background(), entity.HashToken(rawRefresh))
	require.NoError(t, err)
	assert.NotNil(t, storedOld.RevokedAt)
	userRepo.AssertExpectations(t)
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
