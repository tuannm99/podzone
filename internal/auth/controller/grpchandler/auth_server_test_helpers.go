package grpchandler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
	iammocks "github.com/tuannm99/podzone/internal/iam/domain/mocks"
	"google.golang.org/grpc/metadata"
)

var testAuthCfg = config.AuthConfig{
	JWTSecret:      "secret",
	JWTKey:         "app-key",
	AppRedirectURL: "https://app.example.com/auth/google/callback",
}

type iamUsecaseMockConfig struct {
	createTenantFunc       func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error)
	addPlatformRoleFunc    func(ctx context.Context, userID uint, roleName string) error
	addMemberFunc          func(ctx context.Context, tenantID string, userID uint, roleName string) error
	createInviteFunc       func(ctx context.Context, tenantID, email, roleName string, invitedByUserID uint) (*iamdomain.TenantInvite, string, error)
	getInviteFunc          func(ctx context.Context, inviteID string) (*iamdomain.TenantInvite, error)
	listInvitesFunc        func(ctx context.Context, tenantID string) ([]iamdomain.TenantInvite, error)
	revokeInviteFunc       func(ctx context.Context, inviteID string) error
	acceptInviteFunc       func(ctx context.Context, inviteToken string, userID uint, email string) (*iamdomain.Membership, error)
	getMembershipFunc      func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error)
	listPlatformFunc       func(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error)
	listUserTenantsFunc    func(ctx context.Context, userID uint) ([]iamdomain.Membership, error)
	listTenantFunc         func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error)
	removePlatformRoleFunc func(ctx context.Context, userID uint, roleName string) error
	removeMemberFunc       func(ctx context.Context, tenantID string, userID uint) error
	checkPermissionFunc    func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
}

type sessionRepoMockConfig struct {
	getByIDFunc    func(ctx context.Context, id string) (*entity.Session, error)
	listByUserFunc func(ctx context.Context, userID uint) ([]entity.Session, error)
	revokeFunc     func(ctx context.Context, id string, revokedAt time.Time) error
}

type auditRepoMockConfig struct {
	createFunc      func(ctx context.Context, log entity.AuditLog) error
	listByActorFunc func(ctx context.Context, actorUserID uint, limit int) ([]entity.AuditLog, error)
}

func newIAMUsecaseMock(t *testing.T, cfg iamUsecaseMockConfig) *iammocks.MockIAMUsecase {
	t.Helper()
	iamUC := iammocks.NewMockIAMUsecase(t)
	if cfg.createTenantFunc != nil {
		iamUC.EXPECT().CreateTenant(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.createTenantFunc).Maybe()
	}
	if cfg.addPlatformRoleFunc != nil {
		iamUC.EXPECT().AddPlatformRole(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.addPlatformRoleFunc).Maybe()
	}
	if cfg.addMemberFunc != nil {
		iamUC.EXPECT().AddMember(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.addMemberFunc).Maybe()
	}
	if cfg.createInviteFunc != nil {
		iamUC.EXPECT().CreateInvite(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.createInviteFunc).Maybe()
	}
	if cfg.getInviteFunc != nil {
		iamUC.EXPECT().GetInvite(mock.Anything, mock.Anything).RunAndReturn(cfg.getInviteFunc).Maybe()
	}
	if cfg.listInvitesFunc != nil {
		iamUC.EXPECT().ListTenantInvites(mock.Anything, mock.Anything).RunAndReturn(cfg.listInvitesFunc).Maybe()
	}
	if cfg.revokeInviteFunc != nil {
		iamUC.EXPECT().RevokeInvite(mock.Anything, mock.Anything).RunAndReturn(cfg.revokeInviteFunc).Maybe()
	}
	if cfg.acceptInviteFunc != nil {
		iamUC.EXPECT().AcceptInvite(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.acceptInviteFunc).Maybe()
	}
	if cfg.getMembershipFunc != nil {
		iamUC.EXPECT().GetMembership(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.getMembershipFunc).Maybe()
	}
	if cfg.listPlatformFunc != nil {
		iamUC.EXPECT().ListPlatformRoles(mock.Anything, mock.Anything).RunAndReturn(cfg.listPlatformFunc).Maybe()
	}
	if cfg.listUserTenantsFunc != nil {
		iamUC.EXPECT().ListUserTenants(mock.Anything, mock.Anything).RunAndReturn(cfg.listUserTenantsFunc).Maybe()
	}
	if cfg.listTenantFunc != nil {
		iamUC.EXPECT().ListTenantMembers(mock.Anything, mock.Anything).RunAndReturn(cfg.listTenantFunc).Maybe()
	}
	if cfg.removePlatformRoleFunc != nil {
		iamUC.EXPECT().RemovePlatformRole(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.removePlatformRoleFunc).Maybe()
	}
	if cfg.removeMemberFunc != nil {
		iamUC.EXPECT().RemoveMember(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.removeMemberFunc).Maybe()
	}
	if cfg.checkPermissionFunc != nil {
		iamUC.EXPECT().CheckPermission(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.checkPermissionFunc).Maybe()
		iamUC.EXPECT().CheckPlatformPermission(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, userID uint, permission string) (bool, error) {
			return cfg.checkPermissionFunc(ctx, "", userID, permission)
		}).Maybe()
		iamUC.EXPECT().RequirePermission(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, tenantID string, userID uint, permission string) error {
			ok, err := cfg.checkPermissionFunc(ctx, tenantID, userID, permission)
			if err != nil {
				return err
			}
			if !ok {
				return iamdomain.ErrPermissionDenied
			}
			return nil
		}).Maybe()
		iamUC.EXPECT().RequirePlatformPermission(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, userID uint, permission string) error {
			ok, err := cfg.checkPermissionFunc(ctx, "", userID, permission)
			if err != nil {
				return err
			}
			if !ok {
				return iamdomain.ErrPermissionDenied
			}
			return nil
		}).Maybe()
	}
	return iamUC
}

func newSessionRepoMock(t *testing.T, cfg sessionRepoMockConfig) *outputmocks.MockSessionRepository {
	t.Helper()
	repo := outputmocks.NewMockSessionRepository(t)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Maybe()
	repo.EXPECT().UpdateActiveTenant(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	if cfg.getByIDFunc != nil {
		repo.EXPECT().GetByID(mock.Anything, mock.Anything).RunAndReturn(cfg.getByIDFunc).Maybe()
	}
	if cfg.listByUserFunc != nil {
		repo.EXPECT().ListByUser(mock.Anything, mock.Anything).RunAndReturn(cfg.listByUserFunc).Maybe()
	}
	if cfg.revokeFunc != nil {
		repo.EXPECT().Revoke(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.revokeFunc).Maybe()
	}
	return repo
}

func newAuditRepoMock(t *testing.T, cfg auditRepoMockConfig) *outputmocks.MockAuditLogRepository {
	t.Helper()
	repo := outputmocks.NewMockAuditLogRepository(t)
	if cfg.createFunc != nil {
		repo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(cfg.createFunc).Maybe()
	} else {
		repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Maybe()
	}
	if cfg.listByActorFunc != nil {
		repo.EXPECT().ListByActor(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.listByActorFunc).Maybe()
	}
	return repo
}

func newServerWithMock(t *testing.T) (*AuthServer, *inputmocks.MockAuthUsecase) {
	t.Helper()
	authUC := &inputmocks.MockAuthUsecase{}
	srv := NewAuthServer(
		authUC,
		newSessionRepoMock(t, sessionRepoMockConfig{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
			return nil, entity.ErrSessionNotFound
		}}),
		newAuditRepoMock(t, auditRepoMockConfig{}),
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)
	return srv, authUC
}

func newServerWithIAM(t *testing.T, iamUC iamdomain.IAMUsecase) *AuthServer {
	t.Helper()
	return NewIAMServer(iamUC, newAuditRepoMock(t, auditRepoMockConfig{}), &outputmocks.MockUserRepository{}, testAuthCfg)
}

func authContextForUser(t *testing.T, userID uint) context.Context {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testAuthCfg).
		CreateJwtTokenForSession(entity.User{Id: userID}, "", "session-test")
	require.NoError(t, err)
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}
