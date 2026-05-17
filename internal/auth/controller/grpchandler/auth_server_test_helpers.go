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
	createTenantFunc                   func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error)
	createPolicyFunc                   func(ctx context.Context, input iamdomain.CreatePolicyInput) (*iamdomain.Policy, []iamdomain.PolicyStatement, error)
	getPolicyFunc                      func(ctx context.Context, name string) (*iamdomain.Policy, []iamdomain.PolicyStatement, error)
	listPoliciesFunc                   func(ctx context.Context, scope string) ([]iamdomain.Policy, error)
	listPolicyAttachmentsFunc          func(ctx context.Context, name string) ([]iamdomain.PolicyAttachment, error)
	deletePolicyFunc                   func(ctx context.Context, name string) error
	putRoleTrustPolicyFunc             func(ctx context.Context, input iamdomain.PutRoleTrustPolicyInput) error
	getRoleTrustPolicyFunc             func(ctx context.Context, roleName string) ([]iamdomain.RoleTrustStatement, error)
	deleteRoleTrustPolicyFunc          func(ctx context.Context, roleName string) error
	assumeRoleFunc                     func(ctx context.Context, input iamdomain.AssumeRoleInput) (*iamdomain.AssumedRole, error)
	createGroupFunc                    func(ctx context.Context, input iamdomain.CreateGroupInput) (*iamdomain.Group, error)
	listGroupsFunc                     func(ctx context.Context, scope string, tenantID string) ([]iamdomain.Group, error)
	deleteGroupFunc                    func(ctx context.Context, groupID uint64) error
	addGroupMemberFunc                 func(ctx context.Context, groupID uint64, userID uint) error
	removeGroupMemberFunc              func(ctx context.Context, groupID uint64, userID uint) error
	listGroupMembersFunc               func(ctx context.Context, groupID uint64) ([]uint, error)
	attachGroupPolicyFunc              func(ctx context.Context, groupID uint64, policyName string) error
	detachGroupPolicyFunc              func(ctx context.Context, groupID uint64, policyName string) error
	listGroupPoliciesFunc              func(ctx context.Context, groupID uint64) ([]iamdomain.Policy, error)
	putGroupInlinePolicyFunc           func(ctx context.Context, input iamdomain.PutGroupInlinePolicyInput) error
	getGroupInlinePolicyFunc           func(ctx context.Context, groupID uint64, name string) (*iamdomain.GroupInlinePolicy, error)
	listGroupInlinePoliciesFunc        func(ctx context.Context, groupID uint64) ([]iamdomain.GroupInlinePolicy, error)
	deleteGroupInlinePolicyFunc        func(ctx context.Context, groupID uint64, name string) error
	putPlatformUserInlinePolicyFunc    func(ctx context.Context, input iamdomain.PutPlatformUserInlinePolicyInput) error
	getPlatformUserInlinePolicyFunc    func(ctx context.Context, userID uint, name string) (*iamdomain.UserInlinePolicy, error)
	listPlatformUserInlinePoliciesFunc func(ctx context.Context, userID uint) ([]iamdomain.UserInlinePolicy, error)
	deletePlatformUserInlinePolicyFunc func(ctx context.Context, userID uint, name string) error
	addPlatformRoleFunc                func(ctx context.Context, userID uint, roleName string) error
	attachPlatformPolicyFunc           func(ctx context.Context, userID uint, policyName string) error
	detachPlatformPolicyFunc           func(ctx context.Context, userID uint, policyName string) error
	listPlatformPoliciesFunc           func(ctx context.Context, userID uint) ([]iamdomain.Policy, error)
	addMemberFunc                      func(ctx context.Context, tenantID string, userID uint, roleName string) error
	putTenantUserInlinePolicyFunc      func(ctx context.Context, input iamdomain.PutTenantUserInlinePolicyInput) error
	getTenantUserInlinePolicyFunc      func(ctx context.Context, tenantID string, userID uint, name string) (*iamdomain.UserInlinePolicy, error)
	listTenantUserInlinePoliciesFunc   func(ctx context.Context, tenantID string, userID uint) ([]iamdomain.UserInlinePolicy, error)
	deleteTenantUserInlinePolicyFunc   func(ctx context.Context, tenantID string, userID uint, name string) error
	attachTenantPolicyFunc             func(ctx context.Context, tenantID string, userID uint, policyName string) error
	detachTenantPolicyFunc             func(ctx context.Context, tenantID string, userID uint, policyName string) error
	listTenantPoliciesFunc             func(ctx context.Context, tenantID string, userID uint) ([]iamdomain.Policy, error)
	createInviteFunc                   func(ctx context.Context, tenantID, email, roleName string, invitedByUserID uint) (*iamdomain.TenantInvite, string, error)
	getInviteFunc                      func(ctx context.Context, inviteID string) (*iamdomain.TenantInvite, error)
	listInvitesFunc                    func(ctx context.Context, tenantID string) ([]iamdomain.TenantInvite, error)
	revokeInviteFunc                   func(ctx context.Context, inviteID string) error
	acceptInviteFunc                   func(ctx context.Context, inviteToken string, userID uint, email string) (*iamdomain.Membership, error)
	getMembershipFunc                  func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error)
	listPlatformFunc                   func(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error)
	listUserTenantsFunc                func(ctx context.Context, userID uint) ([]iamdomain.Membership, error)
	listTenantFunc                     func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error)
	removePlatformRoleFunc             func(ctx context.Context, userID uint, roleName string) error
	removeMemberFunc                   func(ctx context.Context, tenantID string, userID uint) error
	checkPermissionFunc                func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
}

type sessionRepoMockConfig struct {
	getByIDFunc           func(ctx context.Context, id string) (*entity.Session, error)
	listByUserFunc        func(ctx context.Context, userID uint) ([]entity.Session, error)
	revokeFunc            func(ctx context.Context, id string, revokedAt time.Time) error
	updateAssumedRoleFunc func(ctx context.Context, session entity.Session, updatedAt time.Time) error
}

type auditRepoMockConfig struct {
	createFunc      func(ctx context.Context, log entity.AuditLog) error
	listByActorFunc func(ctx context.Context, actorUserID uint, limit int) ([]entity.AuditLog, error)
}

func newIAMUsecaseMock(t *testing.T, cfg iamUsecaseMockConfig) *iammocks.MockIAMUsecase {
	t.Helper()
	iamUC := iammocks.NewMockIAMUsecase(t)
	if cfg.createTenantFunc != nil {
		iamUC.EXPECT().
			CreateTenant(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.createTenantFunc).
			Maybe()
	}
	if cfg.createPolicyFunc != nil {
		iamUC.EXPECT().CreatePolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.createPolicyFunc).Maybe()
	}
	if cfg.getPolicyFunc != nil {
		iamUC.EXPECT().GetPolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.getPolicyFunc).Maybe()
	}
	if cfg.listPoliciesFunc != nil {
		iamUC.EXPECT().ListPolicies(mock.Anything, mock.Anything).RunAndReturn(cfg.listPoliciesFunc).Maybe()
	}
	if cfg.listPolicyAttachmentsFunc != nil {
		iamUC.EXPECT().
			ListPolicyAttachments(mock.Anything, mock.Anything).
			RunAndReturn(cfg.listPolicyAttachmentsFunc).
			Maybe()
	}
	if cfg.deletePolicyFunc != nil {
		iamUC.EXPECT().DeletePolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.deletePolicyFunc).Maybe()
	}
	if cfg.putRoleTrustPolicyFunc != nil {
		iamUC.EXPECT().PutRoleTrustPolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.putRoleTrustPolicyFunc).Maybe()
	}
	if cfg.getRoleTrustPolicyFunc != nil {
		iamUC.EXPECT().GetRoleTrustPolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.getRoleTrustPolicyFunc).Maybe()
	}
	if cfg.deleteRoleTrustPolicyFunc != nil {
		iamUC.EXPECT().DeleteRoleTrustPolicy(mock.Anything, mock.Anything).RunAndReturn(cfg.deleteRoleTrustPolicyFunc).Maybe()
	}
	if cfg.assumeRoleFunc != nil {
		iamUC.EXPECT().AssumeRole(mock.Anything, mock.Anything).RunAndReturn(cfg.assumeRoleFunc).Maybe()
	}
	if cfg.createGroupFunc != nil {
		iamUC.EXPECT().CreateGroup(mock.Anything, mock.Anything).RunAndReturn(cfg.createGroupFunc).Maybe()
	}
	if cfg.listGroupsFunc != nil {
		iamUC.EXPECT().ListGroups(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.listGroupsFunc).Maybe()
	}
	if cfg.deleteGroupFunc != nil {
		iamUC.EXPECT().DeleteGroup(mock.Anything, mock.Anything).RunAndReturn(cfg.deleteGroupFunc).Maybe()
	}
	if cfg.addGroupMemberFunc != nil {
		iamUC.EXPECT().
			AddGroupMember(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.addGroupMemberFunc).
			Maybe()
	}
	if cfg.removeGroupMemberFunc != nil {
		iamUC.EXPECT().
			RemoveGroupMember(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.removeGroupMemberFunc).
			Maybe()
	}
	if cfg.listGroupMembersFunc != nil {
		iamUC.EXPECT().ListGroupMembers(mock.Anything, mock.Anything).RunAndReturn(cfg.listGroupMembersFunc).Maybe()
	}
	if cfg.attachGroupPolicyFunc != nil {
		iamUC.EXPECT().
			AttachGroupPolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.attachGroupPolicyFunc).
			Maybe()
	}
	if cfg.detachGroupPolicyFunc != nil {
		iamUC.EXPECT().
			DetachGroupPolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.detachGroupPolicyFunc).
			Maybe()
	}
	if cfg.listGroupPoliciesFunc != nil {
		iamUC.EXPECT().ListGroupPolicies(mock.Anything, mock.Anything).RunAndReturn(cfg.listGroupPoliciesFunc).Maybe()
	}
	if cfg.putGroupInlinePolicyFunc != nil {
		iamUC.EXPECT().
			PutGroupInlinePolicy(mock.Anything, mock.Anything).
			RunAndReturn(cfg.putGroupInlinePolicyFunc).
			Maybe()
	}
	if cfg.getGroupInlinePolicyFunc != nil {
		iamUC.EXPECT().
			GetGroupInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.getGroupInlinePolicyFunc).
			Maybe()
	}
	if cfg.listGroupInlinePoliciesFunc != nil {
		iamUC.EXPECT().
			ListGroupInlinePolicies(mock.Anything, mock.Anything).
			RunAndReturn(cfg.listGroupInlinePoliciesFunc).
			Maybe()
	}
	if cfg.deleteGroupInlinePolicyFunc != nil {
		iamUC.EXPECT().
			DeleteGroupInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.deleteGroupInlinePolicyFunc).
			Maybe()
	}
	if cfg.addPlatformRoleFunc != nil {
		iamUC.EXPECT().
			AddPlatformRole(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.addPlatformRoleFunc).
			Maybe()
	}
	if cfg.attachPlatformPolicyFunc != nil {
		iamUC.EXPECT().
			AttachPlatformUserPolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.attachPlatformPolicyFunc).
			Maybe()
	}
	if cfg.detachPlatformPolicyFunc != nil {
		iamUC.EXPECT().
			DetachPlatformUserPolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.detachPlatformPolicyFunc).
			Maybe()
	}
	if cfg.listPlatformPoliciesFunc != nil {
		iamUC.EXPECT().
			ListPlatformUserPolicies(mock.Anything, mock.Anything).
			RunAndReturn(cfg.listPlatformPoliciesFunc).
			Maybe()
	}
	if cfg.putPlatformUserInlinePolicyFunc != nil {
		iamUC.EXPECT().
			PutPlatformUserInlinePolicy(mock.Anything, mock.Anything).
			RunAndReturn(cfg.putPlatformUserInlinePolicyFunc).
			Maybe()
	}
	if cfg.getPlatformUserInlinePolicyFunc != nil {
		iamUC.EXPECT().
			GetPlatformUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.getPlatformUserInlinePolicyFunc).
			Maybe()
	}
	if cfg.listPlatformUserInlinePoliciesFunc != nil {
		iamUC.EXPECT().
			ListPlatformUserInlinePolicies(mock.Anything, mock.Anything).
			RunAndReturn(cfg.listPlatformUserInlinePoliciesFunc).
			Maybe()
	}
	if cfg.deletePlatformUserInlinePolicyFunc != nil {
		iamUC.EXPECT().
			DeletePlatformUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.deletePlatformUserInlinePolicyFunc).
			Maybe()
	}
	if cfg.addMemberFunc != nil {
		iamUC.EXPECT().
			AddMember(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.addMemberFunc).
			Maybe()
	}
	if cfg.putTenantUserInlinePolicyFunc != nil {
		iamUC.EXPECT().
			PutTenantUserInlinePolicy(mock.Anything, mock.Anything).
			RunAndReturn(cfg.putTenantUserInlinePolicyFunc).
			Maybe()
	}
	if cfg.getTenantUserInlinePolicyFunc != nil {
		iamUC.EXPECT().
			GetTenantUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.getTenantUserInlinePolicyFunc).
			Maybe()
	}
	if cfg.listTenantUserInlinePoliciesFunc != nil {
		iamUC.EXPECT().
			ListTenantUserInlinePolicies(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.listTenantUserInlinePoliciesFunc).
			Maybe()
	}
	if cfg.deleteTenantUserInlinePolicyFunc != nil {
		iamUC.EXPECT().
			DeleteTenantUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.deleteTenantUserInlinePolicyFunc).
			Maybe()
	}
	if cfg.attachTenantPolicyFunc != nil {
		iamUC.EXPECT().
			AttachTenantUserPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.attachTenantPolicyFunc).
			Maybe()
	}
	if cfg.detachTenantPolicyFunc != nil {
		iamUC.EXPECT().
			DetachTenantUserPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.detachTenantPolicyFunc).
			Maybe()
	}
	if cfg.listTenantPoliciesFunc != nil {
		iamUC.EXPECT().
			ListTenantUserPolicies(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.listTenantPoliciesFunc).
			Maybe()
	}
	if cfg.createInviteFunc != nil {
		iamUC.EXPECT().
			CreateInvite(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.createInviteFunc).
			Maybe()
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
		iamUC.EXPECT().
			AcceptInvite(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.acceptInviteFunc).
			Maybe()
	}
	if cfg.getMembershipFunc != nil {
		iamUC.EXPECT().
			GetMembership(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.getMembershipFunc).
			Maybe()
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
		iamUC.EXPECT().
			RemovePlatformRole(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.removePlatformRoleFunc).
			Maybe()
	}
	if cfg.removeMemberFunc != nil {
		iamUC.EXPECT().
			RemoveMember(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.removeMemberFunc).
			Maybe()
	}
	if cfg.checkPermissionFunc != nil {
		iamUC.EXPECT().
			CheckPermission(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(cfg.checkPermissionFunc).
			Maybe()
		iamUC.EXPECT().
			CheckPlatformPermission(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, userID uint, permission string) (bool, error) {
				return cfg.checkPermissionFunc(ctx, "", userID, permission)
			}).
			Maybe()
		iamUC.EXPECT().
			RequirePermission(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, tenantID string, userID uint, permission string) error {
				ok, err := cfg.checkPermissionFunc(ctx, tenantID, userID, permission)
				if err != nil {
					return err
				}
				if !ok {
					return iamdomain.ErrPermissionDenied
				}
				return nil
			}).
			Maybe()
		iamUC.EXPECT().
			RequirePlatformPermission(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, userID uint, permission string) error {
				ok, err := cfg.checkPermissionFunc(ctx, "", userID, permission)
				if err != nil {
					return err
				}
				if !ok {
					return iamdomain.ErrPermissionDenied
				}
				return nil
			}).
			Maybe()
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
	if cfg.updateAssumedRoleFunc != nil {
		repo.EXPECT().UpdateAssumedRole(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(cfg.updateAssumedRoleFunc).Maybe()
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
		newSessionRepoMock(
			t,
			sessionRepoMockConfig{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
				return nil, entity.ErrSessionNotFound
			}},
		),
		newAuditRepoMock(t, auditRepoMockConfig{}),
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)
	return srv, authUC
}

func newServerWithIAM(t *testing.T, iamUC iamdomain.IAMUsecase) *AuthServer {
	t.Helper()
	return NewIAMServer(
		iamUC,
		newAuditRepoMock(t, auditRepoMockConfig{}),
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)
}

func authContextForUser(t *testing.T, userID uint) context.Context {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testAuthCfg).
		CreateJwtTokenForSession(entity.User{Id: userID}, "", "session-test")
	require.NoError(t, err)
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func accessTokenForSession(
	t *testing.T,
	user entity.User,
	activeTenantID, sessionID string,
	sessionPolicy ...entity.SessionPolicyStatement,
) string {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testAuthCfg).
		CreateJwtTokenForScopedSession(user, activeTenantID, sessionID, sessionPolicy)
	require.NoError(t, err)
	return token
}
