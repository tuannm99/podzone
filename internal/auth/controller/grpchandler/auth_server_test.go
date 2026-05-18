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
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGoogleLogin_OK(t *testing.T) {
	srv, uc := newServerWithMock(t)
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
	srv, uc := newServerWithMock(t)
	ctx := context.Background()

	uc.On("GenerateOAuthURL", mock.Anything).Return("", assert.AnError)

	res, err := srv.GoogleLogin(ctx, &pbauthv1.GoogleLoginRequest{})
	require.Error(t, err)
	assert.Nil(t, res)

	uc.AssertExpectations(t)
}

func TestGoogleCallback_OK(t *testing.T) {
	srv, uc := newServerWithMock(t)
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
	srv, uc := newServerWithMock(t)
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
	srv, uc := newServerWithMock(t)
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
	srv, uc := newServerWithMock(t)
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
		newSessionRepoMock(t, sessionRepoMockConfig{
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
		}),
		newAuditRepoMock(t, auditRepoMockConfig{}),
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
		newSessionRepoMock(
			t,
			sessionRepoMockConfig{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
				return nil, entity.ErrSessionNotFound
			}},
		),
		newAuditRepoMock(t, auditRepoMockConfig{
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
		}),
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

func TestCreatePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		createPolicyFunc: func(ctx context.Context, input iamdomain.CreatePolicyInput) (*iamdomain.Policy, []iamdomain.PolicyStatement, error) {
			require.Equal(t, "tenant/orders_editor", input.Name)
			return &iamdomain.Policy{
					ID:          1,
					Scope:       input.Scope,
					Name:        input.Name,
					Description: input.Description,
					CreatedAt:   time.Now().UTC(),
					UpdatedAt:   time.Now().UTC(),
				},
				[]iamdomain.PolicyStatement{{
					ID:              9,
					PolicyID:        1,
					PolicyName:      input.Name,
					Effect:          iamdomain.PolicyEffectAllow,
					ActionPattern:   "order:update",
					ResourcePattern: "*",
					CreatedAt:       time.Now().UTC(),
				}}, nil
		},
	}))

	res, err := srv.CreatePolicy(authContextForUser(t, 7), &pbauthv1.CreatePolicyRequest{
		Scope:       iamdomain.PolicyScopeTenant,
		Name:        "tenant/orders_editor",
		Description: "Edit routed orders",
		Statements: []*pbauthv1.PolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant/orders_editor", res.Policy.Name)
	require.Len(t, res.Statements, 1)
	assert.Equal(t, "order:update", res.Statements[0].ActionPattern)
}

func TestAttachTenantUserPolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return tenantID == "t1" && permission == "tenant:manage_members", nil
		},
		attachTenantPolicyFunc: func(ctx context.Context, tenantID string, userID uint, policyName string) error {
			require.Equal(t, "t1", tenantID)
			require.Equal(t, uint(9), userID)
			require.Equal(t, "tenant/orders_editor", policyName)
			return nil
		},
	}))

	res, err := srv.AttachTenantUserPolicy(authContextForUser(t, 7), &pbauthv1.AttachTenantUserPolicyRequest{
		TenantId:   "t1",
		UserId:     9,
		PolicyName: "tenant/orders_editor",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestGetPolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		getPolicyFunc: func(ctx context.Context, name string) (*iamdomain.Policy, []iamdomain.PolicyStatement, error) {
			return &iamdomain.Policy{
					ID:    1,
					Scope: iamdomain.PolicyScopeTenant,
					Name:  name,
				},
				[]iamdomain.PolicyStatement{{
					ID:              1,
					PolicyID:        1,
					PolicyName:      name,
					Effect:          iamdomain.PolicyEffectAllow,
					ActionPattern:   "order:update",
					ResourcePattern: "*",
					CreatedAt:       time.Now().UTC(),
				}}, nil
		},
	}))

	res, err := srv.GetPolicy(authContextForUser(t, 7), &pbauthv1.GetPolicyRequest{Name: "tenant/orders_editor"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant/orders_editor", res.Policy.Name)
	require.Len(t, res.Statements, 1)
}

func TestListPolicyAttachments_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		listPolicyAttachmentsFunc: func(ctx context.Context, name string) ([]iamdomain.PolicyAttachment, error) {
			require.Equal(t, "tenant/orders_editor", name)
			return []iamdomain.PolicyAttachment{
				{
					AttachmentType: "role",
					RoleID:         2,
					RoleName:       "tenant_editor",
					CreatedAt:      time.Now().UTC(),
				},
				{
					AttachmentType: "group",
					Scope:          iamdomain.PolicyScopeTenant,
					TenantID:       "t1",
					GroupID:        12,
					GroupName:      "ops-team",
					CreatedAt:      time.Now().UTC(),
				},
			}, nil
		},
	}))

	res, err := srv.ListPolicyAttachments(
		authContextForUser(t, 7),
		&pbauthv1.ListPolicyAttachmentsRequest{Name: "tenant/orders_editor"},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Attachments, 2)
	assert.Equal(t, "role", res.Attachments[0].AttachmentType)
	assert.Equal(t, uint64(12), res.Attachments[1].GroupId)
}

func TestDeletePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		deletePolicyFunc: func(ctx context.Context, name string) error {
			require.Equal(t, "tenant/orders_editor", name)
			return nil
		},
	}))

	res, err := srv.DeletePolicy(authContextForUser(t, 7), &pbauthv1.DeletePolicyRequest{Name: "tenant/orders_editor"})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestCreateGroup_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return tenantID == "t1" && permission == "tenant:manage_members", nil
		},
		createGroupFunc: func(ctx context.Context, input iamdomain.CreateGroupInput) (*iamdomain.Group, error) {
			return &iamdomain.Group{
				ID:       12,
				Scope:    input.Scope,
				TenantID: input.TenantID,
				Name:     input.Name,
			}, nil
		},
	}))

	res, err := srv.CreateGroup(authContextForUser(t, 7), &pbauthv1.CreateGroupRequest{
		Scope:    iamdomain.PolicyScopeTenant,
		TenantId: "t1",
		Name:     "ops-team",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "ops-team", res.Group.Name)
}

func TestDeleteGroup_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		deleteGroupFunc: func(ctx context.Context, groupID uint64) error {
			require.Equal(t, uint64(12), groupID)
			return nil
		},
	}))

	res, err := srv.DeleteGroup(authContextForUser(t, 7), &pbauthv1.DeleteGroupRequest{GroupId: 12})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestListGroupMembers_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		listGroupMembersFunc: func(ctx context.Context, groupID uint64) ([]uint, error) {
			require.Equal(t, uint64(12), groupID)
			return []uint{7, 9}, nil
		},
	}))

	res, err := srv.ListGroupMembers(authContextForUser(t, 7), &pbauthv1.ListGroupMembersRequest{
		GroupId: 12,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, []uint64{7, 9}, res.UserIds)
}

func TestDetachGroupPolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		detachGroupPolicyFunc: func(ctx context.Context, groupID uint64, policyName string) error {
			require.Equal(t, uint64(12), groupID)
			require.Equal(t, "tenant/orders_editor", policyName)
			return nil
		},
	}))

	res, err := srv.DetachGroupPolicy(authContextForUser(t, 7), &pbauthv1.DetachGroupPolicyRequest{
		GroupId:    12,
		PolicyName: "tenant/orders_editor",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestPutGroupInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		putGroupInlinePolicyFunc: func(ctx context.Context, input iamdomain.PutGroupInlinePolicyInput) error {
			require.Equal(t, uint64(12), input.GroupID)
			require.Equal(t, "inline-ops", input.Name)
			require.Len(t, input.Statements, 1)
			require.Equal(t, "order:update", input.Statements[0].ActionPattern)
			return nil
		},
	}))

	res, err := srv.PutGroupInlinePolicy(authContextForUser(t, 7), &pbauthv1.PutGroupInlinePolicyRequest{
		GroupId:     12,
		Name:        "inline-ops",
		Description: "Inline order ops access",
		Statements: []*pbauthv1.PolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestGetGroupInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		getGroupInlinePolicyFunc: func(ctx context.Context, groupID uint64, name string) (*iamdomain.GroupInlinePolicy, error) {
			require.Equal(t, uint64(12), groupID)
			require.Equal(t, "inline-ops", name)
			return &iamdomain.GroupInlinePolicy{
				GroupID:     groupID,
				Name:        name,
				Description: "Inline order ops access",
				Statements: []iamdomain.PolicyStatement{{
					Effect:          iamdomain.PolicyEffectAllow,
					ActionPattern:   "order:update",
					ResourcePattern: "*",
					CreatedAt:       time.Now().UTC(),
				}},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}))

	res, err := srv.GetGroupInlinePolicy(authContextForUser(t, 7), &pbauthv1.GetGroupInlinePolicyRequest{
		GroupId: 12,
		Name:    "inline-ops",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "inline-ops", res.Policy.Name)
	require.Len(t, res.Policy.Statements, 1)
}

func TestPutPlatformUserInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		putPlatformUserInlinePolicyFunc: func(ctx context.Context, input iamdomain.PutPlatformUserInlinePolicyInput) error {
			require.Equal(t, uint(21), input.UserID)
			require.Equal(t, "inline-platform", input.Name)
			require.Len(t, input.Statements, 1)
			return nil
		},
	}))

	res, err := srv.PutPlatformUserInlinePolicy(authContextForUser(t, 7), &pbauthv1.PutPlatformUserInlinePolicyRequest{
		TargetUserId: 21,
		Name:         "inline-platform",
		Description:  "Platform inline access",
		Statements: []*pbauthv1.PolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "tenant:create",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestGetPlatformUserInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		getPlatformUserInlinePolicyFunc: func(ctx context.Context, userID uint, name string) (*iamdomain.UserInlinePolicy, error) {
			require.Equal(t, uint(21), userID)
			require.Equal(t, "inline-platform", name)
			return &iamdomain.UserInlinePolicy{
				Scope:       iamdomain.PolicyScopePlatform,
				UserID:      userID,
				Name:        name,
				Description: "Platform inline access",
				Statements: []iamdomain.PolicyStatement{{
					Effect:          iamdomain.PolicyEffectAllow,
					ActionPattern:   "tenant:create",
					ResourcePattern: "*",
					CreatedAt:       time.Now().UTC(),
				}},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}))

	res, err := srv.GetPlatformUserInlinePolicy(authContextForUser(t, 7), &pbauthv1.GetPlatformUserInlinePolicyRequest{
		TargetUserId: 21,
		Name:         "inline-platform",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "inline-platform", res.Policy.Name)
	require.Len(t, res.Policy.Statements, 1)
}

func TestListPlatformUserInlinePolicies_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		listPlatformUserInlinePoliciesFunc: func(ctx context.Context, userID uint) ([]iamdomain.UserInlinePolicy, error) {
			require.Equal(t, uint(21), userID)
			return []iamdomain.UserInlinePolicy{{
				Scope:       iamdomain.PolicyScopePlatform,
				UserID:      userID,
				Name:        "inline-platform",
				Description: "Platform inline access",
				Statements: []iamdomain.PolicyStatement{{
					Effect:          iamdomain.PolicyEffectAllow,
					ActionPattern:   "tenant:create",
					ResourcePattern: "*",
					CreatedAt:       time.Now().UTC(),
				}},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}}, nil
		},
	}))

	res, err := srv.ListPlatformUserInlinePolicies(
		authContextForUser(t, 7),
		&pbauthv1.ListPlatformUserInlinePoliciesRequest{
			TargetUserId: 21,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Policies, 1)
	assert.Equal(t, "inline-platform", res.Policies[0].Name)
}

func TestDeletePlatformUserInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		deletePlatformUserInlinePolicyFunc: func(ctx context.Context, userID uint, name string) error {
			require.Equal(t, uint(21), userID)
			require.Equal(t, "inline-platform", name)
			return nil
		},
	}))

	res, err := srv.DeletePlatformUserInlinePolicy(
		authContextForUser(t, 7),
		&pbauthv1.DeletePlatformUserInlinePolicyRequest{
			TargetUserId: 21,
			Name:         "inline-platform",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestPutTenantUserInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return tenantID == "t1" && permission == "tenant:manage_members", nil
		},
		putTenantUserInlinePolicyFunc: func(ctx context.Context, input iamdomain.PutTenantUserInlinePolicyInput) error {
			require.Equal(t, "t1", input.TenantID)
			require.Equal(t, uint(9), input.UserID)
			require.Equal(t, "inline-tenant", input.Name)
			require.Len(t, input.Statements, 1)
			return nil
		},
	}))

	res, err := srv.PutTenantUserInlinePolicy(authContextForUser(t, 7), &pbauthv1.PutTenantUserInlinePolicyRequest{
		TenantId:    "t1",
		UserId:      9,
		Name:        "inline-tenant",
		Description: "Tenant inline access",
		Statements: []*pbauthv1.PolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:update",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestGetTenantUserInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return tenantID == "t1" && permission == "tenant:manage_members", nil
		},
		getTenantUserInlinePolicyFunc: func(ctx context.Context, tenantID string, userID uint, name string) (*iamdomain.UserInlinePolicy, error) {
			require.Equal(t, "t1", tenantID)
			require.Equal(t, uint(9), userID)
			require.Equal(t, "inline-tenant", name)
			return &iamdomain.UserInlinePolicy{
				Scope:       iamdomain.PolicyScopeTenant,
				TenantID:    tenantID,
				UserID:      userID,
				Name:        name,
				Description: "Tenant inline access",
				Statements: []iamdomain.PolicyStatement{{
					Effect:          iamdomain.PolicyEffectAllow,
					ActionPattern:   "order:update",
					ResourcePattern: "*",
					CreatedAt:       time.Now().UTC(),
				}},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}))

	res, err := srv.GetTenantUserInlinePolicy(authContextForUser(t, 7), &pbauthv1.GetTenantUserInlinePolicyRequest{
		TenantId: "t1",
		UserId:   9,
		Name:     "inline-tenant",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "inline-tenant", res.Policy.Name)
	require.Len(t, res.Policy.Statements, 1)
}

func TestListTenantUserInlinePolicies_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return tenantID == "t1" && permission == "tenant:manage_members", nil
		},
		listTenantUserInlinePoliciesFunc: func(ctx context.Context, tenantID string, userID uint) ([]iamdomain.UserInlinePolicy, error) {
			require.Equal(t, "t1", tenantID)
			require.Equal(t, uint(9), userID)
			return []iamdomain.UserInlinePolicy{{
				Scope:       iamdomain.PolicyScopeTenant,
				TenantID:    tenantID,
				UserID:      userID,
				Name:        "inline-tenant",
				Description: "Tenant inline access",
				Statements: []iamdomain.PolicyStatement{{
					Effect:          iamdomain.PolicyEffectAllow,
					ActionPattern:   "order:update",
					ResourcePattern: "*",
					CreatedAt:       time.Now().UTC(),
				}},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}}, nil
		},
	}))

	res, err := srv.ListTenantUserInlinePolicies(
		authContextForUser(t, 7),
		&pbauthv1.ListTenantUserInlinePoliciesRequest{
			TenantId: "t1",
			UserId:   9,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Policies, 1)
	assert.Equal(t, "inline-tenant", res.Policies[0].Name)
}

func TestDeleteTenantUserInlinePolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return tenantID == "t1" && permission == "tenant:manage_members", nil
		},
		deleteTenantUserInlinePolicyFunc: func(ctx context.Context, tenantID string, userID uint, name string) error {
			require.Equal(t, "t1", tenantID)
			require.Equal(t, uint(9), userID)
			require.Equal(t, "inline-tenant", name)
			return nil
		},
	}))

	res, err := srv.DeleteTenantUserInlinePolicy(
		authContextForUser(t, 7),
		&pbauthv1.DeleteTenantUserInlinePolicyRequest{
			TenantId: "t1",
			UserId:   9,
			Name:     "inline-tenant",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestLogin_OK(t *testing.T) {
	srv, uc := newServerWithMock(t)
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
	srv, uc := newServerWithMock(t)
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
	srv, uc := newServerWithMock(t)
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
	srv, uc := newServerWithMock(t)
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
	srv := NewIAMServer(newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}), newAuditRepoMock(t, auditRepoMockConfig{createFunc: func(ctx context.Context, log entity.AuditLog) error {
		createdAudit = log
		return nil
	}}), &outputmocks.MockUserRepository{}, testAuthCfg)

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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
	srv := NewIAMServer(newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}), newAuditRepoMock(t, auditRepoMockConfig{}), userRepo, testAuthCfg)

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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

	srv := NewIAMServer(newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}), newAuditRepoMock(t, auditRepoMockConfig{}), userRepo, testAuthCfg)

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

	res, err := srv.CheckPlatformPermission(authContextForUser(t, 7), &pbauthv1.CheckPlatformPermissionRequest{
		Permission: "tenant:create",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Allowed)
}

func TestListPlatformRoles_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 13, 0, 0, 0, time.UTC)
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

	res, err := srv.ListPlatformRoles(authContextForUser(t, 7), &pbauthv1.ListPlatformRolesRequest{
		TargetUserId: 9,
	})
	require.NoError(t, err)
	require.Len(t, res.Memberships, 1)
	assert.Equal(t, iamdomain.RolePlatformOwner, res.Memberships[0].RoleName)
}

func TestAddPlatformRole_OK(t *testing.T) {
	called := false
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

	res, err := srv.RemovePlatformRole(authContextForUser(t, 7), &pbauthv1.RemovePlatformRoleRequest{
		TargetUserId: 15,
		RoleName:     iamdomain.RolePlatformAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, called)
}

func TestSwitchActiveTenant_OK(t *testing.T) {
	srv, uc := newServerWithMock(t)
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

func TestAssumeSessionPolicy_OK(t *testing.T) {
	accessToken := accessTokenForSession(
		t,
		entity.User{Id: 7, Email: "neo@mx.io", Username: "neo"},
		"tenant-9",
		"session-test",
	)
	srv := NewAuthServer(
		&inputmocks.MockAuthUsecase{},
		newSessionRepoMock(t, sessionRepoMockConfig{
			getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
				require.Equal(t, "session-test", id)
				return &entity.Session{
					ID:             id,
					UserID:         7,
					ActiveTenantID: "tenant-9",
					SessionPolicy: []entity.SessionPolicyStatement{{
						Effect:          iamdomain.PolicyEffectAllow,
						ActionPattern:   "order:read",
						ResourcePattern: "*",
					}},
					Status:    entity.SessionStatusActive,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			},
		}),
		newAuditRepoMock(t, auditRepoMockConfig{}),
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)
	authUC := srv.authUC.(*inputmocks.MockAuthUsecase)
	authUC.On("AssumeSessionPolicy", mock.Anything, uint(7), accessToken, mock.MatchedBy(func(items []entity.SessionPolicyStatement) bool {
		return len(items) == 1 && items[0].ActionPattern == "order:read"
	})).Return(&inputport.AuthResult{
		JwtToken: "jwt-scoped",
		UserInfo: entity.User{Id: 7, Username: "neo"},
	}, nil)

	res, err := srv.AssumeSessionPolicy(authContextForUser(t, 7), &pbauthv1.AssumeSessionPolicyRequest{
		AccessToken: accessToken,
		Statements: []*pbauthv1.PolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:read",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-scoped", res.JwtToken)
	require.Len(t, res.Session.SessionPolicy, 1)
	assert.Equal(t, "order:read", res.Session.SessionPolicy[0].ActionPattern)
}

func TestClearSessionPolicy_OK(t *testing.T) {
	accessToken := accessTokenForSession(
		t,
		entity.User{Id: 7, Email: "neo@mx.io", Username: "neo"},
		"tenant-9",
		"session-test",
	)
	srv := NewAuthServer(
		&inputmocks.MockAuthUsecase{},
		newSessionRepoMock(t, sessionRepoMockConfig{
			getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
				require.Equal(t, "session-test", id)
				return &entity.Session{
					ID:             id,
					UserID:         7,
					ActiveTenantID: "tenant-9",
					Status:         entity.SessionStatusActive,
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					ExpiresAt:      time.Now().Add(time.Hour),
				}, nil
			},
		}),
		newAuditRepoMock(t, auditRepoMockConfig{}),
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)
	authUC := srv.authUC.(*inputmocks.MockAuthUsecase)
	authUC.On("ClearSessionPolicy", mock.Anything, uint(7), accessToken).Return(&inputport.AuthResult{
		JwtToken: "jwt-cleared",
		UserInfo: entity.User{Id: 7, Username: "neo"},
	}, nil)

	res, err := srv.ClearSessionPolicy(authContextForUser(t, 7), &pbauthv1.ClearSessionPolicyRequest{
		AccessToken: accessToken,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-cleared", res.JwtToken)
	assert.Empty(t, res.Session.SessionPolicy)
}

func TestAssumeRole_OK(t *testing.T) {
	accessToken := accessTokenForSession(
		t,
		entity.User{Id: 7, Email: "neo@mx.io", Username: "neo"},
		"",
		"session-test",
	)
	srv := NewAuthServer(
		&inputmocks.MockAuthUsecase{},
		newSessionRepoMock(t, sessionRepoMockConfig{
			getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
				return &entity.Session{
					ID:                  id,
					UserID:              7,
					ActiveTenantID:      "tenant-9",
					AssumedRoleID:       2,
					AssumedRoleScope:    iamdomain.PolicyScopeTenant,
					AssumedRoleName:     iamdomain.RoleTenantAdmin,
					AssumedRoleTenantID: "tenant-9",
					Status:              entity.SessionStatusActive,
					CreatedAt:           time.Now().UTC(),
					UpdatedAt:           time.Now().UTC(),
					ExpiresAt:           time.Now().Add(time.Hour),
				}, nil
			},
		}),
		newAuditRepoMock(t, auditRepoMockConfig{}),
		&outputmocks.MockUserRepository{},
		testAuthCfg,
	)
	authUC := srv.authUC.(*inputmocks.MockAuthUsecase)
	authUC.On("AssumeRole", mock.Anything, uint(7), accessToken, iamdomain.RoleTenantAdmin, "tenant-9", mock.MatchedBy(func(items []entity.SessionPolicyStatement) bool {
		return len(items) == 1 && items[0].ActionPattern == "order:read"
	}), "", "", "", uint32(0), "", mock.MatchedBy(func(tags map[string]string) bool {
		return len(tags) == 0
	})).Return(&inputport.AuthResult{
		JwtToken: "jwt-assumed",
		UserInfo: entity.User{Id: 7, Username: "neo"},
	}, nil)

	res, err := srv.AssumeRole(authContextForUser(t, 7), &pbauthv1.AssumeRoleRequest{
		AccessToken: accessToken,
		RoleName:    iamdomain.RoleTenantAdmin,
		TenantId:    "tenant-9",
		SessionPolicy: []*pbauthv1.PolicyStatement{{
			Effect:          iamdomain.PolicyEffectAllow,
			ActionPattern:   "order:read",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "jwt-assumed", res.JwtToken)
	assert.Equal(t, iamdomain.RoleTenantAdmin, res.Session.AssumedRoleName)
}

func TestGetRoleTrustPolicy_OK(t *testing.T) {
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return permission == "platform:manage_roles", nil
		},
		getRoleTrustPolicyFunc: func(ctx context.Context, roleName string) ([]iamdomain.RoleTrustStatement, error) {
			require.Equal(t, iamdomain.RoleTenantAdmin, roleName)
			return []iamdomain.RoleTrustStatement{{
				RoleID:           2,
				Effect:           iamdomain.PolicyEffectAllow,
				PrincipalType:    iamdomain.TrustPrincipalTenantRole,
				PrincipalPattern: iamdomain.RoleTenantOwner,
				TenantPattern:    "*",
				CreatedAt:        time.Now().UTC(),
			}}, nil
		},
	}))

	res, err := srv.GetRoleTrustPolicy(authContextForUser(t, 7), &pbauthv1.GetRoleTrustPolicyRequest{
		RoleName: iamdomain.RoleTenantAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Statements, 1)
	assert.Equal(t, iamdomain.TrustPrincipalTenantRole, res.Statements[0].PrincipalType)
}

func TestRefreshToken_OK(t *testing.T) {
	srv, uc := newServerWithMock(t)
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
		newSessionRepoMock(
			t,
			sessionRepoMockConfig{getByIDFunc: func(ctx context.Context, id string) (*entity.Session, error) {
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
		),
		newAuditRepoMock(t, auditRepoMockConfig{}),
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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

	res, err := srv.ListUserTenants(authContextForUser(t, 7), &pbauthv1.ListUserTenantsRequest{UserId: 7})
	require.NoError(t, err)
	require.Len(t, res.Memberships, 1)
	assert.Equal(t, "tenant-1", res.Memberships[0].TenantId)
	assert.Equal(t, iamdomain.RoleTenantOwner, res.Memberships[0].RoleName)
}

func TestListTenantMembers_OK(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

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
	srv := newServerWithIAM(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
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
	}))

	res, err := srv.RemoveTenantMember(authContextForUser(t, 7), &pbauthv1.RemoveTenantMemberRequest{
		TenantId: "tenant-5",
		UserId:   33,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, called)
}
