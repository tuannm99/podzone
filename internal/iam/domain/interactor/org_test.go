package interactor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	iaminteractor "github.com/tuannm99/podzone/internal/iam/domain/interactor"
	outputportmocks "github.com/tuannm99/podzone/internal/iam/domain/outputport/mocks"
)

func TestCheckOrganizationPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		userID          uint
		isRoot          bool
		membership      *entity.OrganizationMembership
		roleAllows      bool
		groupStatements []entity.PolicyStatement
		want            bool
	}{
		{
			name:   "root has implicit permission",
			userID: 7,
			isRoot: true,
			want:   true,
		},
		{
			name:   "organization admin uses role permission",
			userID: 8,
			membership: &entity.OrganizationMembership{
				OrgID:    "org-1",
				UserID:   8,
				RoleID:   12,
				RoleName: entity.RoleOrganizationAdmin,
				Status:   entity.MembershipStatusActive,
			},
			roleAllows: true,
			want:       true,
		},
		{
			name:   "organization viewer cannot manage members",
			userID: 9,
			membership: &entity.OrganizationMembership{
				OrgID:    "org-1",
				UserID:   9,
				RoleID:   13,
				RoleName: entity.RoleOrganizationViewer,
				Status:   entity.MembershipStatusActive,
			},
			want: false,
		},
		{
			name:   "organization group explicit deny overrides role",
			userID: 10,
			membership: &entity.OrganizationMembership{
				OrgID:    "org-1",
				UserID:   10,
				RoleID:   12,
				RoleName: entity.RoleOrganizationAdmin,
				Status:   entity.MembershipStatusActive,
			},
			roleAllows: true,
			groupStatements: []entity.PolicyStatement{{
				PolicyName:      "deny-member-management",
				Effect:          entity.PolicyEffectDeny,
				ActionPattern:   "organization:manage_members",
				ResourcePattern: "podzone:organization/org-1",
			}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			organizations := outputportmocks.NewMockOrganizationQueryRepository(t)
			roles := outputportmocks.NewMockRoleQueryRepository(t)
			policies := outputportmocks.NewMockPolicyQueryRepository(t)
			organizations.EXPECT().
				IsRoot(mock.Anything, "org-1", tt.userID).
				Return(tt.isRoot, nil)
			if !tt.isRoot {
				organizations.EXPECT().
					GetMembership(mock.Anything, "org-1", tt.userID).
					Return(tt.membership, nil)
				roles.EXPECT().
					RoleHasPermission(mock.Anything, tt.membership.RoleID, "organization:manage_members").
					Return(tt.roleAllows, nil)
				policies.EXPECT().
					ListOrganizationGroupStatements(mock.Anything, "org-1", tt.userID).
					Return(tt.groupStatements, nil)
			}
			usecase := iaminteractor.NewQueryInteractor(
				nil,
				roles,
				policies,
				nil,
				organizations,
				nil,
				nil,
				nil,
			)

			allowed, err := usecase.CheckOrganizationPermission(
				context.Background(),
				"org-1",
				tt.userID,
				"organization:manage_members",
			)

			require.NoError(t, err)
			require.Equal(t, tt.want, allowed)
		})
	}
}
