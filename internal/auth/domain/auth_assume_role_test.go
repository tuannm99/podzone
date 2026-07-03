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
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
)

func TestAssumeRole_Success(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	user := &entity.User{Id: 9, Email: "neo@mx.io", Username: "neo"}
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByID", "9").Return(user, nil)
	tokenUC := &inputmocks.MockTokenUsecase{}
	tokenUC.On("CreateJwtTokenForSessionState", *user, mock.MatchedBy(func(session entity.Session) bool {
		return session.AssumedRoleName == iamdomain.RoleTenantAdmin &&
			session.AssumedRoleScope == iamdomain.PolicyScopeTenant &&
			session.AssumedRoleTenantID == "tenant-1" &&
			len(session.SessionPolicy) == 1 &&
			session.SessionPolicy[0].ActionPattern == "order:read"
	})).Return("jwt-assumed", nil)
	sessionRepo := outputmocks.NewMockSessionRepository(t)
	refreshRepo := outputmocks.NewMockRefreshTokenRepository(t)
	tenantAccessChecker := outputmocks.NewMockTenantAccessChecker(t)
	roleAssumer := outputmocks.NewMockRoleAssumer(t)
	state := &authRepoState{sessions: map[string]entity.Session{}, refreshTokens: map[string]entity.RefreshToken{}}
	sessionRepo.EXPECT().
		GetByID(mock.Anything, "session-1").
		RunAndReturn(func(ctx context.Context, id string) (*entity.Session, error) {
			item := state.sessions[id]
			copyItem := item
			return &copyItem, nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		UpdateSessionPolicy(mock.Anything, "session-1", mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, id string, statements []entity.SessionPolicyStatement, updatedAt time.Time) error {
			item := state.sessions[id]
			item.SessionPolicy = append([]entity.SessionPolicyStatement(nil), statements...)
			item.UpdatedAt = updatedAt
			state.sessions[id] = item
			return nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		UpdateAssumedRole(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, session entity.Session, updatedAt time.Time) error {
			item := state.sessions[session.ID]
			item.AssumedRoleID = session.AssumedRoleID
			item.AssumedRoleScope = session.AssumedRoleScope
			item.AssumedRoleName = session.AssumedRoleName
			item.AssumedRoleTenantID = session.AssumedRoleTenantID
			item.AssumedRoleServicePrincipal = session.AssumedRoleServicePrincipal
			item.AssumedRoleSessionName = session.AssumedRoleSessionName
			item.AssumedRoleSourceIdentity = session.AssumedRoleSourceIdentity
			item.AssumedRoleExpiresAt = session.AssumedRoleExpiresAt
			item.SessionTags = cloneStringMap(session.SessionTags)
			item.ActiveTenantID = session.ActiveTenantID
			item.UpdatedAt = updatedAt
			state.sessions[session.ID] = item
			return nil
		}).
		Maybe()
	tenantAccessChecker.EXPECT().EnsureActiveMembership(mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
	roleAssumer.EXPECT().AssumeRole(mock.Anything, mock.MatchedBy(func(input outputport.AssumeRoleInput) bool {
		return input.AccessToken != "" &&
			input.UserID == 9 &&
			input.RoleName == iamdomain.RoleTenantAdmin &&
			input.TenantID == "tenant-1"
	})).Return(&outputport.AssumedRole{
		RoleID:    2,
		RoleScope: iamdomain.PolicyScopeTenant,
		RoleName:  iamdomain.RoleTenantAdmin,
		TenantID:  "tenant-1",
	}, nil)

	uc := NewAuthUsecase(
		&inputmocks.MockUserUsecase{},
		tokenUC,
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		userRepo,
		sessionRepo,
		refreshRepo,
		tenantAccessChecker,
		roleAssumer,
		outputmocks.NewMockAccountBootstrapper(t),
		cfg,
	)
	state.sessions["session-1"] = entity.Session{
		ID:             "session-1",
		UserID:         9,
		ActiveTenantID: "",
		Status:         entity.SessionStatusActive,
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	accessToken, err := NewTokenUsecase(
		cfg,
	).CreateJwtTokenForSession(entity.User{Id: 9, Email: "neo@mx.io", Username: "neo"}, "", "session-1")
	require.NoError(t, err)

	resp, err := uc.AssumeRole(
		context.Background(),
		9,
		accessToken,
		iamdomain.RoleTenantAdmin,
		"tenant-1",
		[]entity.SessionPolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:read",
			ResourcePattern: "*",
		}},
		"",
		"",
		"",
		0,
		"",
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-assumed", resp.JwtToken)
	assert.Equal(t, iamdomain.RoleTenantAdmin, state.sessions["session-1"].AssumedRoleName)
	assert.Equal(t, "tenant-1", state.sessions["session-1"].ActiveTenantID)
}

func TestAssumeRole_IAMDeny(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "secret",
		JWTKey:         "app-key",
		AppRedirectURL: "https://app.example.com/after-auth",
	}
	user := &entity.User{Id: 9, Email: "neo@mx.io", Username: "neo"}
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByID", "9").Return(user, nil)
	tokenUC := &inputmocks.MockTokenUsecase{}
	sessionRepo := outputmocks.NewMockSessionRepository(t)
	refreshRepo := outputmocks.NewMockRefreshTokenRepository(t)
	tenantAccessChecker := outputmocks.NewMockTenantAccessChecker(t)
	roleAssumer := outputmocks.NewMockRoleAssumer(t)
	state := &authRepoState{sessions: map[string]entity.Session{}, refreshTokens: map[string]entity.RefreshToken{}}
	sessionRepo.EXPECT().
		GetByID(mock.Anything, "session-1").
		RunAndReturn(func(ctx context.Context, id string) (*entity.Session, error) {
			item := state.sessions[id]
			copyItem := item
			return &copyItem, nil
		}).
		Maybe()
	roleAssumer.EXPECT().AssumeRole(mock.Anything, mock.MatchedBy(func(input outputport.AssumeRoleInput) bool {
		return input.AccessToken != "" &&
			input.UserID == 9 &&
			input.RoleName == iamdomain.RoleTenantAdmin &&
			input.TenantID == "tenant-2"
	})).Return(nil, iamdomain.ErrAssumeRoleDenied)

	uc := NewAuthUsecase(
		&inputmocks.MockUserUsecase{},
		tokenUC,
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		userRepo,
		sessionRepo,
		refreshRepo,
		tenantAccessChecker,
		roleAssumer,
		outputmocks.NewMockAccountBootstrapper(t),
		cfg,
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

	resp, err := uc.AssumeRole(
		context.Background(),
		9,
		accessToken,
		iamdomain.RoleTenantAdmin,
		"tenant-2",
		nil,
		"",
		"",
		"",
		0,
		"",
		nil,
	)
	require.ErrorIs(t, err, iamdomain.ErrAssumeRoleDenied)
	require.Nil(t, resp)
	assert.Empty(t, state.sessions["session-1"].AssumedRoleName)
}

func TestClearAssumedRole_PreservesSessionPolicy(t *testing.T) {
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
	}}).Return("jwt-cleared-assumed", nil)
	sessionRepo := outputmocks.NewMockSessionRepository(t)
	refreshRepo := outputmocks.NewMockRefreshTokenRepository(t)
	tenantAccessChecker := outputmocks.NewMockTenantAccessChecker(t)
	roleAssumer := outputmocks.NewMockRoleAssumer(t)
	state := &authRepoState{sessions: map[string]entity.Session{}, refreshTokens: map[string]entity.RefreshToken{}}
	sessionRepo.EXPECT().
		GetByID(mock.Anything, "session-1").
		RunAndReturn(func(ctx context.Context, id string) (*entity.Session, error) {
			item := state.sessions[id]
			copyItem := item
			return &copyItem, nil
		}).
		Maybe()
	sessionRepo.EXPECT().
		UpdateAssumedRole(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, session entity.Session, updatedAt time.Time) error {
			item := state.sessions[session.ID]
			item.AssumedRoleID = session.AssumedRoleID
			item.AssumedRoleScope = session.AssumedRoleScope
			item.AssumedRoleName = session.AssumedRoleName
			item.AssumedRoleTenantID = session.AssumedRoleTenantID
			item.UpdatedAt = updatedAt
			state.sessions[session.ID] = item
			return nil
		}).
		Maybe()

	uc := NewAuthUsecase(
		&inputmocks.MockUserUsecase{},
		tokenUC,
		&outputmocks.MockGoogleOauthExternal{},
		&outputmocks.MockOauthStateRepository{},
		userRepo,
		sessionRepo,
		refreshRepo,
		tenantAccessChecker,
		roleAssumer,
		outputmocks.NewMockAccountBootstrapper(t),
		cfg,
	)
	state.sessions["session-1"] = entity.Session{
		ID:                  "session-1",
		UserID:              9,
		ActiveTenantID:      "tenant-1",
		AssumedRoleID:       2,
		AssumedRoleScope:    iamdomain.PolicyScopeTenant,
		AssumedRoleName:     iamdomain.RoleTenantAdmin,
		AssumedRoleTenantID: "tenant-1",
		SessionPolicy: []entity.SessionPolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:read",
			ResourcePattern: "*",
		}},
		Status:    entity.SessionStatusActive,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	accessToken, err := NewTokenUsecase(
		cfg,
	).CreateJwtTokenForSessionState(entity.User{Id: 9, Email: "neo@mx.io", Username: "neo"}, state.sessions["session-1"])
	require.NoError(t, err)

	resp, err := uc.ClearAssumedRole(context.Background(), 9, accessToken)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "jwt-cleared-assumed", resp.JwtToken)
	assert.Zero(t, state.sessions["session-1"].AssumedRoleID)
	assert.Empty(t, state.sessions["session-1"].AssumedRoleName)
	require.Len(t, state.sessions["session-1"].SessionPolicy, 1)
	assert.Equal(t, "order:read", state.sessions["session-1"].SessionPolicy[0].ActionPattern)
}
