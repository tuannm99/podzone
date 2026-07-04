package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

func (s *interactor) CreateGroup(ctx context.Context, input entity.CreateGroupInput) (*entity.Group, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = entity.PolicyScopeTenant
	}
	now := time.Now().UTC()
	group := entity.Group{
		Scope:       scope,
		OrgID:       strings.TrimSpace(input.OrgID),
		TenantID:    strings.TrimSpace(input.TenantID),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := validateGroupOwner(group); err != nil {
		return nil, err
	}
	return s.groupCommands.CreateGroup(ctx, group)
}

func (s *interactor) ListGroups(
	ctx context.Context,
	scope string,
	orgID string,
	tenantID string,
	query collection.Query,
) (collection.Page[entity.Group], error) {
	owner := entity.Group{
		Scope:    strings.TrimSpace(scope),
		OrgID:    strings.TrimSpace(orgID),
		TenantID: strings.TrimSpace(tenantID),
	}
	if err := validateGroupOwner(owner); err != nil {
		return collection.Page[entity.Group]{}, err
	}
	return s.groupQueries.ListGroups(
		ctx,
		owner.Scope,
		owner.OrgID,
		owner.TenantID,
		query.Normalize(),
	)
}

func (s *interactor) GetGroup(ctx context.Context, groupID uint64) (*entity.Group, error) {
	if groupID == 0 {
		return nil, entity.ErrGroupNotFound
	}
	return s.groupQueries.GetByID(ctx, groupID)
}

func (s *interactor) DeleteGroup(ctx context.Context, groupID uint64) error {
	if groupID == 0 {
		return entity.ErrGroupNotFound
	}
	return s.groupCommands.DeleteGroup(ctx, groupID)
}

func (s *interactor) PutGroupInlinePolicy(ctx context.Context, input entity.PutGroupInlinePolicyInput) error {
	if input.GroupID == 0 {
		return entity.ErrGroupNotFound
	}
	if strings.TrimSpace(input.Name) == "" {
		return entity.ErrInvalidPolicyName
	}
	if len(input.Statements) == 0 {
		return fmt.Errorf("iam: at least one policy statement is required")
	}
	now := time.Now().UTC()
	statements := make([]entity.PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return err
		}
		statements = append(statements, normalized)
	}
	return s.groupCommands.PutInlinePolicy(ctx, entity.PutGroupInlinePolicyInput{
		GroupID:     input.GroupID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *interactor) GetGroupInlinePolicy(
	ctx context.Context,
	groupID uint64,
	name string,
) (*entity.GroupInlinePolicy, error) {
	if groupID == 0 {
		return nil, entity.ErrGroupNotFound
	}
	if strings.TrimSpace(name) == "" {
		return nil, entity.ErrInvalidPolicyName
	}
	return s.groupQueries.GetInlinePolicy(ctx, groupID, strings.TrimSpace(name))
}

func (s *interactor) ListGroupInlinePolicies(
	ctx context.Context,
	groupID uint64,
	query collection.Query,
) (collection.Page[entity.GroupInlinePolicy], error) {
	if groupID == 0 {
		return collection.Page[entity.GroupInlinePolicy]{}, entity.ErrGroupNotFound
	}
	return s.groupQueries.ListInlinePolicies(ctx, groupID, query.Normalize())
}

func (s *interactor) DeleteGroupInlinePolicy(ctx context.Context, groupID uint64, name string) error {
	if groupID == 0 {
		return entity.ErrGroupNotFound
	}
	if strings.TrimSpace(name) == "" {
		return entity.ErrInvalidPolicyName
	}
	return s.groupCommands.DeleteInlinePolicy(ctx, groupID, strings.TrimSpace(name))
}

func (s *interactor) AddGroupMember(ctx context.Context, groupID uint64, userID uint) error {
	if groupID == 0 {
		return entity.ErrRoleNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	group, err := s.groupQueries.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.Scope == entity.PolicyScopeOrganization {
		membership, membershipErr := s.orgQueries.GetMembership(ctx, group.OrgID, userID)
		if membershipErr != nil {
			return membershipErr
		}
		if membership.Status != entity.MembershipStatusActive {
			return entity.ErrInactiveMembership
		}
	}
	return s.groupCommands.AddMember(ctx, groupID, userID)
}

func (s *interactor) RemoveGroupMember(ctx context.Context, groupID uint64, userID uint) error {
	if groupID == 0 {
		return entity.ErrRoleNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	return s.groupCommands.RemoveMember(ctx, groupID, userID)
}

func (s *interactor) ListGroupMembers(
	ctx context.Context,
	groupID uint64,
	query collection.Query,
) (collection.Page[uint], error) {
	if groupID == 0 {
		return collection.Page[uint]{}, entity.ErrRoleNotFound
	}
	return s.groupQueries.ListMembers(ctx, groupID, query.Normalize())
}

func (s *interactor) AttachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error {
	if groupID == 0 {
		return entity.ErrRoleNotFound
	}
	group, err := s.groupQueries.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, entity.PolicyRef{
		Scope: group.Scope,
		OrgID: group.OrgID,
		Name:  strings.TrimSpace(policyName),
	})
	if err != nil {
		return err
	}
	if err := s.groupCommands.AttachPolicy(ctx, groupID, policy.ID); err != nil {
		return err
	}
	now := time.Now().UTC()
	record, err := newIAMEventOutboxRecord(
		now,
		"policy.attached",
		group.TenantID,
		group.Name,
		group.Name,
		map[string]any{
			"tenant_id":        group.TenantID,
			"group_id":         group.ID,
			"group_name":       group.Name,
			"policy_id":        policy.ID,
			"policy_name":      policy.Name,
			"policy_scope":     policy.Scope,
			"attachment_type":  "group",
			"attachment_scope": group.Scope,
		},
	)
	if err != nil {
		return err
	}
	return s.appendOutboxRecord(ctx, now, record)
}

func (s *interactor) DetachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error {
	if groupID == 0 {
		return entity.ErrRoleNotFound
	}
	group, err := s.groupQueries.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, entity.PolicyRef{
		Scope: group.Scope,
		OrgID: group.OrgID,
		Name:  strings.TrimSpace(policyName),
	})
	if err != nil {
		return err
	}
	return s.groupCommands.DetachPolicy(ctx, groupID, policy.ID)
}

func validateGroupOwner(group entity.Group) error {
	switch group.Scope {
	case entity.PolicyScopePlatform:
		if group.OrgID != "" || group.TenantID != "" {
			return entity.ErrInvalidPolicyOwner
		}
	case entity.PolicyScopeOrganization:
		if group.OrgID == "" || group.TenantID != "" {
			return entity.ErrInvalidPolicyOwner
		}
	case entity.PolicyScopeTenant:
		if group.OrgID != "" || group.TenantID == "" {
			return entity.ErrInvalidPolicyOwner
		}
	default:
		return entity.ErrInvalidPolicyOwner
	}
	return nil
}

func (s *interactor) ListGroupPolicies(
	ctx context.Context,
	groupID uint64,
	query collection.Query,
) (collection.Page[entity.Policy], error) {
	if groupID == 0 {
		return collection.Page[entity.Policy]{}, entity.ErrRoleNotFound
	}
	return s.groupQueries.ListPolicies(ctx, groupID, query.Normalize())
}
