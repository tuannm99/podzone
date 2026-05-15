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
	"github.com/stretchr/testify/require"
	authconfig "github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	authentity "github.com/tuannm99/podzone/internal/auth/domain/entity"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated"
	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/resolver"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
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

	authz := &graphqlAuthorizerFake{}
	orderUC := &graphqlOrderRoutingUsecase{
		listRoutedOrdersFn: func(ctx context.Context) ([]entity.RoutedOrder, error) {
			tenantID, err := toolkit.GetTenantID(ctx)
			require.NoError(t, err)
			userID, err := toolkit.GetUserID(ctx)
			require.NoError(t, err)
			require.Equal(t, "tenant-ops", tenantID)
			require.Equal(t, "12", userID)
			return []entity.RoutedOrder{
				{ID: "ord-1", CandidateID: "cand-1", ProductTitle: "Vintage Tee", Partner: "Print Partner A", Quantity: 1, Total: "$20.00", CustomerName: "Alex", Status: entity.RoutedOrderStatusQueued, ShipmentStatus: entity.RoutedOrderShipmentStatusAwaitingLabel, OperatorAssignee: "unassigned", BaseCostSnapshot: "$8.00", FulfillmentCost: "$8.00", ShippingCost: "$0.00", IssueCost: "$0.00", IssueResolution: entity.RoutedOrderIssueResolutionMonitor, RealizedMargin: "$12.00", SettlementStatus: entity.RoutedOrderSettlementStatusPending},
			}, nil
		},
	}

	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: &graphqlProductSetupUsecase{},
		OrderRoutingUsecase: orderUC,
	})

	rec := doGraphQLRequest(t, srv, "query { routedOrders { id } }", "Bearer "+token)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Empty(t, payload.Errors)
	require.Equal(t, "ord-1", payload.Data.RoutedOrders[0].ID)
	assert.Equal(t, []graphqlAuthorizeTenantCall{{
		SessionID: "session-1",
		UserID:    "12",
		TenantID:  "tenant-ops",
	}}, authz.authorizeCalls)
	assert.Equal(t, []graphqlRequirePermissionCall{{
		UserID:     "12",
		TenantID:   "tenant-ops",
		Permission: "store:read",
	}}, authz.permissionCalls)
}

func TestTenantMiddlewareGraphQLRejectsMissingAuthorization(t *testing.T) {
	srv := newBackofficeGraphQLTestServer(t, &graphqlAuthorizerFake{}, &resolver.Resolver{
		ProductSetupUsecase: &graphqlProductSetupUsecase{},
		OrderRoutingUsecase: &graphqlOrderRoutingUsecase{},
	})

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

	authz := &graphqlAuthorizerFake{
		requirePermissionErr: errors.New("permission denied"),
	}
	srv := newBackofficeGraphQLTestServer(t, authz, &resolver.Resolver{
		ProductSetupUsecase: &graphqlProductSetupUsecase{},
		OrderRoutingUsecase: &graphqlOrderRoutingUsecase{},
	})

	rec := doGraphQLRequest(t, srv, "query { routedOrders { id } }", "Bearer "+token)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload graphQLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Errors, 1)
	assert.Contains(t, payload.Errors[0].Message, "permission denied")
	assert.Equal(t, []graphqlRequirePermissionCall{{
		UserID:     "99",
		TenantID:   "tenant-ops",
		Permission: "store:read",
	}}, authz.permissionCalls)
}

func newBackofficeGraphQLTestServer(t *testing.T, authz TenantAuthorizer, r *resolver.Resolver) *handler.Server {
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
	}, authz))
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

type graphqlAuthorizerFake struct {
	authorizeTenantErr   error
	requirePermissionErr error
	authorizeCalls       []graphqlAuthorizeTenantCall
	permissionCalls      []graphqlRequirePermissionCall
}

type graphqlAuthorizeTenantCall struct {
	SessionID string
	UserID    string
	TenantID  string
}

type graphqlRequirePermissionCall struct {
	UserID     string
	TenantID   string
	Permission string
}

func (f *graphqlAuthorizerFake) AuthorizeTenant(_ context.Context, sessionID, userID, tenantID string) error {
	f.authorizeCalls = append(f.authorizeCalls, graphqlAuthorizeTenantCall{
		SessionID: sessionID,
		UserID:    userID,
		TenantID:  tenantID,
	})
	return f.authorizeTenantErr
}

func (f *graphqlAuthorizerFake) RequirePermission(_ context.Context, userID, tenantID, permission string) error {
	f.permissionCalls = append(f.permissionCalls, graphqlRequirePermissionCall{
		UserID:     userID,
		TenantID:   tenantID,
		Permission: permission,
	})
	return f.requirePermissionErr
}

type graphqlProductSetupUsecase struct{}

func (graphqlProductSetupUsecase) GetSnapshot(context.Context) (*entity.ProductSetupSnapshot, error) {
	return &entity.ProductSetupSnapshot{}, nil
}

func (graphqlProductSetupUsecase) CreateDraft(context.Context, inputport.CreateProductSetupDraftCmd) (*entity.ProductSetupDraft, error) {
	return nil, errors.New("unexpected CreateDraft call")
}

func (graphqlProductSetupUsecase) PromoteCandidate(context.Context, inputport.PromoteProductSetupCandidateCmd) (*entity.ProductSetupCandidate, error) {
	return nil, errors.New("unexpected PromoteCandidate call")
}

func (graphqlProductSetupUsecase) UpdateCandidateStatus(context.Context, string, string) (*entity.ProductSetupCandidate, error) {
	return nil, errors.New("unexpected UpdateCandidateStatus call")
}

type graphqlOrderRoutingUsecase struct {
	listRoutedOrdersFn func(ctx context.Context) ([]entity.RoutedOrder, error)
}

func (f *graphqlOrderRoutingUsecase) ListRoutedOrders(ctx context.Context) ([]entity.RoutedOrder, error) {
	if f.listRoutedOrdersFn != nil {
		return f.listRoutedOrdersFn(ctx)
	}
	return nil, nil
}

func (graphqlOrderRoutingUsecase) ListRoutedOrderActivities(context.Context, inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error) {
	return nil, errors.New("unexpected ListRoutedOrderActivities call")
}

func (graphqlOrderRoutingUsecase) CreateRoutedOrder(context.Context, inputport.CreateRoutedOrderCmd) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected CreateRoutedOrder call")
}

func (graphqlOrderRoutingUsecase) AdvanceRoutedOrder(context.Context, string) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected AdvanceRoutedOrder call")
}

func (graphqlOrderRoutingUsecase) OpenOrderException(context.Context, inputport.OpenOrderExceptionCmd) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected OpenOrderException call")
}

func (graphqlOrderRoutingUsecase) UpdateOrderExceptionStatus(context.Context, inputport.UpdateOrderExceptionStatusCmd) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected UpdateOrderExceptionStatus call")
}

func (graphqlOrderRoutingUsecase) UpdateOrderShipment(context.Context, inputport.UpdateOrderShipmentCmd) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected UpdateOrderShipment call")
}

func (graphqlOrderRoutingUsecase) UpdateOrderSettlement(context.Context, inputport.UpdateOrderSettlementCmd) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected UpdateOrderSettlement call")
}

func (graphqlOrderRoutingUsecase) UpdateOrderIssueHandling(context.Context, inputport.UpdateOrderIssueHandlingCmd) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected UpdateOrderIssueHandling call")
}

func (graphqlOrderRoutingUsecase) UpdateOrderQueueControl(context.Context, inputport.UpdateOrderQueueControlCmd) (*entity.RoutedOrder, error) {
	return nil, errors.New("unexpected UpdateOrderQueueControl call")
}

func (graphqlOrderRoutingUsecase) BulkUpdateRoutedOrders(context.Context, inputport.BulkUpdateRoutedOrdersCmd) ([]entity.RoutedOrder, error) {
	return nil, errors.New("unexpected BulkUpdateRoutedOrders call")
}
