package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
)

func TestAssumeSessionPolicy_Success(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	user := &entity.User{Id: 9, Email: "neo@mx.io", Username: "neo"}
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByID", "9").Return(user, nil)
	tokenUC := &inputmocks.MockTokenUsecase{}
	tokenUC.On("CreateJwtTokenForScopedSession", *user, "tenant-1", "session-1", []entity.SessionPolicyStatement{{
		Effect:          iamdomain.PolicyEffectAllow,
		ActionPattern:   "order:read",
		ResourcePattern: "*",
	}}).Return("jwt-scoped", nil)
	uc, state, _, _ := newStatefulAuthUC(
		t,
		cfg,
		&inputmocks.MockUserUsecase{},
		tokenUC,
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
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	accessToken, err := NewTokenUsecase(
		cfg,
	).CreateJwtTokenForSession(entity.User{Id: 9, Email: "neo@mx.io", Username: "neo"}, "tenant-1", "session-1")
	require.NoError(t, err)

	resp, err := uc.AssumeSessionPolicy(context.Background(), 9, accessToken, []entity.SessionPolicyStatement{{
		Effect:          iamdomain.PolicyEffectAllow,
		ActionPattern:   "order:read",
		ResourcePattern: "*",
	}})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-scoped", resp.JwtToken)
	require.Len(t, state.sessions["session-1"].SessionPolicy, 1)
	assert.Equal(t, "order:read", state.sessions["session-1"].SessionPolicy[0].ActionPattern)
}

func TestClearSessionPolicy_Success(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	user := &entity.User{Id: 9, Email: "neo@mx.io", Username: "neo"}
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByID", "9").Return(user, nil)
	tokenUC := &inputmocks.MockTokenUsecase{}
	tokenUC.On("CreateJwtTokenForSession", *user, "tenant-1", "session-1").Return("jwt-cleared", nil)
	uc, state, _, _ := newStatefulAuthUC(
		t,
		cfg,
		&inputmocks.MockUserUsecase{},
		tokenUC,
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		userRepo,
		func(ctx context.Context, tenantID string, userID uint) error { return nil },
	)
	state.sessions["session-1"] = entity.Session{
		ID:             "session-1",
		UserID:         9,
		ActiveTenantID: "tenant-1",
		SessionPolicy: []entity.SessionPolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:read",
			ResourcePattern: "*",
		}},
		Status:    entity.SessionStatusActive,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	accessToken, err := NewTokenUsecase(cfg).CreateJwtTokenForScopedSession(
		entity.User{
			Id:       9,
			Email:    "neo@mx.io",
			Username: "neo",
		}, "tenant-1", "session-1", state.sessions["session-1"].SessionPolicy)
	require.NoError(t, err)

	resp, err := uc.ClearSessionPolicy(context.Background(), 9, accessToken)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-cleared", resp.JwtToken)
	assert.Empty(t, state.sessions["session-1"].SessionPolicy)
}
