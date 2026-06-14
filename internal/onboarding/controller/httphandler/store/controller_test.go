package store

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storemocks "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport/mocks"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestController_ListStoreRequestsUsesRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := storemocks.NewMockUsecase(t)
	controller := &Controller{service: usecase}
	router := gin.New()
	controller.RegisterRoutes(router.Group("/onboarding/v1"))

	usecase.EXPECT().
		ListStoreRequests(
			mock.MatchedBy(func(ctx context.Context) bool {
				tenantID, tenantErr := toolkit.GetTenantID(ctx)
				userID, userErr := toolkit.GetUserID(ctx)
				return tenantErr == nil &&
					userErr == nil &&
					tenantID == "tenant-1" &&
					userID == "7"
			}),
			"tenant-1",
		).
		Return([]*storeinputport.Request{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/onboarding/v1/requests", nil)
	requestCtx := toolkit.WithTenantID(request.Context(), "tenant-1")
	requestCtx = toolkit.WithUserID(requestCtx, "7")
	request = request.WithContext(requestCtx)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `[]`, response.Body.String())
}
