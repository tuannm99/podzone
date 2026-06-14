package httphandler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	storemocks "github.com/tuannm99/podzone/internal/backoffice/domain/store/mocks"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestStoreBootstrapHandler_BootstrapsStoreWithTenantScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stores := storemocks.NewMockStoreUsecase(t)
	handler := NewStoreBootstrapHandler(
		boconfig.Config{InternalServiceToken: "service-token"},
		stores,
	)
	router := gin.New()
	handler.RegisterRoutes()(router)

	stores.EXPECT().
		BootstrapStore(
			mock.MatchedBy(func(ctx context.Context) bool {
				workspaceID, workspaceErr := toolkit.GetTenantID(ctx)
				ownerID, ownerErr := toolkit.GetUserID(ctx)
				return workspaceErr == nil &&
					ownerErr == nil &&
					workspaceID == "workspace-1" &&
					ownerID == "user-1"
			}),
			storectx.BootstrapStoreCmd{
				ID:      "store-1",
				Name:    "Urban Finds",
				OwnerID: "user-1",
			},
		).
		Return(&storectx.Store{
			ID:      "store-1",
			Name:    "Urban Finds",
			OwnerID: "user-1",
			Status:  storectx.StoreStatusActive,
		}, nil)

	request := httptest.NewRequest(
		http.MethodPost,
		"/internal/backoffice/v1/stores:bootstrap",
		bytes.NewBufferString(
			`{"workspace_id":"workspace-1","store_id":"store-1","name":"Urban Finds","owner_id":"user-1"}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Backoffice-Service-Token", "service-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(
		t,
		`{"id":"store-1","name":"Urban Finds","owner_id":"user-1","status":"active",`+
			`"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		response.Body.String(),
	)
}

func TestStoreBootstrapHandler_RejectsInvalidServiceToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stores := storemocks.NewMockStoreUsecase(t)
	handler := NewStoreBootstrapHandler(
		boconfig.Config{InternalServiceToken: "service-token"},
		stores,
	)
	router := gin.New()
	handler.RegisterRoutes()(router)

	request := httptest.NewRequest(
		http.MethodPost,
		"/internal/backoffice/v1/stores:bootstrap",
		bytes.NewBufferString(`{}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Backoffice-Service-Token", "wrong-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
}
