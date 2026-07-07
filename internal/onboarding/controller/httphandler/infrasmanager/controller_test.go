package infrasmanager

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	inputmocks "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestListConnectionsAuthorizesAndForwardsCollectionQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := inputmocks.NewMockUsecase(t)
	authorizer := outputmocks.NewMockAccessAuthorizer(t)
	authorizer.EXPECT().
		AuthorizeInfrastructureRead(mock.Anything, "7").
		Return(nil).
		Once()
	usecase.EXPECT().
		ListConnections(
			mock.Anything,
			"tenant-1",
			false,
			mock.MatchedBy(func(query collection.Query) bool {
				return query.Page == 2 &&
					query.PageSize == 10 &&
					query.Search == "postgres" &&
					query.SortBy == "updatedAt"
			}),
		).
		Return(collection.NewPage([]inputport.Connection{{Name: "primary"}}, 12, collection.Query{
			Page:     2,
			PageSize: 10,
		}), nil).
		Once()

	router := newInfrastructureTestRouter(usecase, authorizer)
	request := httptest.NewRequest(
		http.MethodGet,
		"/onboarding/v1/infras/connections?collection.page=2&collection.pageSize=10"+
			"&collection.search=postgres&collection.sortBy=updatedAt",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.JSONEq(t, `{
		"items": [{"tenant_id":"","infra_type":"","name":"primary","endpoint":"",
			"secret_ref":"","status":"","version":0,"meta":null,"config":null,
			"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}],
		"pageInfo": {"total":12,"page":2,"pageSize":10,"totalPages":2,
			"hasNext":false,"hasPrevious":true}
	}`, response.Body.String())
}

func TestListConnectionsRejectsMissingInfrastructurePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := inputmocks.NewMockUsecase(t)
	authorizer := outputmocks.NewMockAccessAuthorizer(t)
	authorizer.EXPECT().
		AuthorizeInfrastructureRead(mock.Anything, "7").
		Return(entity.ErrAccessDenied).
		Once()

	router := newInfrastructureTestRouter(usecase, authorizer)
	request := httptest.NewRequest(
		http.MethodGet,
		"/onboarding/v1/infras/connections",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusForbidden, response.Code)
	require.JSONEq(t, `{"error":"infrasmanager: access denied"}`, response.Body.String())
}

func TestListDatabaseClustersAuthorizesAndForwardsCollectionQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := inputmocks.NewMockUsecase(t)
	authorizer := outputmocks.NewMockAccessAuthorizer(t)
	authorizer.EXPECT().
		AuthorizeInfrastructureRead(mock.Anything, "7").
		Return(nil).
		Once()
	usecase.EXPECT().
		ListDatabaseClusters(
			mock.Anything,
			mock.MatchedBy(func(query collection.Query) bool {
				return query.Page == 2 &&
					query.PageSize == 5 &&
					query.Search == "postgres"
			}),
		).
		Return(collection.NewPage([]inputport.DatabaseClusterResource{{
			Name:        "pg-primary",
			Engine:      "postgres",
			PlacementDB: "podzone_tenants",
			Status:      "active",
			Healthy:     true,
		}}, 6, collection.Query{Page: 2, PageSize: 5}), nil).
		Once()

	router := newInfrastructureTestRouter(usecase, authorizer)
	request := httptest.NewRequest(
		http.MethodGet,
		"/onboarding/v1/infras/resources/database-clusters?collection.page=2"+
			"&collection.pageSize=5&collection.search=postgres",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"name":"pg-primary"`)
	require.Contains(t, response.Body.String(), `"total":6`)
}

func TestUpsertDatabaseClusterRejectsPathNameMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := inputmocks.NewMockUsecase(t)
	authorizer := outputmocks.NewMockAccessAuthorizer(t)
	authorizer.EXPECT().
		AuthorizeInfrastructureManage(mock.Anything, "7").
		Return(nil).
		Once()

	router := newInfrastructureTestRouter(usecase, authorizer)
	request := httptest.NewRequest(
		http.MethodPut,
		"/onboarding/v1/infras/resources/database-clusters/pg-primary",
		bytes.NewBufferString(`{
			"name":"pg-secondary",
			"engine":"postgres",
			"placement_db":"podzone_tenants"
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{
		"error":"resource_name_mismatch",
		"message":"resource name must match the URL path"
	}`, response.Body.String())
}

func TestCheckDatabaseClusterHealthAuthorizesAndForwardsName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := inputmocks.NewMockUsecase(t)
	authorizer := outputmocks.NewMockAccessAuthorizer(t)
	authorizer.EXPECT().
		AuthorizeInfrastructureManage(mock.Anything, "7").
		Return(nil).
		Once()
	usecase.EXPECT().
		CheckDatabaseClusterHealth(mock.Anything, "pg-default").
		Return(&inputport.DatabaseClusterHealthCheckResponse{
			Name:               "pg-default",
			Healthy:            true,
			CurrentTenants:     2,
			CurrentSchemas:     3,
			CurrentConnections: 4,
			Message:            "ok",
		}, nil).
		Once()

	router := newInfrastructureTestRouter(usecase, authorizer)
	request := httptest.NewRequest(
		http.MethodPost,
		"/onboarding/v1/infras/resources/database-clusters/pg-default/health-check",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"name":"pg-default"`)
	require.Contains(t, response.Body.String(), `"current_schemas":3`)
}

func TestGetTenantPlacementStatusAuthorizesAndForwardsTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := inputmocks.NewMockUsecase(t)
	authorizer := outputmocks.NewMockAccessAuthorizer(t)
	authorizer.EXPECT().
		AuthorizeInfrastructureRead(mock.Anything, "7").
		Return(nil).
		Once()
	usecase.EXPECT().
		GetTenantPlacementStatus(mock.Anything, "workspace-2").
		Return(&inputport.PlacementStatus{
			TenantID:        "workspace-2",
			AllocationReady: true,
			RouteReady:      false,
			NeedsRepair:     true,
			Reason:          "placement route is missing",
		}, nil).
		Once()

	router := newInfrastructureTestRouter(usecase, authorizer)
	request := httptest.NewRequest(
		http.MethodGet,
		"/onboarding/v1/infras/placements/workspace-2/status",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"tenant_id":"workspace-2"`)
	require.Contains(t, response.Body.String(), `"needs_repair":true`)
}

func TestReconcileTenantPlacementReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usecase := inputmocks.NewMockUsecase(t)
	authorizer := outputmocks.NewMockAccessAuthorizer(t)
	authorizer.EXPECT().
		AuthorizeInfrastructureManage(mock.Anything, "7").
		Return(nil).
		Once()
	usecase.EXPECT().
		ReconcileTenantPlacement(mock.Anything, "workspace-2", mock.Anything).
		Return(nil, entity.ErrPlacementNotFound).
		Once()

	router := newInfrastructureTestRouter(usecase, authorizer)
	request := httptest.NewRequest(
		http.MethodPost,
		"/onboarding/v1/infras/placements/workspace-2/reconcile",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNotFound, response.Code, response.Body.String())
	require.JSONEq(t, `{
		"error":"placement_not_found",
		"message":"infrasmanager: placement not found"
	}`, response.Body.String())
}

func newInfrastructureTestRouter(
	usecase inputport.Usecase,
	authorizer *outputmocks.MockAccessAuthorizer,
) *gin.Engine {
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		requestCtx := toolkit.WithTenantID(ctx.Request.Context(), "tenant-1")
		requestCtx = toolkit.WithUserID(requestCtx, "7")
		ctx.Request = ctx.Request.WithContext(requestCtx)
		ctx.Next()
	})
	controller := NewController(ControllerParams{
		InfrasUsecase: usecase,
		Authorizer:    authorizer,
	})
	controller.RegisterRoutes(router.Group("/onboarding/v1"))
	return router
}
