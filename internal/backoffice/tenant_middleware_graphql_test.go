package backoffice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	authconfig "github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	authentity "github.com/tuannm99/podzone/internal/auth/domain/entity"
	backofficeoperations "github.com/tuannm99/podzone/internal/backoffice/application/operations"
	operationsmocks "github.com/tuannm99/podzone/internal/backoffice/application/operations/mocks"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	cataloginputmocks "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/mocks"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	storemocks "github.com/tuannm99/podzone/internal/backoffice/domain/store/mocks"
	backofficemocks "github.com/tuannm99/podzone/internal/backoffice/mocks"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/scope"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/storeaccess"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/tenancy"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestTenantMiddlewareGraphQLInjectsIdentityAndChecksPermission(t *testing.T) {
	tokenUC := authdomain.NewTokenUsecase(authconfig.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	})
	token, err := tokenUC.CreateJwtTokenForSession(authentity.User{
		Id:       12,
		Email:    "owner@podzone.io",
		Username: "owner",
	}, "tenant-ops", "session-1")
	require.NoError(t, err)

	authz := backofficemocks.NewMockTenantAuthorizer(t)
	bootstrapper := backofficemocks.NewMockTenantBootstrapper(t)
	orderUC := operationsmocks.NewMockOrderRoutingUsecase(t)
	storeRepo := storemocks.NewMockStoreRepository(t)
	const storeID = "store-ops"
	authz.EXPECT().AuthorizeTenant(mock.Anything, "session-1", "12", "tenant-ops").Return(nil).Once()
	authz.EXPECT().
		RequirePermission(mock.Anything, "12", "tenant-ops", "store:read", "podzone:tenant/tenant-ops/store/"+storeID).
		Return(nil).
		Once()
	bootstrapper.EXPECT().EnsureReady(mock.Anything, "tenant-ops").Return(nil).Once()
	storeRepo.EXPECT().
		FindByID(mock.Anything, storeID).
		Return(&storectx.Store{ID: storeID, Name: "Ops Store"}, nil).
		Once()
	orderUC.EXPECT().
		ListRoutedOrders(mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			query backofficeoperations.ListRoutedOrdersQuery,
		) ([]routingctx.RoutedOrder, error) {
			tenantID, err := toolkit.GetTenantID(ctx)
			require.NoError(t, err)
			userID, err := toolkit.GetUserID(ctx)
			require.NoError(t, err)
			require.Equal(t, "tenant-ops", tenantID)
			require.Equal(t, "12", userID)
			require.Equal(t, storeID, query.StoreID)
			return []routingctx.RoutedOrder{
				{
					ID:               "ord-1",
					CandidateID:      "cand-1",
					ProductTitle:     "Vintage Tee",
					Partner:          "Print Partner A",
					Quantity:         1,
					Total:            "$20.00",
					CustomerName:     "Alex",
					Status:           routingctx.RoutedOrderStatusQueued,
					ShipmentStatus:   routingctx.RoutedOrderShipmentStatusAwaitingLabel,
					OperatorAssignee: "unassigned",
					BaseCostSnapshot: "$8.00",
					FulfillmentCost:  "$8.00",
					ShippingCost:     "$0.00",
					IssueCost:        "$0.00",
					IssueResolution:  routingctx.RoutedOrderIssueResolutionMonitor,
					RealizedMargin:   "$12.00",
					SettlementStatus: routingctx.RoutedOrderSettlementStatusPending,
				},
			}, nil
		}).
		Once()
	productUC := cataloginputmocks.NewMockProductSetupUsecase(t)

	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: orderUC,
	}, bootstrapper, storeaccess.New(storeRepo))

	rec := doGraphQLRequest(t, srv, "Bearer "+token, map[string]string{
		"X-Store-ID": storeID,
	})
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Empty(t, payload.Errors)
	require.Equal(t, "ord-1", payload.Data.RoutedOrders[0].ID)
}

func TestTenantMiddlewareGraphQLInjectsStoreScope(t *testing.T) {
	tokenUC := authdomain.NewTokenUsecase(authconfig.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	})
	token, err := tokenUC.CreateJwtTokenForSession(authentity.User{
		Id:       12,
		Email:    "owner@podzone.io",
		Username: "owner",
	}, "tenant-ops", "session-1")
	require.NoError(t, err)

	authz := backofficemocks.NewMockTenantAuthorizer(t)
	bootstrapper := backofficemocks.NewMockTenantBootstrapper(t)
	orderUC := operationsmocks.NewMockOrderRoutingUsecase(t)
	storeRepo := storemocks.NewMockStoreRepository(t)
	const storeID = "store-ops"

	authz.EXPECT().AuthorizeTenant(mock.Anything, "session-1", "12", "tenant-ops").Return(nil).Once()
	authz.EXPECT().
		RequirePermission(mock.Anything, "12", "tenant-ops", "store:read", "podzone:tenant/tenant-ops/store/"+storeID).
		Return(nil).
		Once()
	bootstrapper.EXPECT().EnsureReady(mock.Anything, "tenant-ops").Return(nil).Once()
	storeRepo.EXPECT().
		FindByID(mock.Anything, storeID).
		Return(&storectx.Store{ID: storeID, Name: "Ops Store"}, nil).
		Once()
	orderUC.EXPECT().
		ListRoutedOrders(mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			query backofficeoperations.ListRoutedOrdersQuery,
		) ([]routingctx.RoutedOrder, error) {
			currentStoreID := scope.CurrentStoreID(ctx)
			require.Equal(t, storeID, currentStoreID)
			require.Equal(t, storeID, query.StoreID)
			return []routingctx.RoutedOrder{
				{ID: "ord-1", CandidateID: "cand-1", ProductTitle: "Vintage Tee"},
			}, nil
		}).
		Once()
	productUC := cataloginputmocks.NewMockProductSetupUsecase(t)

	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: orderUC,
	}, bootstrapper, storeaccess.New(storeRepo))

	rec := doGraphQLRequest(t, srv, "Bearer "+token, map[string]string{
		"X-Store-ID": storeID,
	})
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Empty(t, payload.Errors)
	require.Equal(t, "ord-1", payload.Data.RoutedOrders[0].ID)
}

func TestTenantMiddlewareGraphQLRejectsMissingAuthorization(t *testing.T) {
	srv := newBackofficeGraphQLTestServer(t, backofficemocks.NewMockTenantAuthorizer(t), &resolver.Resolver{
		ProductSetupUsecase: cataloginputmocks.NewMockProductSetupUsecase(t),
		OrderRoutingUsecase: operationsmocks.NewMockOrderRoutingUsecase(t),
	}, backofficemocks.NewMockTenantBootstrapper(t), nil)

	rec := doGraphQLRequest(t, srv, "", nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Errors, 1)
	assert.Contains(t, payload.Errors[0].Message, "authorization bearer token is required")
}

func TestTenantMiddlewareGraphQLRejectsPermissionDenied(t *testing.T) {
	tokenUC := authdomain.NewTokenUsecase(authconfig.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	})
	token, err := tokenUC.CreateJwtTokenForSession(authentity.User{
		Id:       99,
		Email:    "ops@podzone.io",
		Username: "ops",
	}, "tenant-ops", "session-2")
	require.NoError(t, err)

	authz := backofficemocks.NewMockTenantAuthorizer(t)
	bootstrapper := backofficemocks.NewMockTenantBootstrapper(t)
	authz.EXPECT().AuthorizeTenant(mock.Anything, "session-2", "99", "tenant-ops").Return(nil).Once()
	authz.EXPECT().
		RequirePermission(mock.Anything, "99", "tenant-ops", "store:read", "*").
		Return(&PermissionDeniedError{Permission: "store:read", Resource: "*"}).
		Once()
	bootstrapper.EXPECT().EnsureReady(mock.Anything, "tenant-ops").Return(nil).Once()
	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: cataloginputmocks.NewMockProductSetupUsecase(t),
		OrderRoutingUsecase: operationsmocks.NewMockOrderRoutingUsecase(t),
	}, bootstrapper, nil)

	rec := doGraphQLRequest(t, srv, "Bearer "+token, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Errors, 1)
	assert.Contains(t, payload.Errors[0].Message, "missing permission: store:read")
	assert.Equal(t, graphQLErrorCodeForbidden, payload.Errors[0].Extensions["code"])
	assert.Equal(t, "store:read", payload.Errors[0].Extensions["permission"])
	assert.Equal(t, "*", payload.Errors[0].Extensions["resource"])
}

func TestGraphQLErrorReportsPermissionMappingFailure(t *testing.T) {
	err := graphQLError(
		context.Background(),
		&PermissionMappingError{Object: "Query", Field: "newField"},
	)

	require.NotNil(t, err)
	assert.Equal(t, "permission mapping is missing for GraphQL field: Query.newField", err.Message)
	assert.Equal(t, graphQLErrorCodeInternal, err.Extensions["code"])
	assert.Equal(t, "Query.newField", err.Extensions["field"])
}

func TestTenantMiddlewareGraphQLRejectsUnknownStore(t *testing.T) {
	tokenUC := authdomain.NewTokenUsecase(authconfig.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	})
	token, err := tokenUC.CreateJwtTokenForSession(authentity.User{
		Id:       42,
		Email:    "ops@podzone.io",
		Username: "ops",
	}, "tenant-ops", "session-3")
	require.NoError(t, err)

	authz := backofficemocks.NewMockTenantAuthorizer(t)
	bootstrapper := backofficemocks.NewMockTenantBootstrapper(t)
	storeRepo := storemocks.NewMockStoreRepository(t)
	storeRepo.EXPECT().FindByID(mock.Anything, "missing-store").Return(nil, nil).Once()
	authz.EXPECT().AuthorizeTenant(mock.Anything, "session-3", "42", "tenant-ops").Return(nil).Once()
	bootstrapper.EXPECT().EnsureReady(mock.Anything, "tenant-ops").Return(nil).Once()

	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: cataloginputmocks.NewMockProductSetupUsecase(t),
		OrderRoutingUsecase: operationsmocks.NewMockOrderRoutingUsecase(t),
	}, bootstrapper, storeaccess.New(storeRepo))

	rec := doGraphQLRequest(t, srv, "Bearer "+token, map[string]string{
		"X-Store-ID": "missing-store",
	})
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Errors, 1)
	assert.Contains(t, payload.Errors[0].Message, "store not found")
}

func TestTenantMiddlewareGraphQLRejectsBootstrapFailure(t *testing.T) {
	tokenUC := authdomain.NewTokenUsecase(authconfig.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	})
	token, err := tokenUC.CreateJwtTokenForSession(authentity.User{
		Id:       42,
		Email:    "ops@podzone.io",
		Username: "ops",
	}, "tenant-ops", "session-3")
	require.NoError(t, err)

	authz := backofficemocks.NewMockTenantAuthorizer(t)
	bootstrapper := backofficemocks.NewMockTenantBootstrapper(t)
	authz.EXPECT().AuthorizeTenant(mock.Anything, "session-3", "42", "tenant-ops").Return(nil).Once()
	bootstrapper.EXPECT().EnsureReady(mock.Anything, "tenant-ops").Return(errors.New("tenant bootstrap failed")).Once()
	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: cataloginputmocks.NewMockProductSetupUsecase(t),
		OrderRoutingUsecase: operationsmocks.NewMockOrderRoutingUsecase(t),
	}, bootstrapper, nil)

	rec := doGraphQLRequest(t, srv, "Bearer "+token, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Errors, 1)
	assert.Contains(t, payload.Errors[0].Message, "tenant bootstrap failed")
}

func newBackofficeGraphQLTestServer(
	t *testing.T,
	authz TenantAuthorizer,
	r *resolver.Resolver,
	bootstrapper TenantBootstrapper,
	storeAccess storeaccess.Access,
) *handler.Server {
	t.Helper()

	schema := generated.NewExecutableSchema(generated.Config{Resolvers: r})
	srv := handler.New(schema)
	srv.SetErrorPresenter(graphQLError)
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](100))
	srv.Use(extension.Introspection{})
	srv.Use(NewTenantMiddleware(boconfig.Config{
		Auth: boconfig.RPCConfig{
			JWTSecret: "secret",
			JWTKey:    "app-key",
		},
	}, authz, tenancy.New(bootstrapper, storeAccess, nil)))
	return srv
}

func doGraphQLRequest(
	t *testing.T,
	srv http.Handler,
	authHeader string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"query": "query { routedOrders { id } }",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

type graphQLResponse struct {
	Data struct {
		RoutedOrders []struct {
			ID string `json:"id"`
		} `json:"routedOrders"`
	} `json:"data"`
	Errors []struct {
		Message    string         `json:"message"`
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
}
