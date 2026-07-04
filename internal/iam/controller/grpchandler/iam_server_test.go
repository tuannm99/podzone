package grpchandler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	iamentity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	pbcommonv1 "github.com/tuannm99/podzone/pkg/api/proto/common/v1"
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
	"github.com/tuannm99/podzone/pkg/collection"
)

func TestIAMStatusErrorIncludesMissingPermissionDetails(t *testing.T) {
	t.Parallel()

	err := iamStatusError(iamentity.NewPermissionDeniedError(
		"platform:manage_roles",
		"podzone:platform",
	))
	grpcStatus, ok := status.FromError(err)

	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, grpcStatus.Code())
	require.Equal(
		t,
		`iam: missing permission "platform:manage_roles" on "podzone:platform"`,
		grpcStatus.Message(),
	)
	require.Len(t, grpcStatus.Details(), 1)
	detail, ok := grpcStatus.Details()[0].(*errdetails.ErrorInfo)
	require.True(t, ok)
	require.Equal(t, "IAM_PERMISSION_DENIED", detail.Reason)
	require.Equal(t, "platform:manage_roles", detail.Metadata["permission"])
	require.Equal(t, "podzone:platform", detail.Metadata["resource"])
}

func TestCreateTenant_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamentity.Membership, error) {
			return nil, nil
		},
		createTenantFunc: func(
			ctx context.Context,
			ownerUserID uint,
			cmd iamentity.CreateTenantCmd,
		) (*iamentity.Tenant, error) {
			return &iamentity.Tenant{ID: "tenant-1", Name: cmd.Name, Slug: cmd.Slug}, nil
		},
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamentity.Membership, error) {
			return &iamentity.Membership{
				TenantID: tenantID,
				UserID:   userID,
				RoleName: iamentity.RoleTenantOwner,
				Status:   iamentity.MembershipStatusActive,
			}, nil
		},
	}))

	res, err := srv.CreateTenant(authContextForIAMUser(t, 7), &pbiamv1.CreateTenantRequest{
		OwnerUserId: 7,
		Name:        "Demo Tenant",
		Slug:        "demo-tenant",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-1", res.Tenant.Id)
	assert.Equal(t, uint64(7), res.OwnerMembership.UserId)
}

func TestCreateTenant_RequiresPlatformPermissionAfterFirstWorkspace(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamentity.Membership, error) {
			return []iamentity.Membership{{TenantID: "existing", UserID: userID}}, nil
		},
		checkPlatformPermissionFunc: func(ctx context.Context, userID uint, permission string) (bool, error) {
			return false, nil
		},
	}))

	_, err := srv.CreateTenant(authContextForIAMUser(t, 7), &pbiamv1.CreateTenantRequest{
		OwnerUserId: 7,
		Name:        "Second Tenant",
		Slug:        "second-tenant",
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestCheckPermission_InactiveMembershipReturnsNotAllowed(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, iamentity.ErrInactiveMembership
		},
	}))

	res, err := srv.CheckPermission(authContextForIAMUser(t, 9), &pbiamv1.CheckPermissionRequest{
		TenantId:   "tenant-1",
		UserId:     9,
		Permission: "order:update",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Allowed)
}

func TestCheckPermission_MissingMembershipReturnsNotAllowed(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPermissionFunc: func(ctx context.Context, tenantID string, userID uint, permission string) (bool, error) {
			return false, iamentity.ErrMembershipNotFound
		},
	}))

	res, err := srv.CheckPermission(authContextForIAMUser(t, 9), &pbiamv1.CheckPermissionRequest{
		TenantId:   "tenant-1",
		UserId:     9,
		Permission: "store:read",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Allowed)
}

func TestCheckPermission_RequiresAuthenticatedActor(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{}))

	res, err := srv.CheckPermission(context.Background(), &pbiamv1.CheckPermissionRequest{
		TenantId:   "tenant-1",
		UserId:     9,
		Permission: "store:read",
	})

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestCheckPermission_RejectsDifferentActor(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{}))

	res, err := srv.CheckPermission(authContextForIAMUser(t, 7), &pbiamv1.CheckPermissionRequest{
		TenantId:   "tenant-1",
		UserId:     9,
		Permission: "store:read",
	})

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestCreatePolicy_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		requirePlatformPermissionFunc: func(ctx context.Context, userID uint, permission string) error { return nil },
		createPolicyFunc: func(
			ctx context.Context,
			input iamentity.CreatePolicyInput,
		) (*iamentity.Policy, []iamentity.PolicyStatement, error) {
			return &iamentity.Policy{Name: input.Name, Scope: input.Scope}, input.Statements, nil
		},
	}))

	res, err := srv.CreatePolicy(authContextForIAMUser(t, 7), &pbiamv1.CreatePolicyRequest{
		Scope: "platform",
		Name:  "managed/test",
		Statements: []*pbcommonv1.PolicyStatement{{
			Effect:          "allow",
			ActionPattern:   "tenant:create",
			ResourcePattern: "*",
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "managed/test", res.Policy.Name)
	assert.Len(t, res.Statements, 1)
}

func TestListIAMCollections_MapCollectionQueryAndPageInfo(t *testing.T) {
	t.Parallel()

	assertQuery := func(t *testing.T, query collection.Query) {
		t.Helper()
		assert.Equal(t, 2, query.Page)
		assert.Equal(t, 5, query.PageSize)
		assert.Equal(t, "ops", query.Search)
		assert.Equal(t, "name", query.SortBy)
		assert.Equal(t, collection.SortAscending, query.SortDirection)
	}
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPlatformPermissionFunc: func(context.Context, uint, string) (bool, error) {
			return true, nil
		},
		requirePlatformPermissionFunc: func(context.Context, uint, string) error {
			return nil
		},
		listOrganizationsFunc: func(
			ctx context.Context,
			query collection.Query,
		) (collection.Page[iamentity.Organization], error) {
			assertQuery(t, query)
			return collection.NewPage(
				[]iamentity.Organization{{ID: "org-1", Name: "Ops"}},
				11,
				query,
			), nil
		},
		listPoliciesFunc: func(
			ctx context.Context,
			scope string,
			query collection.Query,
		) (collection.Page[iamentity.Policy], error) {
			assert.Equal(t, "platform", scope)
			assertQuery(t, query)
			return collection.NewPage(
				[]iamentity.Policy{{ID: 1, Name: "ops", Scope: scope}},
				11,
				query,
			), nil
		},
		listGroupsFunc: func(
			ctx context.Context,
			scope string,
			tenantID string,
			query collection.Query,
		) (collection.Page[iamentity.Group], error) {
			assert.Equal(t, "platform", scope)
			assert.Empty(t, tenantID)
			assertQuery(t, query)
			return collection.NewPage(
				[]iamentity.Group{{ID: 1, Name: "ops", Scope: scope}},
				11,
				query,
			), nil
		},
	}))
	request := &pbcommonv1.CollectionRequest{
		Page:          2,
		PageSize:      5,
		Search:        "ops",
		SortBy:        "name",
		SortDirection: pbcommonv1.SortDirection_SORT_DIRECTION_ASC,
	}

	organizations, err := srv.ListOrganizations(
		authContextForIAMUser(t, 7),
		&pbiamv1.ListOrganizationsRequest{Collection: request},
	)
	require.NoError(t, err)
	assert.Len(t, organizations.Organizations, 1)
	assert.Equal(t, int64(11), organizations.PageInfo.Total)
	assert.Equal(t, int32(3), organizations.PageInfo.TotalPages)

	policies, err := srv.ListPolicies(
		authContextForIAMUser(t, 7),
		&pbiamv1.ListPoliciesRequest{Scope: "platform", Collection: request},
	)
	require.NoError(t, err)
	assert.Len(t, policies.Policies, 1)
	assert.Equal(t, int32(2), policies.PageInfo.Page)

	groups, err := srv.ListGroups(
		authContextForIAMUser(t, 7),
		&pbiamv1.ListGroupsRequest{Scope: "platform", Collection: request},
	)
	require.NoError(t, err)
	assert.Len(t, groups.Groups, 1)
	assert.True(t, groups.PageInfo.HasNext)
}

func TestListOrganizationsFiltersForOrganizationRoot(t *testing.T) {
	t.Parallel()

	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		checkPlatformPermissionFunc: func(context.Context, uint, string) (bool, error) {
			return false, nil
		},
		listOrganizationsForUserFunc: func(
			_ context.Context,
			userID uint,
			query collection.Query,
		) (collection.Page[iamentity.Organization], error) {
			assert.Equal(t, uint(7), userID)
			return collection.NewPage([]iamentity.Organization{{
				ID:         "org-1",
				Name:       "Root org",
				RootUserID: userID,
			}}, 1, query), nil
		},
	}))

	response, err := srv.ListOrganizations(
		authContextForIAMUser(t, 7),
		&pbiamv1.ListOrganizationsRequest{},
	)

	require.NoError(t, err)
	require.Len(t, response.Organizations, 1)
	assert.False(t, response.CanManagePlatform)
	assert.Equal(t, uint64(7), response.Organizations[0].RootUserId)
}

func TestListOrganizationMembersRequiresOrganizationReadPermission(t *testing.T) {
	t.Parallel()

	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		requireOrganizationPermissionFunc: func(
			_ context.Context,
			orgID string,
			userID uint,
			permission string,
		) error {
			assert.Equal(t, "org-1", orgID)
			assert.Equal(t, uint(7), userID)
			assert.Equal(t, "organization:read", permission)
			return nil
		},
		listOrganizationMembersFunc: func(
			_ context.Context,
			orgID string,
			query collection.Query,
		) (collection.Page[iamentity.OrganizationMembership], error) {
			assert.Equal(t, "org-1", orgID)
			return collection.NewPage([]iamentity.OrganizationMembership{{
				OrgID:    orgID,
				UserID:   7,
				RoleID:   11,
				RoleName: iamentity.RoleOrganizationRoot,
				Status:   iamentity.MembershipStatusActive,
			}}, 1, query), nil
		},
	}))

	response, err := srv.ListOrganizationMembers(
		authContextForIAMUser(t, 7),
		&pbiamv1.ListOrganizationMembersRequest{OrgId: "org-1"},
	)

	require.NoError(t, err)
	require.Len(t, response.Memberships, 1)
	assert.Equal(t, uint64(7), response.Memberships[0].UserId)
	assert.Equal(t, iamentity.RoleOrganizationRoot, response.Memberships[0].RoleName)
}

func TestGetTenantMembership_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		getMembershipFunc: func(ctx context.Context, tenantID string, userID uint) (*iamentity.Membership, error) {
			return &iamentity.Membership{
				TenantID: tenantID,
				UserID:   userID,
				RoleName: iamentity.RoleTenantAdmin,
				Status:   iamentity.MembershipStatusActive,
			}, nil
		},
	}))

	res, err := srv.GetTenantMembership(context.Background(), &pbiamv1.GetTenantMembershipRequest{
		TenantId: "tenant-1",
		UserId:   9,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "tenant-1", res.Membership.TenantId)
	assert.Equal(t, uint64(9), res.Membership.UserId)
}

func TestListUserTenants_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		listUserTenantsFunc: func(ctx context.Context, userID uint) ([]iamentity.Membership, error) {
			return []iamentity.Membership{
				{
					TenantID: "tenant-1",
					UserID:   userID,
					RoleName: iamentity.RoleTenantAdmin,
					Status:   iamentity.MembershipStatusActive,
				},
			}, nil
		},
	}))

	res, err := srv.ListUserTenants(authContextForIAMUser(t, 7), &pbiamv1.ListUserTenantsRequest{UserId: 7})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Memberships, 1)
	assert.Equal(t, "tenant-1", res.Memberships[0].TenantId)
}

func TestAssumeRoleRPC_OK(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		assumeRoleFunc: func(ctx context.Context, input iamentity.AssumeRoleInput) (*iamentity.AssumedRole, error) {
			return &iamentity.AssumedRole{
				RoleID:    5,
				RoleScope: iamentity.PolicyScopeTenant,
				RoleName:  iamentity.RoleTenantAdmin,
				TenantID:  input.TenantID,
				ExpiresAt: nowPlusHour(),
			}, nil
		},
	}))

	res, err := srv.AssumeRole(context.Background(), &pbiamv1.IAMAssumeRoleRequest{
		AccessToken: rawAccessTokenForIAMUser(t, 7),
		RoleName:    iamentity.RoleTenantAdmin,
		TenantId:    "tenant-1",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(5), res.AssumedRole.RoleId)
	assert.Equal(t, "tenant-1", res.AssumedRole.TenantId)
}

func TestEnsureRootOrganization_UsesAuthenticatedSelfServiceUser(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{
		ensureRootOrganizationFunc: func(
			_ context.Context,
			rootUserID uint,
			name string,
			slug string,
		) (*iamentity.Organization, error) {
			assert.Equal(t, uint(7), rootUserID)
			return &iamentity.Organization{
				ID:         "org-1",
				Name:       name,
				Slug:       slug,
				RootUserID: rootUserID,
			}, nil
		},
	}))

	res, err := srv.EnsureRootOrganization(
		authContextForIAMUser(t, 7),
		&pbiamv1.EnsureRootOrganizationRequest{Name: "Neo", Slug: "account-7"},
	)
	require.NoError(t, err)
	require.NotNil(t, res.Organization)
	assert.Equal(t, uint64(7), res.Organization.RootUserId)
}

func TestEnsureRootOrganization_RejectsNonSelfServiceUser(t *testing.T) {
	srv := newIAMServerForTest(t, newIAMUsecaseMock(t, iamUsecaseMockConfig{}))
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(
			"authorization",
			"Bearer "+rawAccessTokenForIAMUserSource(t, 7, "iam-invite"),
		),
	)

	_, err := srv.EnsureRootOrganization(
		ctx,
		&pbiamv1.EnsureRootOrganizationRequest{Name: "Neo", Slug: "account-7"},
	)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}
