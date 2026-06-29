package domain

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/collection"
)

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
	uc, _, _, _ := newStatefulAuthUC(
		t,
		cfg,
		uuc,
		tuc,
		ext,
		sr,
		ur,
		func(ctx context.Context, tenantID string, userID uint) error {
			return entity.ErrMembershipNotFound
		},
	)
	return uc
}

type authRepoState struct {
	sessions      map[string]entity.Session
	refreshTokens map[string]entity.RefreshToken
}

func newStatefulAuthUC(
	t *testing.T,
	cfg config.AuthConfig,
	uuc *inputmocks.MockUserUsecase,
	tuc inputport.TokenUsecase,
	ext *outputmocks.MockGoogleOauthExternal,
	sr outputport.OauthStateRepository,
	ur *outputmocks.MockUserRepository,
	tenantAccessFn func(ctx context.Context, tenantID string, userID uint) error,
) (*authInteractorImpl, *authRepoState, *outputmocks.MockSessionRepository, *outputmocks.MockRefreshTokenRepository) {
	t.Helper()

	state := &authRepoState{
		sessions:      map[string]entity.Session{},
		refreshTokens: map[string]entity.RefreshToken{},
	}
	sessionRepo := outputmocks.NewMockSessionRepository(t)
	refreshRepo := outputmocks.NewMockRefreshTokenRepository(t)
	tenantAccessChecker := outputmocks.NewMockTenantAccessChecker(t)
	roleAssumer := outputmocks.NewMockRoleAssumer(t)

	sessionRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, session entity.Session) error {
			state.sessions[session.ID] = session
			return nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, id string) (*entity.Session, error) {
			item, ok := state.sessions[id]
			if !ok {
				return nil, entity.ErrSessionNotFound
			}
			copyItem := item
			return &copyItem, nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		ListByUser(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			userID uint,
			query collection.Query,
		) (collection.Page[entity.Session], error) {
			out := make([]entity.Session, 0)
			for _, item := range state.sessions {
				if item.UserID == userID {
					out = append(out, item)
				}
			}
			return collection.NewPage(out, int64(len(out)), query), nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		UpdateActiveTenant(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, id, tenantID string, updatedAt time.Time) error {
			item, ok := state.sessions[id]
			if !ok {
				return entity.ErrSessionNotFound
			}
			item.ActiveTenantID = tenantID
			item.UpdatedAt = updatedAt
			state.sessions[id] = item
			return nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		UpdateSessionPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			id string,
			statements []entity.SessionPolicyStatement,
			updatedAt time.Time,
		) error {
			item, ok := state.sessions[id]
			if !ok {
				return entity.ErrSessionNotFound
			}
			item.SessionPolicy = append([]entity.SessionPolicyStatement(nil), statements...)
			item.UpdatedAt = updatedAt
			state.sessions[id] = item
			return nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		Revoke(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, id string, revokedAt time.Time) error {
			item, ok := state.sessions[id]
			if !ok {
				return entity.ErrSessionNotFound
			}
			item.Status = entity.SessionStatusRevoked
			item.RevokedAt = &revokedAt
			state.sessions[id] = item
			return nil
		}).
		Maybe()

	refreshRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, token entity.RefreshToken) error {
			state.refreshTokens[token.TokenHash] = token
			return nil
		}).
		Maybe()
	refreshRepo.EXPECT().
		GetByTokenHash(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
			item, ok := state.refreshTokens[tokenHash]
			if !ok {
				return nil, entity.ErrRefreshTokenInvalid
			}
			copyItem := item
			return &copyItem, nil
		}).
		Maybe()
	refreshRepo.EXPECT().
		Revoke(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, id string, revokedAt time.Time, replacedByTokenID *string) error {
			for key, item := range state.refreshTokens {
				if item.ID == id {
					item.RevokedAt = &revokedAt
					item.ReplacedByTokenID = replacedByTokenID
					state.refreshTokens[key] = item
					return nil
				}
			}
			return entity.ErrRefreshTokenInvalid
		}).
		Maybe()
	refreshRepo.EXPECT().
		RevokeBySession(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, sessionID string, revokedAt time.Time) error {
			for key, item := range state.refreshTokens {
				if item.SessionID == sessionID {
					item.RevokedAt = &revokedAt
					state.refreshTokens[key] = item
				}
			}
			return nil
		}).
		Maybe()

	tenantAccessChecker.EXPECT().
		EnsureActiveMembership(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(tenantAccessFn).
		Maybe()

	return NewAuthUsecase(
		uuc,
		tuc,
		ext,
		sr,
		ur,
		sessionRepo,
		refreshRepo,
		tenantAccessChecker,
		roleAssumer,
		cfg,
	), state, sessionRepo, refreshRepo
}

func makeTokenServerOK(t *testing.T) *httptest.Server {
	t.Helper()
	return newOptionalHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	return newOptionalHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadRequest)
	}))
}

func newOptionalHTTPServer(t *testing.T, handler http.Handler) (server *httptest.Server) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Skipf("skipping local httptest server: %v", recovered)
		}
	}()
	return httptest.NewServer(handler)
}

func makeOAuthCfg(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "cid",
		ClientSecret: "sec",
		Endpoint: oauth2.Endpoint{
			AuthURL:   tokenURL + "/auth",
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: "https://app.example.com/callback",
		Scopes:      []string{"openid", "email", "profile"},
	}
}
