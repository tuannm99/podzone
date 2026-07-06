package store

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	storedomain "github.com/tuannm99/podzone/internal/onboarding/domain/store"
	storeinputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport"
	storemocks "github.com/tuannm99/podzone/internal/onboarding/domain/store/inputport/mocks"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestController_CreateStoreRequestMapsOwnerID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := storemocks.NewMockUsecase(t)
	controller := &Controller{service: usecase}
	router := gin.New()
	controller.RegisterRoutes(router.Group("/onboarding/v1"))

	usecase.EXPECT().
		CreateStoreRequest(
			mock.Anything,
			storeinputport.CreateStoreRequestCommand{
				Name:      "Urban Finds",
				Subdomain: "urban-finds",
				OwnerID:   "tenant-root",
			},
		).
		Return(&storeinputport.Request{
			ID:          "request-1",
			WorkspaceID: "tenant-1",
			Name:        "Urban Finds",
			Subdomain:   "urban-finds",
			RequestedBy: "platform-admin",
			OwnerID:     "tenant-root",
			Status:      storeinputport.RequestStatusQueued,
		}, nil)

	request := httptest.NewRequest(
		http.MethodPost,
		"/onboarding/v1/requests",
		strings.NewReader(`{"name":"Urban Finds","subdomain":"urban-finds","owner_id":"tenant-root"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(toolkit.WithUserID(
		toolkit.WithTenantID(request.Context(), "tenant-1"),
		"platform-admin",
	))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusCreated, response.Code)
	require.Contains(t, response.Body.String(), `"owner_id":"tenant-root"`)
}

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
			collection.Query{
				Page:          2,
				PageSize:      5,
				Search:        "urban",
				Filters:       []collection.Filter{},
				SortDirection: collection.SortDescending,
			},
		).
		Return(collection.NewPage([]*storeinputport.Request{}, 0, collection.Query{Page: 2, PageSize: 5}), nil)

	request := httptest.NewRequest(
		http.MethodGet,
		"/onboarding/v1/requests?collection.page=2&collection.pageSize=5&collection.search=urban",
		nil,
	)
	requestCtx := toolkit.WithTenantID(request.Context(), "tenant-1")
	requestCtx = toolkit.WithUserID(requestCtx, "7")
	request = request.WithContext(requestCtx)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
		"items": [],
		"pageInfo": {
			"total": 0,
			"page": 2,
			"pageSize": 5,
			"totalPages": 0,
			"hasNext": false,
			"hasPrevious": false
		}
	}`, response.Body.String())
}

func TestController_ListStoreRequestsReturnsForbiddenForAccessDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := storemocks.NewMockUsecase(t)
	controller := &Controller{service: usecase}
	router := gin.New()
	controller.RegisterRoutes(router.Group("/onboarding/v1"))

	usecase.EXPECT().
		ListStoreRequests(mock.Anything, "tenant-1", mock.Anything).
		Return(
			collection.Page[*storeinputport.Request]{},
			errors.Join(storedomain.ErrAccessDenied, errors.New("store:read")),
		)

	request := httptest.NewRequest(http.MethodGet, "/onboarding/v1/requests", nil)
	requestCtx := toolkit.WithTenantID(request.Context(), "tenant-1")
	requestCtx = toolkit.WithUserID(requestCtx, "7")
	request = request.WithContext(requestCtx)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusForbidden, response.Code)
	require.Contains(t, response.Body.String(), storedomain.ErrAccessDenied.Error())
}
