package backoffice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authconfig "github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	authentity "github.com/tuannm99/podzone/internal/auth/domain/entity"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	backofficemocks "github.com/tuannm99/podzone/internal/backoffice/mocks"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/tenancy"
)

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

	authz := backofficemocks.NewMockTenantAuthorizer(t)
	bootstrapper := backofficemocks.NewMockTenantBootstrapper(t)
	m := NewTenantMiddleware(boconfig.Config{
		Auth: boconfig.RPCConfig{
			JWTSecret: "secret",
			JWTKey:    "app-key",
		},
	}, authz, tenancy.New(bootstrapper, nil, nil))

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

	permission, ok = permissionForField("Query", "routedOrderActivities")
	require.True(t, ok)
	assert.Equal(t, "store:read", permission)

	permission, ok = permissionForField("Query", "unknown")
	assert.False(t, ok)
	assert.Empty(t, permission)
}

func TestBackofficeGraphQLRootFieldsRequirePermissionMapping(t *testing.T) {
	tests := []struct {
		object string
		fields []string
	}{
		{
			object: "Query",
			fields: []string{
				"stores",
				"store",
				"productSetupSnapshot",
				"routedOrders",
				"routedOrderActivities",
				"routedOrderRecommendation",
			},
		},
		{
			object: "Mutation",
			fields: []string{
				"createStore",
				"activateStore",
				"deactivateStore",
				"createProductSetupDraft",
				"promoteProductSetupCandidate",
				"updateProductSetupCandidateStatus",
				"createRoutedOrder",
				"forceRerouteBlockedOrder",
				"advanceRoutedOrder",
				"openOrderException",
				"updateOrderExceptionStatus",
				"updateOrderShipment",
				"updateOrderSettlement",
				"updateOrderIssueHandling",
				"updateOrderQueueControl",
				"bulkUpdateRoutedOrders",
			},
		},
	}

	for _, tt := range tests {
		for _, field := range tt.fields {
			t.Run(tt.object+"."+field, func(t *testing.T) {
				permission, ok := permissionForField(tt.object, field)
				require.True(t, ok)
				assert.NotEmpty(t, permission)
			})
		}
	}
}

func TestRequiresPermissionMapping(t *testing.T) {
	assert.True(t, requiresPermissionMapping("Query"))
	assert.True(t, requiresPermissionMapping("Mutation"))
	assert.False(t, requiresPermissionMapping("Store"))
}

func TestPermissionDeniedErrorIncludesPermissionAndResource(t *testing.T) {
	err := &PermissionDeniedError{
		Permission: "store_config:update",
		Resource:   "podzone:tenant/tenant-1/store/store-1",
	}

	require.Equal(
		t,
		"missing permission: store_config:update on podzone:tenant/tenant-1/store/store-1",
		err.Error(),
	)
}

func TestPermissionMappingErrorIncludesGraphQLField(t *testing.T) {
	err := &PermissionMappingError{Object: "Query", Field: "newField"}

	require.Equal(
		t,
		"permission mapping is missing for GraphQL field: Query.newField",
		err.Error(),
	)
}
