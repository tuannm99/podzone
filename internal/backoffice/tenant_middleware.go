package backoffice

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

// TenantMiddleware is an app-level GraphQL extension.
type TenantMiddleware struct{}

func (m *TenantMiddleware) ExtensionName() string                          { return "TenantMiddleware" }
func (m *TenantMiddleware) Validate(schema graphql.ExecutableSchema) error { return nil }
func (m *TenantMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	op := graphql.GetOperationContext(ctx)
	if op == nil {
		return next(ctx)
	}

	tenantID := op.Headers.Get("X-Tenant-ID")
	if tenantID == "" {
		return graphql.ErrorResponse(ctx, "tenant_id is required")
	}

	ctx = toolkit.WithTenantID(ctx, tenantID)
	return next(ctx)
}
