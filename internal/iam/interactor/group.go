package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/tuannm99/podzone/internal/iam/entity"
)

func (s *interactor) CreateGroup(ctx context.Context, input CreateGroupInput) (*Group, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = PolicyScopeTenant
	}
	now := time.Now().UTC()
	group := Group{
		Scope:       scope,
		TenantID:    strings.TrimSpace(input.TenantID),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.groups.CreateGroup(ctx, group)
}

func (s *interactor) ListGroups(ctx context.Context, scope string, tenantID string) ([]Group, error) {
	return s.groups.ListGroups(ctx, strings.TrimSpace(scope), strings.TrimSpace(tenantID))
}

func (s *interactor) DeleteGroup(ctx context.Context, groupID uint64) error {
	if groupID == 0 {
		return ErrGroupNotFound
	}
	return s.groups.DeleteGroup(ctx, groupID)
}

func (s *interactor) PutGroupInlinePolicy(ctx context.Context, input PutGroupInlinePolicyInput) error {
	if input.GroupID == 0 {
		return ErrGroupNotFound
	}
	if strings.TrimSpace(input.Name) == "" {
		return ErrInvalidPolicyName
	}
	if len(input.Statements) == 0 {
		return fmt.Errorf("iam: at least one policy statement is required")
	}
	now := time.Now().UTC()
	statements := make([]PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return err
		}
		statements = append(statements, normalized)
	}
	return s.groups.PutInlinePolicy(ctx, PutGroupInlinePolicyInput{
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
) (*GroupInlinePolicy, error) {
	if groupID == 0 {
		return nil, ErrGroupNotFound
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidPolicyName
	}
	return s.groups.GetInlinePolicy(ctx, groupID, strings.TrimSpace(name))
}

func (s *interactor) ListGroupInlinePolicies(ctx context.Context, groupID uint64) ([]GroupInlinePolicy, error) {
	if groupID == 0 {
		return nil, ErrGroupNotFound
	}
	return s.groups.ListInlinePolicies(ctx, groupID)
}

func (s *interactor) DeleteGroupInlinePolicy(ctx context.Context, groupID uint64, name string) error {
	if groupID == 0 {
		return ErrGroupNotFound
	}
	if strings.TrimSpace(name) == "" {
		return ErrInvalidPolicyName
	}
	return s.groups.DeleteInlinePolicy(ctx, groupID, strings.TrimSpace(name))
}

func (s *interactor) AddGroupMember(ctx context.Context, groupID uint64, userID uint) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.groups.AddMember(ctx, groupID, userID)
}

func (s *interactor) RemoveGroupMember(ctx context.Context, groupID uint64, userID uint) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.groups.RemoveMember(ctx, groupID, userID)
}

func (s *interactor) ListGroupMembers(ctx context.Context, groupID uint64) ([]uint, error) {
	if groupID == 0 {
		return nil, ErrRoleNotFound
	}
	return s.groups.ListMembers(ctx, groupID)
}

func (s *interactor) AttachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.groups.AttachPolicy(ctx, groupID, policy.ID)
}

func (s *interactor) DetachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.groups.DetachPolicy(ctx, groupID, policy.ID)
}

func (s *interactor) ListGroupPolicies(ctx context.Context, groupID uint64) ([]Policy, error) {
	if groupID == 0 {
		return nil, ErrRoleNotFound
	}
	return s.groups.ListPolicies(ctx, groupID)
}
