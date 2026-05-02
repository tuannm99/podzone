package auth

import (
	"context"

	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type iamTenantAccessChecker struct {
	iamUC iamdomain.IAMUsecase
}

func NewTenantAccessChecker(iamUC iamdomain.IAMUsecase) authdomain.TenantAccessChecker {
	return &iamTenantAccessChecker{iamUC: iamUC}
}

func (c *iamTenantAccessChecker) EnsureActiveMembership(ctx context.Context, tenantID string, userID uint) error {
	membership, err := c.iamUC.GetMembership(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if membership.Status != iamdomain.MembershipStatusActive {
		return iamdomain.ErrInactiveMembership
	}
	return nil
}
