package backoffice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authconfig "github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	authentity "github.com/tuannm99/podzone/internal/auth/domain/entity"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
)

type authorizerFake struct{}

func (authorizerFake) AuthorizeTenant(ctx context.Context, sessionID, userID, tenantID string) error {
	return nil
}

func (authorizerFake) RequirePermission(ctx context.Context, userID, tenantID, permission string) error {
	return nil
}

func TestIdentityFromAuthorization_ReadsUserAndActiveTenant(t *testing.T) {
	tokenUC := authdomain.NewTokenUsecase(authconfig.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	})
	token, err := tokenUC.CreateJwtTokenForSession(authentity.User{
		Id:       12,
		Email:    "owner@podzone.io",
		Username: "owner",
	}, "tenant-1", "session-1")
	require.NoError(t, err)

	m := NewTenantMiddleware(boconfig.Config{
		Auth: boconfig.AuthConfig{
			JWTSecret: "secret",
			JWTKey:    "app-key",
		},
	}, authorizerFake{})

	userID, tenantID, sessionID, err := m.identityFromAuthorization("Bearer " + token)
	require.NoError(t, err)
	assert.Equal(t, "12", userID)
	assert.Equal(t, "tenant-1", tenantID)
	assert.Equal(t, "session-1", sessionID)
}

func TestResolveTenantID(t *testing.T) {
	t.Run("require claim tenant", func(t *testing.T) {
		tenantID, err := resolveTenantID("tenant-1")
		require.NoError(t, err)
		assert.Equal(t, "tenant-1", tenantID)
	})

	t.Run("reject missing active tenant", func(t *testing.T) {
		tenantID, err := resolveTenantID("")
		require.Error(t, err)
		assert.Empty(t, tenantID)
	})
}

func TestPermissionForField(t *testing.T) {
	permission, ok := permissionForField("Mutation", "deactivateStore")
	require.True(t, ok)
	assert.Equal(t, "store:deactivate", permission)

	permission, ok = permissionForField("Query", "stores")
	require.True(t, ok)
	assert.Equal(t, "store:read", permission)

	permission, ok = permissionForField("Query", "unknown")
	assert.False(t, ok)
	assert.Empty(t, permission)
}
