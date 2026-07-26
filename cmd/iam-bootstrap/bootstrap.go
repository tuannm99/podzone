package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
)

// validRoles matches PZEP-0002's Proposed Solution: only these two platform
// roles may be granted through this tool. Any other value is rejected
// before touching the database.
var validRoles = map[string]bool{
	"platform_owner": true,
	"platform_admin": true,
}

// errExistingAdmins is returned when the target role already has at least
// one active grant and --force was not set - the guardrail that separates
// the fast bootstrap path (no admins yet) from break-glass recovery
// (admins exist but are unreachable).
var errExistingAdmins = errors.New("iam-bootstrap: existing platform admin(s) present, rerun with --force")

type bootstrapOptions struct {
	UserID uint
	Role   string
	Reason string
	Force  bool
}

// bootstrapDeps are the exact repository/client interfaces this tool needs -
// deliberately narrower than the full IAM command usecase, so wiring stays
// minimal and this logic is unit-testable with mocks, no database required.
type bootstrapDeps struct {
	Roles              outputport.RoleQueryRepository
	MembershipCommands outputport.PlatformMembershipCommandRepository
	MembershipQueries  outputport.PlatformMembershipQueryRepository
	Audit              outputport.AuditLogRepository
	Users              outputport.UserDirectory
}

func validateFlags(opts bootstrapOptions) error {
	if !validRoles[opts.Role] {
		return fmt.Errorf("iam-bootstrap: invalid --role %q: must be platform_owner or platform_admin", opts.Role)
	}
	if opts.UserID == 0 {
		return errors.New("iam-bootstrap: --user-id is required")
	}
	if strings.TrimSpace(opts.Reason) == "" {
		return errors.New("iam-bootstrap: --reason is required")
	}
	return nil
}

// runBootstrap implements PZEP-0002's Proposed Solution steps 1-5. Returns
// the created audit log's ID on success.
func runBootstrap(ctx context.Context, deps bootstrapDeps, opts bootstrapOptions) (string, error) {
	if err := validateFlags(opts); err != nil {
		return "", err
	}

	if _, err := deps.Users.GetByID(ctx, opts.UserID); err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return "", fmt.Errorf("iam-bootstrap: user %d not found in auth", opts.UserID)
		}
		return "", fmt.Errorf("iam-bootstrap: look up user %d in auth: %w", opts.UserID, err)
	}

	role, err := deps.Roles.GetByName(ctx, opts.Role)
	if err != nil {
		return "", fmt.Errorf("iam-bootstrap: role %q: %w", opts.Role, err)
	}

	existingCount, err := deps.MembershipQueries.CountByRole(ctx, role.ID)
	if err != nil {
		return "", fmt.Errorf("iam-bootstrap: count existing %s admins: %w", opts.Role, err)
	}
	if existingCount > 0 && !opts.Force {
		return "", fmt.Errorf("%w (%d existing %s grant(s))", errExistingAdmins, existingCount, opts.Role)
	}

	if err := deps.MembershipCommands.Upsert(ctx, opts.UserID, role.ID, entity.MembershipStatusActive); err != nil {
		return "", fmt.Errorf("iam-bootstrap: grant role: %w", err)
	}

	payloadJSON, err := json.Marshal(map[string]any{
		"reason":               opts.Reason,
		"force":                opts.Force,
		"existing_admin_count": existingCount,
	})
	if err != nil {
		payloadJSON = []byte("{}")
	}
	auditLog := entity.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  0, // sentinel: no real user is the actor - see PZEP-0002
		Action:       "platform.bootstrap.role_granted",
		ResourceType: "platform_role_membership",
		ResourceID:   fmt.Sprintf("%s:%d", opts.Role, opts.UserID),
		Status:       "success",
		PayloadJSON:  string(payloadJSON),
		CreatedAt:    time.Now().UTC(),
	}
	if err := deps.Audit.Create(ctx, auditLog); err != nil {
		return "", fmt.Errorf("iam-bootstrap: write audit log: %w", err)
	}
	return auditLog.ID, nil
}
