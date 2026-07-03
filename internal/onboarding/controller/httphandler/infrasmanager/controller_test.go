package infrasmanager

import (
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
