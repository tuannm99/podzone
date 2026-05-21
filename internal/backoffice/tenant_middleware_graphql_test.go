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
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	inputmocks "github.com/tuannm99/podzone/internal/backoffice/domain/inputport/mocks"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	backofficemocks "github.com/tuannm99/podzone/internal/backoffice/mocks"
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
	orderUC := inputmocks.NewMockOrderRoutingUsecase(t)
	authz.EXPECT().AuthorizeTenant(mock.Anything, "session-1", "12", "tenant-ops").Return(nil).Once()
	authz.EXPECT().RequirePermission(mock.Anything, "12", "tenant-ops", "store:read").Return(nil).Once()
	bootstrapper.EXPECT().EnsureReady(mock.Anything, "tenant-ops").Return(nil).Once()
	orderUC.EXPECT().
		ListRoutedOrders(mock.Anything).
		RunAndReturn(func(ctx context.Context) ([]routingentity.RoutedOrder, error) {
			tenantID, err := toolkit.GetTenantID(ctx)
			require.NoError(t, err)
			userID, err := toolkit.GetUserID(ctx)
			require.NoError(t, err)
			require.Equal(t, "tenant-ops", tenantID)
			require.Equal(t, "12", userID)
			return []routingentity.RoutedOrder{
				{
					ID:               "ord-1",
					CandidateID:      "cand-1",
					ProductTitle:     "Vintage Tee",
					Partner:          "Print Partner A",
					Quantity:         1,
					Total:            "$20.00",
					CustomerName:     "Alex",
					Status:           routingentity.RoutedOrderStatusQueued,
					ShipmentStatus:   routingentity.RoutedOrderShipmentStatusAwaitingLabel,
					OperatorAssignee: "unassigned",
					BaseCostSnapshot: "$8.00",
					FulfillmentCost:  "$8.00",
					ShippingCost:     "$0.00",
					IssueCost:        "$0.00",
					IssueResolution:  routingentity.RoutedOrderIssueResolutionMonitor,
					RealizedMargin:   "$12.00",
					SettlementStatus: routingentity.RoutedOrderSettlementStatusPending,
				},
			}, nil
		}).
		Once()
	productUC := inputmocks.NewMockProductSetupUsecase(t)

	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: orderUC,
	}, bootstrapper)

	rec := doGraphQLRequest(t, srv, "query { routedOrders { id } }", "Bearer "+token)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Empty(t, payload.Errors)
	require.Equal(t, "ord-1", payload.Data.RoutedOrders[0].ID)
}

func TestTenantMiddlewareGraphQLRejectsMissingAuthorization(t *testing.T) {
	srv := newBackofficeGraphQLTestServer(t, backofficemocks.NewMockTenantAuthorizer(t), &resolver.Resolver{
		ProductSetupUsecase: inputmocks.NewMockProductSetupUsecase(t),
		OrderRoutingUsecase: inputmocks.NewMockOrderRoutingUsecase(t),
	}, backofficemocks.NewMockTenantBootstrapper(t))

	rec := doGraphQLRequest(t, srv, "query { routedOrders { id } }", "")
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
		RequirePermission(mock.Anything, "99", "tenant-ops", "store:read").
		Return(errors.New("permission denied")).
		Once()
	bootstrapper.EXPECT().EnsureReady(mock.Anything, "tenant-ops").Return(nil).Once()
	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: inputmocks.NewMockProductSetupUsecase(t),
		OrderRoutingUsecase: inputmocks.NewMockOrderRoutingUsecase(t),
	}, bootstrapper)

	rec := doGraphQLRequest(t, srv, "query { routedOrders { id } }", "Bearer "+token)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Errors, 1)
	assert.Contains(t, payload.Errors[0].Message, "permission denied")
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
		ProductSetupUsecase: inputmocks.NewMockProductSetupUsecase(t),
		OrderRoutingUsecase: inputmocks.NewMockOrderRoutingUsecase(t),
	}, bootstrapper)

	rec := doGraphQLRequest(t, srv, "query { routedOrders { id } }", "Bearer "+token)
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
) *handler.Server {
	t.Helper()

	schema := generated.NewExecutableSchema(generated.Config{Resolvers: r})
	srv := handler.New(schema)
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](100))
	srv.Use(extension.Introspection{})
	srv.Use(NewTenantMiddleware(boconfig.Config{
		Auth: boconfig.RPCConfig{
			JWTSecret: "secret",
			JWTKey:    "app-key",
		},
	}, authz, bootstrapper))
	return srv
}

func doGraphQLRequest(t *testing.T, srv http.Handler, query string, authHeader string) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"query": query,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
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
		Message string `json:"message"`
	} `json:"errors"`
}
