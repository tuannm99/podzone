package backoffice

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golang-jwt/jwt/v5"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/scope"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/tenancy"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"google.golang.org/grpc/metadata"
)

// TenantMiddleware is an app-level GraphQL extension.
type TenantMiddleware struct {
	authCfg    boconfig.RPCConfig
	authorizer TenantAuthorizer
	tenancy    tenancy.Runtime
}

type backofficeJWTClaims struct {
	UserID         uint   `json:"user_id"`
	ActiveTenantID string `json:"active_tenant_id"`
	SessionID      string `json:"session_id"`
	Key            string `json:"key"`
	jwt.RegisteredClaims
}

func NewTenantMiddleware(
	cfg boconfig.Config,
	authorizer TenantAuthorizer,
	tenancyRuntime tenancy.Runtime,
) *TenantMiddleware {
	return &TenantMiddleware{
		authCfg:    cfg.Auth,
		authorizer: authorizer,
		tenancy:    tenancyRuntime,
	}
}

func (m *TenantMiddleware) ExtensionName() string                          { return "TenantMiddleware" }
func (m *TenantMiddleware) Validate(schema graphql.ExecutableSchema) error { return nil }
func (m *TenantMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	return next(ctx)
}

func (m *TenantMiddleware) InterceptOperation(
	ctx context.Context,
	next graphql.OperationHandler,
) graphql.ResponseHandler {
	op := graphql.GetOperationContext(ctx)
	if op == nil {
		return next(ctx)
	}

	authHeader := op.Headers.Get("Authorization")
	userID, activeTenantID, sessionID, err := m.identityFromAuthorization(authHeader)
	if err != nil {
		return func(ctx context.Context) *graphql.Response {
			return graphQLErrorResponse(ctx, err)
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authHeader)

	tenantID, err := resolveTenantID(activeTenantID)
	if err != nil {
		return func(ctx context.Context) *graphql.Response {
			return graphQLErrorResponse(ctx, err)
		}
	}

	if err := m.authorizer.AuthorizeTenant(ctx, sessionID, userID, tenantID); err != nil {
		return func(ctx context.Context) *graphql.Response {
			return graphQLErrorResponse(ctx, err)
		}
	}
	ctx = toolkit.WithTenantID(ctx, tenantID)
	ctx = toolkit.WithUserID(ctx, userID)

	storeID := strings.TrimSpace(op.Headers.Get("X-Store-ID"))
	requestScope, err := m.tenancy.ResolveRequestScope(ctx, tenantID, storeID)
	if err != nil {
		return func(ctx context.Context) *graphql.Response {
			return graphQLErrorResponse(ctx, err)
		}
	}

	tenantCtx := scope.TenantContext{
		TenantID:  requestScope.TenantID,
		UserID:    userID,
		SessionID: sessionID,
	}
	if requestScope.Placement != nil {
		tenantCtx.ClusterName = requestScope.Placement.ClusterName
		tenantCtx.DBName = requestScope.Placement.DBName
		tenantCtx.SchemaName = requestScope.Placement.SchemaName
		tenantCtx.PlacementMode = string(requestScope.Placement.Mode)
	}
	ctx = scope.WithTenantContext(ctx, tenantCtx)
	if requestScope.Store != nil {
		ctx = scope.WithStoreContext(ctx, scope.StoreContext{StoreID: requestScope.Store.ID})
	}
	return next(ctx)
}

func (m *TenantMiddleware) InterceptField(ctx context.Context, next graphql.Resolver) (res any, err error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil || !fc.IsResolver {
		return next(ctx)
	}

	permission, ok := permissionForField(fc.Object, fc.Field.Name)
	if !ok {
		if requiresPermissionMapping(fc.Object) {
			return nil, &PermissionMappingError{
				Object: fc.Object,
				Field:  fc.Field.Name,
			}
		}
		return next(ctx)
	}

	userID, err := toolkit.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.authorizer.RequirePermission(
		ctx,
		userID,
		tenantID,
		permission,
		resourceForPermission(ctx, tenantID),
	); err != nil {
		return nil, err
	}
	return next(ctx)
}

func requiresPermissionMapping(objectName string) bool {
	return objectName == "Query" || objectName == "Mutation"
}

func (m *TenantMiddleware) identityFromAuthorization(header string) (string, string, string, error) {
	if header == "" {
		return "", "", "", fmt.Errorf("authorization bearer token is required")
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", "", "", fmt.Errorf("authorization header must use bearer token")
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	if tokenStr == "" {
		return "", "", "", fmt.Errorf("authorization bearer token is required")
	}

	claims := &backofficeJWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(m.authCfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", "", fmt.Errorf("invalid authorization token")
	}
	if m.authCfg.JWTKey != "" && claims.Key != m.authCfg.JWTKey {
		return "", "", "", fmt.Errorf("invalid authorization token")
	}
	if claims.UserID == 0 {
		return "", "", "", fmt.Errorf("authorization token missing user_id")
	}
	if claims.SessionID == "" {
		return "", "", "", fmt.Errorf("authorization token missing session_id")
	}
	return strconv.FormatUint(uint64(claims.UserID), 10), claims.ActiveTenantID, claims.SessionID, nil
}

func resolveTenantID(claimTenantID string) (string, error) {
	claimTenantID = strings.TrimSpace(claimTenantID)
	if claimTenantID == "" {
		return "", fmt.Errorf("authorization token missing active_tenant_id")
	}
	return claimTenantID, nil
}

func permissionForField(objectName, fieldName string) (string, bool) {
	switch objectName {
	case "Query":
		switch fieldName {
		case "stores", "store":
			return "store:read", true
		case "productSetupSnapshot":
			return "store_config:read", true
		case "routedOrders", "routedOrderActivities", "routedOrderRecommendation":
			return "store:read", true
		}
	case "Mutation":
		switch fieldName {
		case "createStore":
			return "store:create", true
		case "activateStore":
			return "store:activate", true
		case "deactivateStore":
			return "store:deactivate", true
		case "createProductSetupDraft", "promoteProductSetupCandidate", "updateProductSetupCandidateStatus":
			return "store_config:update", true
		case "createRoutedOrder",
			"forceRerouteBlockedOrder",
			"advanceRoutedOrder",
			"openOrderException",
			"updateOrderExceptionStatus",
			"updateOrderShipment",
			"updateOrderSettlement",
			"updateOrderIssueHandling",
			"updateOrderQueueControl",
			"bulkUpdateRoutedOrders":
			return "store:update", true
		}
	}
	return "", false
}

func resourceForPermission(ctx context.Context, tenantID string) string {
	storeID := strings.TrimSpace(scope.CurrentStoreID(ctx))
	if storeID == "" {
		return "*"
	}
	return "podzone:tenant/" + strings.TrimSpace(tenantID) + "/store/" + storeID
}
