package grpchandler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	authentity "github.com/tuannm99/podzone/internal/auth/domain/entity"
	authoutputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	iamentity "github.com/tuannm99/podzone/internal/iam/entity"
	iaminputport "github.com/tuannm99/podzone/internal/iam/inputport"
	iammocks "github.com/tuannm99/podzone/internal/iam/inputport/mocks"
)

var testIAMCfg = config.AuthConfig{
	JWTSecret:      "secret",
	JWTKey:         "app-key",
	AppRedirectURL: "https://app.example.com/auth/google/callback",
}

type iamUsecaseMockConfig struct {
	createTenantFunc      func(ctx context.Context, ownerUserID uint, cmd iamentity.CreateTenantCmd) (*iamentity.Tenant, error)
	getMembershipFunc     func(ctx context.Context, tenantID string, userID uint) (*iamentity.Membership, error)
	checkPermissionFunc   func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
	listUserTenantsFunc   func(ctx context.Context, userID uint) ([]iamentity.Membership, error)
	createPolicyFunc      func(ctx context.Context, input iamentity.CreatePolicyInput) (*iamentity.Policy, []iamentity.PolicyStatement, error)
	assumeRoleFunc        func(ctx context.Context, input iamentity.AssumeRoleInput) (*iamentity.AssumedRole, error)
	requirePlatformPermissionFunc func(ctx context.Context, userID uint, permission string) error
}

func newIAMUsecaseMock(t *testing.T, cfg iamUsecaseMockConfig) *iammocks.MockIAMUsecase {
	t.Helper()
	iamUC := iammocks.NewMockIAMUsecase(t)
	if cfg.createTenantFunc != nil {
		iamUC.EXPECT().CreateTenant(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.createTenantFunc).Maybe()
	}
	if cfg.getMembershipFunc != nil {
		iamUC.EXPECT().GetMembership(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.getMembershipFunc).Maybe()
	}
	if cfg.checkPermissionFunc != nil {
		iamUC.EXPECT().CheckPermission(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.checkPermissionFunc).Maybe()
	}
	if cfg.listUserTenantsFunc != nil {
		iamUC.EXPECT().ListUserTenants(mock.Anything, mock.Anything).RunAndReturn(cfg.listUserTenantsFunc).Maybe()
	}
	if cfg.createPolicyFunc != nil {
		iamUC.EXPECT().CreatePolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.createPolicyFunc).Maybe()
	}
	if cfg.assumeRoleFunc != nil {
		iamUC.EXPECT().AssumeRole(mock.Anything, mock.Anything).RunAndReturn(cfg.assumeRoleFunc).Maybe()
	}
	if cfg.requirePlatformPermissionFunc != nil {
		iamUC.EXPECT().RequirePlatformPermission(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.requirePlatformPermissionFunc).Maybe()
	}
	return iamUC
}

func newIAMServerForTest(t *testing.T, iamUC iaminputport.IAMUsecase) *IAMServer {
	t.Helper()
	auditRepo := authoutputmocks.NewMockAuditLogRepository(t)
	auditRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Maybe()
	return NewIAMServer(
		iamUC,
		auditRepo,
		authoutputmocks.NewMockUserRepository(t),
		testIAMCfg,
	)
}

func authContextForIAMUser(t *testing.T, userID uint) context.Context {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testIAMCfg).
		CreateJwtTokenForSession(authentity.User{Id: userID}, "", "session-test")
	require.NoError(t, err)
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func rawAccessTokenForIAMUser(t *testing.T, userID uint) string {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testIAMCfg).
		CreateJwtTokenForSession(authentity.User{Id: userID}, "", "session-test")
	require.NoError(t, err)
	return token
}

func nowPlusHour() time.Time {
	return time.Now().UTC().Add(time.Hour)
}
