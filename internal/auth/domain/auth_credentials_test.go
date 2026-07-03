package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
)

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

	ur.On("Create", entity.User{Username: "neo", Password: "TheOne!", Email: "neo@mx.io"}).Return(created, nil)
	ur.On("UpdateById", uint(10), entity.User{InitialFrom: "podzone"}).Return(nil)
	tuc.On(
		"CreateJwtTokenForSession",
		mock.MatchedBy(func(user entity.User) bool {
			return user.Id == created.Id && user.InitialFrom == "podzone"
		}),
		"",
		mock.AnythingOfType("string"),
	).Return("jwt-register", nil)

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

func TestSwitchActiveTenant_Success(t *testing.T) {
	ctx := context.Background()
	uuc, tuc, ext, ur, sr := initMock()

	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	uc, state, _, _ := newStatefulAuthUC(
		t,
		cfg,
		uuc,
		tuc,
		ext,
		sr,
		ur,
		func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
	)
	state.sessions["session-1"] = entity.Session{
		ID:        "session-1",
		UserID:    5,
		Status:    entity.SessionStatusActive,
		ExpiresAt: time.Now().Add(time.Hour),
	}

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
			return entity.ErrInactiveMembership
		},
	)

	resp, err := uc.SwitchActiveTenant(ctx, 5, "tenant-1", "access-token")
	require.ErrorIs(t, err, entity.ErrInactiveMembership)
	assert.Nil(t, resp)
}

func TestLogout(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	uc, state, _, _ := newStatefulAuthUC(
		t,
		cfg,
		&inputmocks.MockUserUsecase{},
		&inputmocks.MockTokenUsecase{},
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		&outputmocks.MockUserRepository{},
		func(ctx context.Context, tenantID string, userID uint) error { return nil },
	)
	state.sessions["session-1"] = entity.Session{
		ID:        "session-1",
		UserID:    9,
		Status:    entity.SessionStatusActive,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	state.refreshTokens["h1"] = entity.RefreshToken{ID: "refresh-1", SessionID: "session-1"}
	tokenUC := NewTokenUsecase(cfg)
	accessToken, err := tokenUC.CreateJwtTokenForSession(entity.User{Id: 9}, "", "session-1")
	require.NoError(t, err)

	loc, err := uc.Logout(context.Background(), accessToken)
	require.NoError(t, err)
	assert.Equal(t, "/", loc)
	assert.Equal(t, entity.SessionStatusRevoked, state.sessions["session-1"].Status)
}

func TestRefreshAccessToken_Success(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	now := time.Now().UTC()
	rawRefresh := "refresh-raw-token"
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByID", "9").Return(&entity.User{
		Id:       9,
		Email:    "neo@mx.io",
		Username: "neo",
	}, nil)
	uc, state, _, _ := newStatefulAuthUC(
		t,
		cfg,
		&inputmocks.MockUserUsecase{},
		NewTokenUsecase(cfg),
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		userRepo,
		func(ctx context.Context, tenantID string, userID uint) error { return nil },
	)
	state.sessions["session-1"] = entity.Session{
		ID:             "session-1",
		UserID:         9,
		ActiveTenantID: "tenant-1",
		Status:         entity.SessionStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(time.Hour),
	}
	state.refreshTokens[entity.HashToken(rawRefresh)] = entity.RefreshToken{
		ID:        "refresh-1",
		SessionID: "session-1",
		TokenHash: entity.HashToken(rawRefresh),
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp, err := uc.RefreshAccessToken(context.Background(), rawRefresh)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.JwtToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, rawRefresh, resp.RefreshToken)
	assert.Equal(t, uint(9), resp.UserInfo.Id)

	storedOld := state.refreshTokens[entity.HashToken(rawRefresh)]
	assert.NotNil(t, storedOld.RevokedAt)
	userRepo.AssertExpectations(t)
}
