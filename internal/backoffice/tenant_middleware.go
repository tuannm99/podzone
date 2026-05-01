package backoffice

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golang-jwt/jwt"
	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

// TenantMiddleware is an app-level GraphQL extension.
type TenantMiddleware struct {
	authCfg    boconfig.RPCConfig
	authorizer TenantAuthorizer
}

type backofficeJWTClaims struct {
	UserID         uint   `json:"user_id"`
	ActiveTenantID string `json:"active_tenant_id"`
	SessionID      string `json:"session_id"`
	Key            string `json:"key"`
	jwt.StandardClaims
}

func NewTenantMiddleware(cfg boconfig.Config, authorizer TenantAuthorizer) *TenantMiddleware {
	return &TenantMiddleware{authCfg: cfg.Auth, authorizer: authorizer}
}

func (m *TenantMiddleware) ExtensionName() string                          { return "TenantMiddleware" }
func (m *TenantMiddleware) Validate(schema graphql.ExecutableSchema) error { return nil }
func (m *TenantMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	return next(ctx)
}

func (m *TenantMiddleware) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	op := graphql.GetOperationContext(ctx)
	if op == nil {
		return next(ctx)
	}

	userID, activeTenantID, sessionID, err := m.identityFromAuthorization(op.Headers.Get("Authorization"))
	if err != nil {
		return func(ctx context.Context) *graphql.Response {
			return graphql.ErrorResponse(ctx, "%s", err.Error())
		}
	}

	tenantID, err := resolveTenantID(activeTenantID)
	if err != nil {
		return func(ctx context.Context) *graphql.Response {
			return graphql.ErrorResponse(ctx, "%s", err.Error())
		}
	}

	if err := m.authorizer.AuthorizeTenant(ctx, sessionID, userID, tenantID); err != nil {
		return func(ctx context.Context) *graphql.Response {
			return graphql.ErrorResponse(ctx, "%s", err.Error())
		}
	}

	ctx = toolkit.WithTenantID(ctx, tenantID)
	ctx = toolkit.WithUserID(ctx, userID)
	return next(ctx)
}

func (m *TenantMiddleware) InterceptField(ctx context.Context, next graphql.Resolver) (res any, err error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil || !fc.IsResolver {
		return next(ctx)
	}

	permission, ok := permissionForField(fc.Object, fc.Field.Name)
	if !ok {
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
	if err := m.authorizer.RequirePermission(ctx, userID, tenantID, permission); err != nil {
		return nil, err
	}
	return next(ctx)
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
		case "storeConfigs", "storeConfig":
			return "store_config:read", true
		}
	case "Mutation":
		switch fieldName {
		case "createStore":
			return "store:create", true
		case "activateStore":
			return "store:activate", true
		case "deactivateStore":
			return "store:deactivate", true
		}
	}
	return "", false
}
