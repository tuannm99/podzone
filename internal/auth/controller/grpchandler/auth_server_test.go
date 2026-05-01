package grpchandler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type iamUsecaseFake struct {
	createTenantFunc    func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error)
	addMemberFunc       func(ctx context.Context, tenantID string, userID uint, roleName string) error
	getMembershipFunc   func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error)
	listUserTenantsFunc func(ctx context.Context, userID uint) ([]iamdomain.Membership, error)
	listTenantFunc      func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error)
	removeMemberFunc    func(ctx context.Context, tenantID string, userID uint) error
	checkPermissionFunc func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error)
}

func (f *iamUsecaseFake) CreateTenant(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
	return f.createTenantFunc(ctx, ownerUserID, cmd)
}

func (f *iamUsecaseFake) AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error {
	return f.addMemberFunc(ctx, tenantID, userID, roleName)
}

func (f *iamUsecaseFake) GetMembership(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
	return f.getMembershipFunc(ctx, tenantID, userID)
}

func (f *iamUsecaseFake) CheckPermission(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
	return f.checkPermissionFunc(ctx, tenantID, userID, permission)
}

func (f *iamUsecaseFake) ListUserTenants(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
	return f.listUserTenantsFunc(ctx, userID)
}

func (f *iamUsecaseFake) ListTenantMembers(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
	return f.listTenantFunc(ctx, tenantID)
}

func (f *iamUsecaseFake) RemoveMember(ctx context.Context, tenantID string, userID uint) error {
	return f.removeMemberFunc(ctx, tenantID, userID)
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

type sessionRepoFake struct {
	getByIDFunc func(ctx context.Context, id string) (*entity.Session, error)
}

func (f *sessionRepoFake) Create(ctx context.Context, session entity.Session) error { return nil }
func (f *sessionRepoFake) UpdateActiveTenant(ctx context.Context, id, tenantID string, updatedAt time.Time) error {
	return nil
}
func (f *sessionRepoFake) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	return nil
}
func (f *sessionRepoFake) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	return f.getByIDFunc(ctx, id)
}

func newServerWithMock() (*AuthServer, *inputmocks.MockAuthUsecase) {
	authUC := &inputmocks.MockAuthUsecase{}
	iamUC := &iamUsecaseFake{
		createTenantFunc: func(ctx context.Context, ownerUserID uint, cmd iamdomain.CreateTenantCmd) (*iamdomain.Tenant, error) {
			return nil, errors.New("unexpected CreateTenant call")
		},
		addMemberFunc: func(ctx context.Context, tenantID string, userID uint, roleName string) error {
			return errors.New("unexpected AddMember call")
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamdomain.Membership, error) {
			return nil, errors.New("unexpected GetMembership call")
		},
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
			return nil, errors.New("unexpected ListUserTenants call")
		},
		listTenantFunc: func(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
			return nil, errors.New("unexpected ListTenantMembers call")
		},
		removeMemberFunc: func(ctx context.Context, tenantID string, userID uint) error {
			return errors.New("unexpected RemoveMember call")
		},
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, errors.New("unexpected CheckPermission call")
		},
	}
	srv := NewAuthServer(authUC, iamUC, &sessionRepoFake{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
		return nil, entity.ErrSessionNotFound
	}})
	return srv, authUC
}

func newServerWithIAM(iamUC iamdomain.IAMUsecase) *AuthServer {
	return NewAuthServer(&inputmocks.MockAuthUsecase{}, iamUC, &sessionRepoFake{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
		return nil, entity.ErrSessionNotFound
	}})
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
		JwtToken:    "jwt-abc",
		RedirectUrl: "https://app.example.com?token=jwt-abc",
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
	assert.Equal(t, "jwt-abc", res.JwtToken)
	assert.Equal(t, "https://app.example.com?token=jwt-abc", res.RedirectUrl)
	assert.Equal(t, "neo@mx.io", res.UserInfo.Email)
	assert.Equal(t, "Neo", res.UserInfo.Name)

	uc.AssertExpectations(t)
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
	srv := newServerWithIAM(&iamUsecaseFake{
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
			return false, nil
		},
	})

	res, err := srv.CreateTenant(context.Background(), &pbauthv1.CreateTenantRequest{
		OwnerUserId: 7,
		Slug:        "seller-one",
		Name:        "Seller One",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-1", res.Tenant.Id)
	assert.Equal(t, iamdomain.RoleTenantOwner, res.OwnerMembership.RoleName)
	assert.Equal(t, "2026-04-30T12:00:00Z", res.Tenant.CreatedAt)
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
			return false, nil
		},
	})

	res, err := srv.AddTenantMember(context.Background(), &pbauthv1.AddTenantMemberRequest{
		TenantId: "missing",
		UserId:   9,
		RoleName: iamdomain.RoleTenantAdmin,
	})
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, codes.NotFound, status.Code(err))
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

func TestSwitchActiveTenant_OK(t *testing.T) {
	srv, uc := newServerWithMock()
	ctx := context.Background()

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
		&iamUsecaseFake{
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
				return false, nil
			},
		},
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

	res, err := srv.ListUserTenants(context.Background(), &pbauthv1.ListUserTenantsRequest{UserId: 7})
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
			return false, nil
		},
	})

	res, err := srv.ListTenantMembers(context.Background(), &pbauthv1.ListTenantMembersRequest{TenantId: "tenant-9"})
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
			return false, nil
		},
	})

	res, err := srv.RemoveTenantMember(context.Background(), &pbauthv1.RemoveTenantMemberRequest{
		TenantId: "tenant-5",
		UserId:   33,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, called)
}
