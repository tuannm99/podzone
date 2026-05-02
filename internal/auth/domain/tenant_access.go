package domain

import "context"

type TenantAccessChecker interface {
	EnsureActiveMembership(ctx context.Context, tenantID string, userID uint) error
}
