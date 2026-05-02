package grpchandler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var testAuthCfg = config.AuthConfig{
	JWTSecret:      "secret",
	JWTKey:         "app-key",
	AppRedirectURL: "https://app.example.com/auth/google/callback",
}

type iamUsecaseFake struct {
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

func (f *iamUsecaseFake) CreateTenant(
	ctx context.Context,
	ownerUserID uint,
	cmd iamdomain.CreateTenantCmd,
) (*iamdomain.Tenant, error) {
	return f.createTenantFunc(ctx, ownerUserID, cmd)
}

func (f *iamUsecaseFake) AddPlatformRole(ctx context.Context, userID uint, roleName string) error {
	return f.addPlatformRoleFunc(ctx, userID, roleName)
}

func (f *iamUsecaseFake) AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error {
	return f.addMemberFunc(ctx, tenantID, userID, roleName)
}

func (f *iamUsecaseFake) CreateInvite(
	ctx context.Context,
	tenantID, email, roleName string,
	invitedByUserID uint,
) (*iamdomain.TenantInvite, string, error) {
	return f.createInviteFunc(ctx, tenantID, email, roleName, invitedByUserID)
}

func (f *iamUsecaseFake) GetInvite(ctx context.Context, inviteID string) (*iamdomain.TenantInvite, error) {
	return f.getInviteFunc(ctx, inviteID)
}

func (f *iamUsecaseFake) ListTenantInvites(ctx context.Context, tenantID string) ([]iamdomain.TenantInvite, error) {
	return f.listInvitesFunc(ctx, tenantID)
}

func (f *iamUsecaseFake) RevokeInvite(ctx context.Context, inviteID string) error {
	return f.revokeInviteFunc(ctx, inviteID)
}

func (f *iamUsecaseFake) AcceptInvite(
	ctx context.Context,
	inviteToken string,
	userID uint,
	email string,
) (*iamdomain.Membership, error) {
	return f.acceptInviteFunc(ctx, inviteToken, userID, email)
}

func (f *iamUsecaseFake) GetMembership(
	ctx context.Context,
	tenantID string,
	userID uint,
) (*iamdomain.Membership, error) {
	return f.getMembershipFunc(ctx, tenantID, userID)
}

func (f *iamUsecaseFake) CheckPermission(
	ctx context.Context,
	tenantID string,
	userID uint,
	permission string,
) (bool, error) {
	return f.checkPermissionFunc(ctx, tenantID, userID, permission)
}

func (f *iamUsecaseFake) CheckPlatformPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	return f.checkPermissionFunc(ctx, "", userID, permission)
}

func (f *iamUsecaseFake) ListUserTenants(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
	return f.listUserTenantsFunc(ctx, userID)
}

func (f *iamUsecaseFake) ListPlatformRoles(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error) {
	return f.listPlatformFunc(ctx, userID)
}

func (f *iamUsecaseFake) ListTenantMembers(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
	return f.listTenantFunc(ctx, tenantID)
}

func (f *iamUsecaseFake) RemoveMember(ctx context.Context, tenantID string, userID uint) error {
	return f.removeMemberFunc(ctx, tenantID, userID)
}

func (f *iamUsecaseFake) RemovePlatformRole(ctx context.Context, userID uint, roleName string) error {
	return f.removePlatformRoleFunc(ctx, userID, roleName)
}

func (f *iamUsecaseFake) RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error {
	ok, err := f.checkPermissionFunc(ctx, tenantID, userID, permission)
	if err != nil {
		return err
	}
	if !ok {
		return iamdomain.ErrPermissionDenied
	}
	return nil
}

func (f *iamUsecaseFake) RequirePlatformPermission(ctx context.Context, userID uint, permission string) error {
	ok, err := f.checkPermissionFunc(ctx, "", userID, permission)
	if err != nil {
		return err
	}
	if !ok {
		return iamdomain.ErrPermissionDenied
	}
	return nil
}

type sessionRepoFake struct {
	getByIDFunc    func(ctx context.Context, id string) (*entity.Session, error)
	listByUserFunc func(ctx context.Context, userID uint) ([]entity.Session, error)
	revokeFunc     func(ctx context.Context, id string, revokedAt time.Time) error
}

func (f *sessionRepoFake) Create(ctx context.Context, session entity.Session) error { return nil }
func (f *sessionRepoFake) ListByUser(ctx context.Context, userID uint) ([]entity.Session, error) {
	if f.listByUserFunc != nil {
		return f.listByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (f *sessionRepoFake) UpdateActiveTenant(ctx context.Context, id, tenantID string, updatedAt time.Time) error {
	return nil
}

func (f *sessionRepoFake) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	if f.revokeFunc != nil {
		return f.revokeFunc(ctx, id, revokedAt)
	}
	return nil
}

func (f *sessionRepoFake) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	return f.getByIDFunc(ctx, id)
}

type auditRepoFake struct {
	createFunc      func(ctx context.Context, log entity.AuditLog) error
	listByActorFunc func(ctx context.Context, actorUserID uint, limit int) ([]entity.AuditLog, error)
}

func (f *auditRepoFake) Create(ctx context.Context, log entity.AuditLog) error {
	if f.createFunc != nil {
		return f.createFunc(ctx, log)
	}
	return nil
}

func (f *auditRepoFake) ListByActor(ctx context.Context, actorUserID uint, limit int) ([]entity.AuditLog, error) {
	if f.listByActorFunc != nil {
		return f.listByActorFunc(ctx, actorUserID, limit)
	}
	return nil, nil
}

func newServerWithMock() (*AuthServer, *inputmocks.MockAuthUsecase) {
	authUC := &inputmocks.MockAuthUsecase{}
	srv := NewAuthServer(
		authUC,
		&sessionRepoFake{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
			return nil, entity.ErrSessionNotFound
		}},
		&auditRepoFake{},
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)
	return srv, authUC
}

func newServerWithIAM(iamUC iamdomain.IAMUsecase) *AuthServer {
	return NewIAMServer(iamUC, &auditRepoFake{}, &outputmocks.MockUserRepository{}, testAuthCfg)
}

func authContextForUser(t *testing.T, userID uint) context.Context {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testAuthCfg).
		CreateJwtTokenForSession(entity.User{Id: userID}, "", "session-test")
	require.NoError(t, err)
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func TestGoogleLogin_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("GenerateOAuthURL", mock.Anything).
		Return("https://accounts.google.com/auth?state=xyz", nil)

	res, err := srv.GoogleLogin(ctx, &pbauthv1.GoogleLoginRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "https://accounts.google.com/auth?state=xyz", res.RedirectUrl)

	uc.AssertExpectations(t)
}

func TestGoogleLogin_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("GenerateOAuthURL", mock.Anything).
		Return("", assert.AnError)

	res, err := srv.GoogleLogin(ctx, &pbauthv1.GoogleLoginRequest{})
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestGoogleCallback_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	cb := &inputport.GoogleCallbackResult{
		ExchangeCode: "oauth-code-1",
		RedirectUrl:  "https://app.example.com/auth/google/callback?exchange_code=oauth-code-1",
		UserInfo: inputport.GoogleUserInfo{
			Id:    "gid-1",
			Email: "neo@mx.io",
			Name:  "Neo",
		},
	}
	uc.On("HandleOAuthCallback", mock.Anything, "CODE", "STATE").
		Return(cb, nil)

	res, err := srv.GoogleCallback(ctx, &pbauthv1.GoogleCallbackRequest{
		Code:  "CODE",
		State: "STATE",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "oauth-code-1", res.ExchangeCode)
	assert.Equal(t, "https://app.example.com/auth/google/callback?exchange_code=oauth-code-1", res.RedirectUrl)
	assert.Equal(t, "neo@mx.io", res.UserInfo.Email)
	assert.Equal(t, "Neo", res.UserInfo.Name)

	uc.AssertExpectations(t)
}

func TestExchangeGoogleLogin_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("ExchangeOAuthLogin", mock.Anything, "oauth-code-1").
		Return(&inputport.AuthResult{
			JwtToken:     "jwt-login",
			RefreshToken: "refresh-login",
			UserInfo: entity.User{
				Id:       7,
				Email:    "neo@mx.io",
				Username: "neo",
			},
		}, nil)

	res, err := srv.ExchangeGoogleLogin(ctx, &pbauthv1.ExchangeGoogleLoginRequest{
		ExchangeCode: "oauth-code-1",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-login", res.JwtToken)
	assert.Equal(t, "refresh-login", res.RefreshToken)
	assert.Equal(t, "neo", res.UserInfo.Username)
}

func TestGoogleCallback_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("HandleOAuthCallback", mock.Anything, "BAD", "STATE").
		Return((*inputport.GoogleCallbackResult)(nil), assert.AnError)

	res, err := srv.GoogleCallback(ctx, &pbauthv1.GoogleCallbackRequest{
		Code:  "BAD",
		State: "STATE",
	})
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestLogout_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("Logout", mock.Anything).
		Return("/", nil)
	uc.ExpectedCalls = nil
	uc.On("Logout", mock.Anything, "access-token").
		Return("/", nil)

	res, err := srv.Logout(ctx, &pbauthv1.LogoutRequest{Token: "access-token"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Success)
	assert.Equal(t, "/", res.RedirectUrl)

	uc.AssertExpectations(t)
}

func TestListSessions_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)
	srv := NewAuthServer(
		&inputmocks.MockAuthUsecase{},
		&sessionRepoFake{
			getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) { return nil, entity.ErrSessionNotFound },
			listByUserFunc: func(ctx context.Context, userID uint) ([]entity.Session, error) {
				require.Equal(t, uint(7), userID)
				return []entity.Session{{
					ID:             "session-1",
					UserID:         7,
					ActiveTenantID: "tenant-1",
					Status:         entity.SessionStatusActive,
					CreatedAt:      now,
					UpdatedAt:      now,
					ExpiresAt:      now.Add(time.Hour),
				}}, nil
			},
		},
		&auditRepoFake{},
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)

	res, err := srv.ListSessions(authContextForUser(t, 7), &pbauthv1.ListSessionsRequest{})
	require.NoError(t, err)
	require.Len(t, res.Sessions, 1)
	assert.Equal(t, "session-1", res.Sessions[0].Id)
}

func TestListAuditLogs_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 16, 0, 0, 0, time.UTC)
	srv := NewAuthServer(
		&inputmocks.MockAuthUsecase{},
		&sessionRepoFake{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
			return nil, entity.ErrSessionNotFound
		}},
		&auditRepoFake{
			listByActorFunc: func(ctx context.Context, actorUserID uint, limit int) ([]entity.AuditLog, error) {
				require.Equal(t, uint(7), actorUserID)
				require.Equal(t, 25, limit)
				return []entity.AuditLog{{
					ID:           "audit-1",
					ActorUserID:  7,
					Action:       "tenant.created",
					ResourceType: "tenant",
					ResourceID:   "tenant-1",
					TenantID:     "tenant-1",
					Status:       entity.AuditStatusSuccess,
					PayloadJSON:  `{"slug":"seller-one"}`,
					CreatedAt:    now,
				}}, nil
			},
		},
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)

	res, err := srv.ListAuditLogs(authContextForUser(t, 7), &pbauthv1.ListAuditLogsRequest{PageSize: 25})
	require.NoError(t, err)
	require.Len(t, res.Logs, 1)
	assert.Equal(t, "audit-1", res.Logs[0].Id)
	assert.Equal(t, "tenant.created", res.Logs[0].Action)
	assert.Equal(t, `{"slug":"seller-one"}`, res.Logs[0].PayloadJson)
}

func TestLogin_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	user := entity.User{
		Id:       1,
		Email:    "jdoe@example.com",
		Username: "jdoe",
	}
	uc.On("Login", mock.Anything, "jdoe", "pass").
		Return(&inputport.AuthResult{
			JwtToken: "jwt-login",
			UserInfo: user,
		}, nil)

	res, err := srv.Login(ctx, &pbauthv1.LoginRequest{
		Username: "jdoe",
		Password: "pass",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-login", res.JwtToken)
	assert.Equal(t, "jdoe@example.com", res.UserInfo.Email)
	assert.Equal(t, "jdoe", res.UserInfo.Username)

	uc.AssertExpectations(t)
}

func TestLogin_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("Login", mock.Anything, "jdoe", "bad").
		Return((*inputport.AuthResult)(nil), assert.AnError)

	res, err := srv.Login(ctx, &pbauthv1.LoginRequest{
		Username: "jdoe",
		Password: "bad",
	})
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestRegister_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	inReq := &pbauthv1.RegisterRequest{
		Username: "neo",
		Password: "TheOne!",
		Email:    "neo@mx.io",
	}
	out := &inputport.AuthResult{
		JwtToken: "jwt-reg",
		UserInfo: entity.User{
			Id:       9,
			Email:    "neo@mx.io",
			Username: "neo",
		},
	}

	// Dùng MatchedBy để không phụ thuộc chi tiết mapping (toolkit.MapStruct)
	uc.On("Register", mock.Anything, mock.MatchedBy(func(r inputport.RegisterCmd) bool {
		return r.Username == "neo" && r.Password == "TheOne!" && r.Email == "neo@mx.io"
	})).Return(out, nil)

	res, err := srv.Register(ctx, inReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-reg", res.JwtToken)
	assert.Equal(t, "neo@mx.io", res.UserInfo.Email)
	assert.Equal(t, "neo", res.UserInfo.Username)

	uc.AssertExpectations(t)
}

func TestRegister_Err(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	inReq := &pbauthv1.RegisterRequest{
		Username: "neo",
		Password: "x",
		Email:    "neo@mx.io",
	}
	uc.On("Register", mock.Anything, mock.AnythingOfType("inputport.RegisterCmd")).
		Return((*inputport.AuthResult)(nil), assert.AnError)

	res, err := srv.Register(ctx, inReq)
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestCreateTenant_OK(t *testing.T) {
	now := time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
	var createdAudit entity.AuditLog
	srv := NewIAMServer(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			require.Equal(t, uint(7), ownerUserID)
			require.Equal(t, "seller-one", cmd.Slug)
			require.Equal(t, "Seller One", cmd.Name)
			return &iamdomain.Tenant{
				ID:        "tenant-1",
				Slug:      cmd.Slug,
				Name:      cmd.Name,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			require.Equal(t, "tenant-1", tenantID)
			require.Equal(t, uint(7), userID)
			return &iamdomain.Membership{
				TenantID:  tenantID,
				UserID:    userID,
				RoleID:    1,
				RoleName:  iamdomain.RoleTenantOwner,
				Status:    iamdomain.MembershipStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			if tenantID == "" && permission == "tenant:create" && userID == 7 {
				return true, nil
			}
			return false, nil
		},
	}, &auditRepoFake{createFunc: func(ctx context.Context, log entity.AuditLog) error {
		createdAudit = log
		return nil
	}}, &outputmocks.MockUserRepository{}, testAuthCfg)

	res, err := srv.CreateTenant(authContextForUser(t, 7), &pbauthv1.CreateTenantRequest{
		OwnerUserId: 7,
		Slug:        "seller-one",
		Name:        "Seller One",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-1", res.Tenant.Id)
	assert.Equal(t, iamdomain.RoleTenantOwner, res.OwnerMembership.RoleName)
	assert.Equal(t, "2026-04-30T12:00:00Z", res.Tenant.CreatedAt)
	assert.Equal(t, uint(7), createdAudit.ActorUserID)
	assert.Equal(t, "tenant.created", createdAudit.Action)
	assert.Equal(t, "tenant-1", createdAudit.ResourceID)
}

func TestCreateTenant_PermissionDenied(t *testing.T) {
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, errors.New("unexpected CreateTenant call")
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, nil
		},
	})

	res, err := srv.CreateTenant(authContextForUser(t, 7), &pbauthv1.CreateTenantRequest{
		OwnerUserId: 7,
		Slug:        "seller-one",
		Name:        "Seller One",
	})
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestAddTenantMember_NotFound(t *testing.T) {
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return iamdomain.ErrTenantNotFound
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			require.Equal(t, "missing", tenantID)
			require.Equal(t, uint(9), userID)
			require.Equal(t, "tenant:manage_members", permission)
			return true, nil
		},
	})

	res, err := srv.AddTenantMember(authContextForUser(t, 9), &pbauthv1.AddTenantMemberRequest{
		TenantId: "missing",
		UserId:   9,
		RoleName: iamdomain.RoleTenantAdmin,
	})
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestAddTenantMemberByIdentity_CreatesEmailUser(t *testing.T) {
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByUsernameOrEmail", "owner@shop.com").Return((*entity.User)(nil), entity.ErrUserNotFound)
	userRepo.On("CreateByEmailIfNotExisted", "owner@shop.com").Return(&entity.User{
		Id:    77,
		Email: "owner@shop.com",
	}, nil)

	called := false
	srv := NewIAMServer(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			called = true
			require.Equal(t, "tenant-9", tenantID)
			require.Equal(t, uint(77), userID)
			require.Equal(t, iamdomain.RoleTenantAdmin, roleName)
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) { return nil, nil },
		listTenantFunc:      func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) { return nil, nil },
		listPlatformFunc:    func(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error) { return nil, nil },
		removeMemberFunc:    func(ctx context.Context, tenantID string, userID uint) error { return nil },
		removePlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			require.Equal(t, "tenant-9", tenantID)
			require.Equal(t, uint(7), userID)
			require.Equal(t, "tenant:manage_members", permission)
			return true, nil
		},
	}, &auditRepoFake{}, userRepo, testAuthCfg)

	res, err := srv.AddTenantMemberByIdentity(authContextForUser(t, 7), &pbauthv1.AddTenantMemberByIdentityRequest{
		TenantId: "tenant-9",
		Identity: "owner@shop.com",
		RoleName: iamdomain.RoleTenantAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, called)
	assert.Equal(t, uint64(77), res.UserId)
	assert.True(t, res.CreatedUser)
	userRepo.AssertExpectations(t)
}

func TestCreateTenantInvite_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	srv := newServerWithIAM(&iamUsecaseFake{
		createInviteFunc: func(ctx context.Context, tenantID, email, roleName string, invitedByUserID uint) (*iamdomain.TenantInvite, string, error) {
			require.Equal(t, "tenant-9", tenantID)
			require.Equal(t, "owner@shop.com", email)
			require.Equal(t, iamdomain.RoleTenantAdmin, roleName)
			require.Equal(t, uint(7), invitedByUserID)
			return &iamdomain.TenantInvite{
				ID:              "invite-1",
				TenantID:        tenantID,
				Email:           email,
				RoleID:          2,
				RoleName:        roleName,
				Status:          iamdomain.InviteStatusPending,
				InvitedByUserID: invitedByUserID,
				CreatedAt:       now,
				UpdatedAt:       now,
				ExpiresAt:       now.Add(7 * 24 * time.Hour),
			}, "raw-token-1", nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			require.Equal(t, "tenant-9", tenantID)
			require.Equal(t, uint(7), userID)
			require.Equal(t, "tenant:manage_members", permission)
			return true, nil
		},
	})

	res, err := srv.CreateTenantInvite(authContextForUser(t, 7), &pbauthv1.CreateTenantInviteRequest{
		TenantId: "tenant-9",
		Email:    "owner@shop.com",
		RoleName: iamdomain.RoleTenantAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "invite-1", res.Invite.Id)
	assert.Equal(t, "raw-token-1", res.InviteToken)
	assert.Equal(t, "https://app.example.com/auth/invite/accept?token=raw-token-1", res.AcceptUrl)
}

func TestAcceptTenantInvite_OK(t *testing.T) {
	userRepo := &outputmocks.MockUserRepository{}
	userRepo.On("GetByID", "7").Return(&entity.User{
		Id:    7,
		Email: "owner@shop.com",
	}, nil)

	srv := NewIAMServer(&iamUsecaseFake{
		acceptInviteFunc: func(ctx context.Context, inviteToken string, userID uint, email string) (*iamdomain.Membership, error) {
			require.Equal(t, "raw-token-1", inviteToken)
			require.Equal(t, uint(7), userID)
			require.Equal(t, "owner@shop.com", email)
			return &iamdomain.Membership{
				TenantID:  "tenant-9",
				UserID:    userID,
				RoleID:    2,
				RoleName:  iamdomain.RoleTenantAdmin,
				Status:    iamdomain.MembershipStatusActive,
				CreatedAt: time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC),
			}, nil
		},
	}, &auditRepoFake{}, userRepo, testAuthCfg)

	res, err := srv.AcceptTenantInvite(authContextForUser(t, 7), &pbauthv1.AcceptTenantInviteRequest{
		InviteToken: "raw-token-1",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-9", res.Membership.TenantId)
	assert.Equal(t, iamdomain.RoleTenantAdmin, res.Membership.RoleName)
	userRepo.AssertExpectations(t)
}

func TestGetTenantMembership_OK(t *testing.T) {
	now := time.Date(2026, 4, 30, 13, 0, 0, 0, time.UTC)
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return &iamdomain.Membership{
				TenantID:  tenantID,
				UserID:    userID,
				RoleID:    2,
				RoleName:  iamdomain.RoleTenantAdmin,
				Status:    iamdomain.MembershipStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, nil
		},
	})

	res, err := srv.GetTenantMembership(context.Background(), &pbauthv1.GetTenantMembershipRequest{
		TenantId: "tenant-2",
		UserId:   11,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-2", res.Membership.TenantId)
	assert.Equal(t, uint64(11), res.Membership.UserId)
	assert.Equal(t, iamdomain.RoleTenantAdmin, res.Membership.RoleName)
}

func TestAddTenantMember_RequiresTenantManagePermission(t *testing.T) {
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addPlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error { return nil },
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return errors.New("unexpected AddMember call")
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listPlatformFunc:       func(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error) { return nil, nil },
		listUserTenantsFunc:    func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) { return nil, nil },
		listTenantFunc:         func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) { return nil, nil },
		removePlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error { return nil },
		removeMemberFunc:       func(ctx context.Context, tenantID string, userID uint) error { return nil },
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			require.Equal(t, "tenant-1", tenantID)
			require.Equal(t, uint(7), userID)
			require.Equal(t, "tenant:manage_members", permission)
			return false, nil
		},
	})

	res, err := srv.AddTenantMember(authContextForUser(t, 7), &pbauthv1.AddTenantMemberRequest{
		TenantId: "tenant-1",
		UserId:   9,
		RoleName: iamdomain.RoleTenantAdmin,
	})
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestCheckPermission_InactiveMembershipReturnsNotAllowed(t *testing.T) {
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, iamdomain.ErrInactiveMembership
		},
	})

	res, err := srv.CheckPermission(context.Background(), &pbauthv1.CheckPermissionRequest{
		TenantId:   "tenant-3",
		UserId:     21,
		Permission: "store:update",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Allowed)
}

func TestCheckPlatformPermission_OK(t *testing.T) {
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			require.Equal(t, "", tenantID)
			require.Equal(t, uint(7), userID)
			require.Equal(t, "tenant:create", permission)
			return true, nil
		},
	})

	res, err := srv.CheckPlatformPermission(authContextForUser(t, 7), &pbauthv1.CheckPlatformPermissionRequest{
		Permission: "tenant:create",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Allowed)
}

func TestListPlatformRoles_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 13, 0, 0, 0, time.UTC)
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addPlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error {
			return nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listPlatformFunc: func(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error) {
			require.Equal(t, uint(9), userID)
			return []iamdomain.PlatformMembership{{
				UserID:    userID,
				RoleID:    1,
				RoleName:  iamdomain.RolePlatformOwner,
				Status:    iamdomain.MembershipStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}}, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		removePlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error {
			return nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			if tenantID == "" && userID == 7 && permission == "platform:manage_roles" {
				return true, nil
			}
			return false, nil
		},
	})

	res, err := srv.ListPlatformRoles(authContextForUser(t, 7), &pbauthv1.ListPlatformRolesRequest{
		TargetUserId: 9,
	})
	require.NoError(t, err)
	require.Len(t, res.Memberships, 1)
	assert.Equal(t, iamdomain.RolePlatformOwner, res.Memberships[0].RoleName)
}

func TestAddPlatformRole_OK(t *testing.T) {
	called := false
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addPlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error {
			called = true
			require.Equal(t, uint(15), userID)
			require.Equal(t, iamdomain.RolePlatformAdmin, roleName)
			return nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listPlatformFunc: func(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		removePlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error {
			return nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			if tenantID == "" && userID == 7 && permission == "platform:manage_roles" {
				return true, nil
			}
			return false, nil
		},
	})

	res, err := srv.AddPlatformRole(authContextForUser(t, 7), &pbauthv1.AddPlatformRoleRequest{
		TargetUserId: 15,
		RoleName:     iamdomain.RolePlatformAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, called)
}

func TestRemovePlatformRole_OK(t *testing.T) {
	called := false
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addPlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error {
			return nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listPlatformFunc: func(ctx context.Context, userID uint) ([]iamdomain.PlatformMembership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		removePlatformRoleFunc: func(ctx context.Context, userID uint, roleName string) error {
			called = true
			require.Equal(t, uint(15), userID)
			require.Equal(t, iamdomain.RolePlatformAdmin, roleName)
			return nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			if tenantID == "" && userID == 7 && permission == "platform:manage_roles" {
				return true, nil
			}
			return false, nil
		},
	})

	res, err := srv.RemovePlatformRole(authContextForUser(t, 7), &pbauthv1.RemovePlatformRoleRequest{
		TargetUserId: 15,
		RoleName:     iamdomain.RolePlatformAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, called)
}

func TestSwitchActiveTenant_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := authContextForUser(t, 7)

	uc.On("SwitchActiveTenant", mock.Anything, uint(7), "tenant-9", "access-token").
		Return(&inputport.AuthResult{
			JwtToken: "jwt-tenant",
			UserInfo: entity.User{
				Id:       7,
				Email:    "neo@mx.io",
				Username: "neo",
			},
		}, nil)

	res, err := srv.SwitchActiveTenant(ctx, &pbauthv1.SwitchActiveTenantRequest{
		UserId:      7,
		TenantId:    "tenant-9",
		AccessToken: "access-token",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-tenant", res.JwtToken)
	assert.Equal(t, "neo", res.UserInfo.Username)
}

func TestRefreshToken_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

	uc.On("RefreshAccessToken", mock.Anything, "refresh-token").
		Return(&inputport.AuthResult{
			JwtToken:     "jwt-refreshed",
			RefreshToken: "refresh-next",
			UserInfo: entity.User{
				Id:       7,
				Email:    "neo@mx.io",
				Username: "neo",
			},
		}, nil)

	res, err := srv.RefreshToken(ctx, &pbauthv1.RefreshTokenRequest{RefreshToken: "refresh-token"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-refreshed", res.JwtToken)
	assert.Equal(t, "refresh-next", res.RefreshToken)
	assert.Equal(t, "neo", res.UserInfo.Username)
}

func TestGetSession_OK(t *testing.T) {
	sessionNow := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	srv := NewAuthServer(
		&inputmocks.MockAuthUsecase{},
		&sessionRepoFake{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
			return &entity.Session{
				ID:             id,
				UserID:         21,
				ActiveTenantID: "tenant-7",
				Status:         entity.SessionStatusActive,
				CreatedAt:      sessionNow,
				UpdatedAt:      sessionNow,
				ExpiresAt:      sessionNow.Add(time.Hour),
			}, nil
		}},
		&auditRepoFake{},
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)

	res, err := srv.GetSession(context.Background(), &pbauthv1.GetSessionRequest{SessionId: "session-1"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "session-1", res.Session.Id)
	assert.Equal(t, uint64(21), res.Session.UserId)
	assert.Equal(t, "tenant-7", res.Session.ActiveTenantId)
}

func TestListUserTenants_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC)
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			require.Equal(t, uint(7), userID)
			return []iamdomain.Membership{{
				TenantID:  "tenant-1",
				UserID:    userID,
				RoleID:    1,
				RoleName:  iamdomain.RoleTenantOwner,
				Status:    iamdomain.MembershipStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}}, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, nil
		},
	})

	res, err := srv.ListUserTenants(authContextForUser(t, 7), &pbauthv1.ListUserTenantsRequest{UserId: 7})
	require.NoError(t, err)
	require.Len(t, res.Memberships, 1)
	assert.Equal(t, "tenant-1", res.Memberships[0].TenantId)
	assert.Equal(t, iamdomain.RoleTenantOwner, res.Memberships[0].RoleName)
}

func TestListTenantMembers_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			require.Equal(t, "tenant-9", tenantID)
			return []iamdomain.Membership{{
				TenantID:  tenantID,
				UserID:    11,
				RoleID:    2,
				RoleName:  iamdomain.RoleTenantAdmin,
				Status:    iamdomain.MembershipStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}}, nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			require.Equal(t, "tenant-9", tenantID)
			require.Equal(t, uint(7), userID)
			require.Equal(t, "tenant:manage_members", permission)
			return true, nil
		},
	})

	res, err := srv.ListTenantMembers(
		authContextForUser(t, 7),
		&pbauthv1.ListTenantMembersRequest{TenantId: "tenant-9"},
	)
	require.NoError(t, err)
	require.Len(t, res.Memberships, 1)
	assert.Equal(t, uint64(11), res.Memberships[0].UserId)
	assert.Equal(t, iamdomain.RoleTenantAdmin, res.Memberships[0].RoleName)
}

func TestRemoveTenantMember_OK(t *testing.T) {
	called := false
	srv := newServerWithIAM(&iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, nil
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, nil
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, nil
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			called = true
			require.Equal(t, "tenant-5", tenantID)
			require.Equal(t, uint(33), userID)
			return nil
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			require.Equal(t, "tenant-5", tenantID)
			require.Equal(t, uint(7), userID)
			require.Equal(t, "tenant:manage_members", permission)
			return true, nil
		},
	})

	res, err := srv.RemoveTenantMember(authContextForUser(t, 7), &pbauthv1.RemoveTenantMemberRequest{
		TenantId: "tenant-5",
		UserId:   33,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, called)
}
