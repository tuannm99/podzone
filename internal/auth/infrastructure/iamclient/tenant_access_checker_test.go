package iamclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	outputportmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
)

func TestTenantAccessCheckerEnsureActiveMembership_UsesProjectionFirst(t *testing.T) {
	projection := outputportmocks.NewMockIAMProjectionRepository(t)
	projection.EXPECT().
		GetTenantMembership(mock.Anything, "t1", uint(7)).
		Return(&outputport.TenantMembershipProjection{
			TenantID: "t1",
			UserID:   7,
			RoleName: "tenant_viewer",
			Status:   entity.MembershipStatusActive,
		}, nil).
		Once()

	checker := &TenantAccessChecker{projection: projection}
	require.NoError(t, checker.EnsureActiveMembership(context.Background(), "t1", 7))
}

func TestTenantAccessCheckerEnsureActiveMembership_ProjectionInactive(t *testing.T) {
	projection := outputportmocks.NewMockIAMProjectionRepository(t)
	projection.EXPECT().
		GetTenantMembership(mock.Anything, "t1", uint(7)).
		Return(&outputport.TenantMembershipProjection{
			TenantID: "t1",
			UserID:   7,
			RoleName: "tenant_viewer",
			Status:   "revoked",
		}, nil).
		Once()

	checker := &TenantAccessChecker{projection: projection}
	require.ErrorIs(t, checker.EnsureActiveMembership(context.Background(), "t1", 7), entity.ErrInactiveMembership)
}
