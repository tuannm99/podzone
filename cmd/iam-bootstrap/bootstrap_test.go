package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport/mocks"
)

func validOptions() bootstrapOptions {
	return bootstrapOptions{UserID: 4, Role: "platform_owner", Reason: "initial platform admin"}
}

func TestValidateFlags(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, validateFlags(validOptions()))
	})
	t.Run("invalid role", func(t *testing.T) {
		opts := validOptions()
		opts.Role = "superuser"
		require.ErrorContains(t, validateFlags(opts), "invalid --role")
	})
	t.Run("empty role", func(t *testing.T) {
		opts := validOptions()
		opts.Role = ""
		require.ErrorContains(t, validateFlags(opts), "invalid --role")
	})
	t.Run("zero user id", func(t *testing.T) {
		opts := validOptions()
		opts.UserID = 0
		require.ErrorContains(t, validateFlags(opts), "--user-id is required")
	})
	t.Run("empty reason", func(t *testing.T) {
		opts := validOptions()
		opts.Reason = "   "
		require.ErrorContains(t, validateFlags(opts), "--reason is required")
	})
}

func TestRunBootstrap_GrantsWhenNoExistingAdmin(t *testing.T) {
	roles := mocks.NewMockRoleQueryRepository(t)
	commands := mocks.NewMockPlatformMembershipCommandRepository(t)
	queries := mocks.NewMockPlatformMembershipQueryRepository(t)
	audit := mocks.NewMockAuditLogRepository(t)
	users := mocks.NewMockUserDirectory(t)

	users.EXPECT().GetByID(mock.Anything, uint(4)).Return(&entity.User{ID: 4}, nil).Once()
	roles.EXPECT().GetByName(mock.Anything, "platform_owner").
		Return(&entity.Role{ID: 5, Name: "platform_owner"}, nil).Once()
	queries.EXPECT().CountByRole(mock.Anything, uint64(5)).Return(0, nil).Once()
	commands.EXPECT().Upsert(mock.Anything, uint(4), uint64(5), entity.MembershipStatusActive).Return(nil).Once()
	var recorded entity.AuditLog
	audit.EXPECT().Create(mock.Anything, mock.Anything).
		Run(func(_ context.Context, log entity.AuditLog) { recorded = log }).
		Return(nil).Once()

	deps := bootstrapDeps{
		Roles:              roles,
		MembershipCommands: commands,
		MembershipQueries:  queries,
		Audit:              audit,
		Users:              users,
	}
	id, err := runBootstrap(context.Background(), deps, validOptions())

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.Equal(t, id, recorded.ID)
	require.Equal(t, uint(0), recorded.ActorUserID)
	require.Equal(t, "platform.bootstrap.role_granted", recorded.Action)
	require.Equal(t, "platform_role_membership", recorded.ResourceType)
	require.Equal(t, "platform_owner:4", recorded.ResourceID)
	require.Equal(t, "success", recorded.Status)
	require.Contains(t, recorded.PayloadJSON, `"reason":"initial platform admin"`)
	require.Contains(t, recorded.PayloadJSON, `"force":false`)
	require.Contains(t, recorded.PayloadJSON, `"existing_admin_count":0`)
}

func TestRunBootstrap_RefusesWhenAdminExistsWithoutForce(t *testing.T) {
	roles := mocks.NewMockRoleQueryRepository(t)
	commands := mocks.NewMockPlatformMembershipCommandRepository(t)
	queries := mocks.NewMockPlatformMembershipQueryRepository(t)
	audit := mocks.NewMockAuditLogRepository(t)
	users := mocks.NewMockUserDirectory(t)

	users.EXPECT().GetByID(mock.Anything, uint(4)).Return(&entity.User{ID: 4}, nil).Once()
	roles.EXPECT().GetByName(mock.Anything, "platform_owner").
		Return(&entity.Role{ID: 5, Name: "platform_owner"}, nil).Once()
	queries.EXPECT().CountByRole(mock.Anything, uint64(5)).Return(1, nil).Once()
	// Upsert/Create must not be called - no .EXPECT() set for them, mockery
	// fails the test if they are.

	deps := bootstrapDeps{
		Roles:              roles,
		MembershipCommands: commands,
		MembershipQueries:  queries,
		Audit:              audit,
		Users:              users,
	}
	_, err := runBootstrap(context.Background(), deps, validOptions())

	require.ErrorIs(t, err, errExistingAdmins)
	require.ErrorContains(t, err, "1 existing platform_owner grant")
}

func TestRunBootstrap_GrantsWithForceWhenAdminExists(t *testing.T) {
	roles := mocks.NewMockRoleQueryRepository(t)
	commands := mocks.NewMockPlatformMembershipCommandRepository(t)
	queries := mocks.NewMockPlatformMembershipQueryRepository(t)
	audit := mocks.NewMockAuditLogRepository(t)
	users := mocks.NewMockUserDirectory(t)

	users.EXPECT().GetByID(mock.Anything, uint(4)).Return(&entity.User{ID: 4}, nil).Once()
	roles.EXPECT().GetByName(mock.Anything, "platform_owner").
		Return(&entity.Role{ID: 5, Name: "platform_owner"}, nil).Once()
	queries.EXPECT().CountByRole(mock.Anything, uint64(5)).Return(2, nil).Once()
	commands.EXPECT().Upsert(mock.Anything, uint(4), uint64(5), entity.MembershipStatusActive).Return(nil).Once()
	var recorded entity.AuditLog
	audit.EXPECT().Create(mock.Anything, mock.Anything).
		Run(func(_ context.Context, log entity.AuditLog) { recorded = log }).
		Return(nil).Once()

	opts := validOptions()
	opts.Force = true
	deps := bootstrapDeps{
		Roles:              roles,
		MembershipCommands: commands,
		MembershipQueries:  queries,
		Audit:              audit,
		Users:              users,
	}
	id, err := runBootstrap(context.Background(), deps, opts)

	require.NoError(t, err)
	require.Equal(t, id, recorded.ID)
	require.Contains(t, recorded.PayloadJSON, `"force":true`)
	require.Contains(t, recorded.PayloadJSON, `"existing_admin_count":2`)
}

func TestRunBootstrap_RunTwiceDoesNotDuplicate(t *testing.T) {
	// Upsert itself is idempotent (ON CONFLICT DO UPDATE, see
	// platform_membership_repo_impl.go) - this test proves runBootstrap
	// calls it the same way on a repeat invocation, not that it inserts a
	// second row (that guarantee lives in the repository/DB layer).
	roles := mocks.NewMockRoleQueryRepository(t)
	commands := mocks.NewMockPlatformMembershipCommandRepository(t)
	queries := mocks.NewMockPlatformMembershipQueryRepository(t)
	audit := mocks.NewMockAuditLogRepository(t)
	users := mocks.NewMockUserDirectory(t)

	users.EXPECT().GetByID(mock.Anything, uint(4)).Return(&entity.User{ID: 4}, nil).Twice()
	roles.EXPECT().GetByName(mock.Anything, "platform_owner").
		Return(&entity.Role{ID: 5, Name: "platform_owner"}, nil).Twice()
	queries.EXPECT().CountByRole(mock.Anything, uint64(5)).Return(0, nil).Once()
	queries.EXPECT().CountByRole(mock.Anything, uint64(5)).Return(1, nil).Once()
	commands.EXPECT().Upsert(mock.Anything, uint(4), uint64(5), entity.MembershipStatusActive).Return(nil).Twice()
	audit.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Twice()

	deps := bootstrapDeps{
		Roles:              roles,
		MembershipCommands: commands,
		MembershipQueries:  queries,
		Audit:              audit,
		Users:              users,
	}
	_, err := runBootstrap(context.Background(), deps, validOptions())
	require.NoError(t, err)

	opts := validOptions()
	opts.Force = true
	_, err = runBootstrap(context.Background(), deps, opts)
	require.NoError(t, err)
}

func TestRunBootstrap_UserNotFound(t *testing.T) {
	roles := mocks.NewMockRoleQueryRepository(t)
	commands := mocks.NewMockPlatformMembershipCommandRepository(t)
	queries := mocks.NewMockPlatformMembershipQueryRepository(t)
	audit := mocks.NewMockAuditLogRepository(t)
	users := mocks.NewMockUserDirectory(t)

	users.EXPECT().GetByID(mock.Anything, uint(4)).Return(nil, entity.ErrUserNotFound).Once()

	deps := bootstrapDeps{
		Roles:              roles,
		MembershipCommands: commands,
		MembershipQueries:  queries,
		Audit:              audit,
		Users:              users,
	}
	_, err := runBootstrap(context.Background(), deps, validOptions())

	require.ErrorContains(t, err, "user 4 not found in auth")
}

func TestRunBootstrap_RoleLookupFailure(t *testing.T) {
	roles := mocks.NewMockRoleQueryRepository(t)
	commands := mocks.NewMockPlatformMembershipCommandRepository(t)
	queries := mocks.NewMockPlatformMembershipQueryRepository(t)
	audit := mocks.NewMockAuditLogRepository(t)
	users := mocks.NewMockUserDirectory(t)

	users.EXPECT().GetByID(mock.Anything, uint(4)).Return(&entity.User{ID: 4}, nil).Once()
	roles.EXPECT().GetByName(mock.Anything, "platform_owner").
		Return(nil, entity.ErrRoleNotFound).Once()

	deps := bootstrapDeps{
		Roles:              roles,
		MembershipCommands: commands,
		MembershipQueries:  queries,
		Audit:              audit,
		Users:              users,
	}
	_, err := runBootstrap(context.Background(), deps, validOptions())

	require.ErrorIs(t, err, entity.ErrRoleNotFound)
}

func TestRunBootstrap_UpsertFailurePropagates(t *testing.T) {
	roles := mocks.NewMockRoleQueryRepository(t)
	commands := mocks.NewMockPlatformMembershipCommandRepository(t)
	queries := mocks.NewMockPlatformMembershipQueryRepository(t)
	audit := mocks.NewMockAuditLogRepository(t)
	users := mocks.NewMockUserDirectory(t)

	boom := errors.New("boom")
	users.EXPECT().GetByID(mock.Anything, uint(4)).Return(&entity.User{ID: 4}, nil).Once()
	roles.EXPECT().GetByName(mock.Anything, "platform_owner").
		Return(&entity.Role{ID: 5, Name: "platform_owner"}, nil).Once()
	queries.EXPECT().CountByRole(mock.Anything, uint64(5)).Return(0, nil).Once()
	commands.EXPECT().Upsert(mock.Anything, uint(4), uint64(5), entity.MembershipStatusActive).Return(boom).Once()

	deps := bootstrapDeps{
		Roles:              roles,
		MembershipCommands: commands,
		MembershipQueries:  queries,
		Audit:              audit,
		Users:              users,
	}
	_, err := runBootstrap(context.Background(), deps, validOptions())

	require.ErrorIs(t, err, boom)
}
