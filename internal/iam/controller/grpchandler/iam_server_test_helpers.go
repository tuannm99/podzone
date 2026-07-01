package grpchandler

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	iamentity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	iaminputport "github.com/tuannm99/podzone/internal/iam/domain/inputport"
	iammocks "github.com/tuannm99/podzone/internal/iam/domain/inputport/mocks"
	iamoutputmocks "github.com/tuannm99/podzone/internal/iam/domain/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/pdauthn"
)

var testIAMServerCfg = iamconfig.ServerConfig{
	Authn: pdauthn.Config{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	},
	AppRedirectURL: "https://app.example.com/auth/google/callback",
}

type iamUsecaseMockConfig struct {
	createTenantFunc func(
		ctx context.Context,
		ownerUserID uint,
		cmd iamentity.CreateTenantCmd,
	) (*iamentity.Tenant, error)
	getMembershipFunc   func(ctx context.Context, tenantID string, userID uint) (*iamentity.Membership, error)
	checkPermissionFunc func(
		ctx context.Context,
		tenantID string,
		userID uint,
		permission string,
	) (bool, error)
	checkPermissionResourceFunc func(
		ctx context.Context,
		tenantID string,
		userID uint,
		permission string,
		resource string,
	) (bool, error)
	listUserTenantsFunc func(ctx context.Context, userID uint) ([]iamentity.Membership, error)
	createPolicyFunc    func(
		ctx context.Context,
		input iamentity.CreatePolicyInput,
	) (*iamentity.Policy, []iamentity.PolicyStatement, error)
	assumeRoleFunc func(
		ctx context.Context,
		input iamentity.AssumeRoleInput,
	) (*iamentity.AssumedRole, error)
	checkPlatformPermissionFunc   func(ctx context.Context, userID uint, permission string) (bool, error)
	requirePlatformPermissionFunc func(
		ctx context.Context,
		userID uint,
		permission string,
	) error
	listOrganizationsFunc func(
		ctx context.Context,
		query collection.Query,
	) (collection.Page[iamentity.Organization], error)
	listPoliciesFunc func(
		ctx context.Context,
		scope string,
		query collection.Query,
	) (collection.Page[iamentity.Policy], error)
	listGroupsFunc func(
		ctx context.Context,
		scope string,
		tenantID string,
		query collection.Query,
	) (collection.Page[iamentity.Group], error)
}

type iamUsecaseMocks struct {
	commands iaminputport.IAMCommandUsecase
	queries  iaminputport.IAMQueryUsecase
}

func newIAMUsecaseMock(t *testing.T, cfg iamUsecaseMockConfig) iamUsecaseMocks {
	t.Helper()
	commands := iammocks.NewMockIAMCommandUsecase(t)
	queries := iammocks.NewMockIAMQueryUsecase(t)
	if cfg.createTenantFunc != nil {
		commands.EXPECT().
			CreateTenant(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.createTenantFunc).
			Maybe()
	}
	if cfg.getMembershipFunc != nil {
		queries.EXPECT().
			GetMembership(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.getMembershipFunc).
			Maybe()
	}
	if cfg.checkPermissionFunc != nil {
		queries.EXPECT().
			CheckPermission(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.checkPermissionFunc).
			Maybe()
		queries.EXPECT().
			CheckPermissionForResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, tenantID string, userID uint, permission string, _ string) (bool, error) {
				return cfg.checkPermissionFunc(ctx, tenantID, userID, permission)
			}).
			Maybe()
	}
	if cfg.checkPermissionResourceFunc != nil {
		queries.EXPECT().
			CheckPermissionForResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.checkPermissionResourceFunc).
			Maybe()
	}
	if cfg.listUserTenantsFunc != nil {
		queries.EXPECT().ListUserTenants(mock.Anything, mock.Anything).RunAndReturn(cfg.listUserTenantsFunc).Maybe()
	}
	if cfg.createPolicyFunc != nil {
		commands.EXPECT().CreatePolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.createPolicyFunc).Maybe()
	}
	if cfg.assumeRoleFunc != nil {
		commands.EXPECT().AssumeRole(mock.Anything, mock.Anything).RunAndReturn(cfg.assumeRoleFunc).Maybe()
	}
	if cfg.checkPlatformPermissionFunc != nil {
		queries.EXPECT().
			CheckPlatformPermission(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.checkPlatformPermissionFunc).
			Maybe()
	}
	if cfg.requirePlatformPermissionFunc != nil {
		queries.EXPECT().
			RequirePlatformPermission(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.requirePlatformPermissionFunc).
			Maybe()
	}
	if cfg.listOrganizationsFunc != nil {
		queries.EXPECT().
			ListOrganizations(mock.Anything, mock.Anything).
			RunAndReturn(cfg.listOrganizationsFunc).
			Maybe()
	}
	if cfg.listPoliciesFunc != nil {
		queries.EXPECT().
			ListPolicies(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.listPoliciesFunc).
			Maybe()
	}
	if cfg.listGroupsFunc != nil {
		queries.EXPECT().
			ListGroups(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.listGroupsFunc).
			Maybe()
	}
	return iamUsecaseMocks{commands: commands, queries: queries}
}

func newIAMServerForTest(t *testing.T, usecases iamUsecaseMocks) *IAMServer {
	t.Helper()
	auditRepo := iamoutputmocks.NewMockAuditLogRepository(t)
	auditRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Maybe()
	userDirectory := iamoutputmocks.NewMockUserDirectory(t)
	commandServer := NewIAMCommandServer(
		usecases.commands,
		usecases.queries,
		auditRepo,
		userDirectory,
		testIAMServerCfg,
	)
	queryServer := NewIAMQueryServer(
		usecases.queries,
		auditRepo,
		userDirectory,
		testIAMServerCfg,
	)
	return NewIAMServer(commandServer, queryServer)
}

func authContextForIAMUser(t *testing.T, userID uint) context.Context {
	t.Helper()
	return metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("authorization", "Bearer "+rawAccessTokenForIAMUser(t, userID)),
	)
}

func rawAccessTokenForIAMUser(t *testing.T, userID uint) string {
	t.Helper()
	claims := pdauthn.Claims{
		UserID:    userID,
		Key:       testIAMServerCfg.Authn.JWTKey,
		SessionID: "session-test",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: nowPlusHour().Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(
		[]byte(testIAMServerCfg.Authn.JWTSecret),
	)
	require.NoError(t, err)
	return token
}

func nowPlusHour() time.Time {
	return time.Now().UTC().Add(time.Hour)
}
