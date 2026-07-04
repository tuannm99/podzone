package interactor

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

func (s *interactor) CreateOrganization(ctx context.Context, name string, slug string) (*entity.Organization, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, entity.ErrInvalidOrganizationName
	}
	if slug == "" {
		return nil, entity.ErrInvalidOrganizationSlug
	}
	now := time.Now().UTC()
	return s.orgCommands.Create(ctx, entity.Organization{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *interactor) EnsureRootOrganization(
	ctx context.Context,
	rootUserID uint,
	name string,
	slug string,
) (*entity.Organization, error) {
	if rootUserID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, entity.ErrInvalidOrganizationName
	}
	if slug == "" {
		return nil, entity.ErrInvalidOrganizationSlug
	}
	now := time.Now().UTC()
	return s.orgCommands.EnsureRoot(ctx, entity.Organization{
		ID:         uuid.NewString(),
		Name:       name,
		Slug:       slug,
		RootUserID: rootUserID,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
}

func (s *interactor) ListOrganizations(
	ctx context.Context,
	query collection.Query,
) (collection.Page[entity.Organization], error) {
	return s.orgQueries.List(ctx, query.Normalize())
}

func (s *interactor) ListOrganizationsForUser(
	ctx context.Context,
	userID uint,
	query collection.Query,
) (collection.Page[entity.Organization], error) {
	if userID == 0 {
		return collection.Page[entity.Organization]{}, entity.ErrInvalidUserID
	}
	return s.orgQueries.ListByUserID(ctx, userID, query.Normalize())
}

func (s *interactor) IsOrganizationRoot(ctx context.Context, orgID string, userID uint) (bool, error) {
	if userID == 0 {
		return false, entity.ErrInvalidUserID
	}
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return false, entity.ErrOrganizationNotFound
	}
	return s.orgQueries.IsRoot(ctx, orgID, userID)
}

func (s *interactor) AddOrganizationMember(
	ctx context.Context,
	orgID string,
	userID uint,
	roleName string,
) error {
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return entity.ErrOrganizationNotFound
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return entity.ErrInvalidRoleName
	}
	if roleName == entity.RoleOrganizationRoot {
		return entity.ErrImmutableOrganizationRoot
	}
	if _, err := s.orgQueries.GetByID(ctx, orgID); err != nil {
		return err
	}
	role, err := s.roleQueries.GetByName(ctx, roleName)
	if err != nil {
		return err
	}
	if role.Scope != entity.PolicyScopeOrganization {
		return entity.ErrInvalidRoleName
	}
	now := time.Now().UTC()
	return s.orgCommands.UpsertMembership(ctx, entity.OrganizationMembership{
		OrgID:     orgID,
		UserID:    userID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    entity.MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *interactor) RemoveOrganizationMember(ctx context.Context, orgID string, userID uint) error {
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	org, err := s.orgQueries.GetByID(ctx, strings.TrimSpace(orgID))
	if err != nil {
		return err
	}
	if org.RootUserID == userID {
		return entity.ErrImmutableOrganizationRoot
	}
	return s.orgCommands.DeleteMembership(ctx, org.ID, userID)
}

func (s *interactor) CheckOrganizationPermission(
	ctx context.Context,
	orgID string,
	userID uint,
	permission string,
) (bool, error) {
	if userID == 0 {
		return false, entity.ErrInvalidUserID
	}
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return false, entity.ErrOrganizationNotFound
	}
	isRoot, err := s.orgQueries.IsRoot(ctx, orgID, userID)
	if err != nil {
		return false, err
	}
	if isRoot {
		return true, nil
	}
	membership, err := s.orgQueries.GetMembership(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, entity.ErrOrganizationMembershipNotFound) {
			return false, nil
		}
		return false, err
	}
	if membership.Status != entity.MembershipStatusActive {
		return false, entity.ErrInactiveMembership
	}
	permission = strings.TrimSpace(permission)
	roleAllowed, err := s.roleQueries.RoleHasPermission(ctx, membership.RoleID, permission)
	if err != nil {
		return false, err
	}
	groupStatements, err := s.policyQueries.ListOrganizationGroupStatements(ctx, orgID, userID)
	if err != nil {
		return false, err
	}
	groupDecision := explainPolicyStatements("organization_group", entity.AccessRequest{
		OrgID:      orgID,
		UserID:     userID,
		Action:     permission,
		Resource:   "podzone:organization/" + orgID,
		Attributes: map[string]string{"org_id": orgID},
	}, groupStatements)
	if groupDecision.Reason == "explicit deny matched" {
		return false, nil
	}
	return roleAllowed || groupDecision.Allowed, nil
}

func (s *interactor) RequireOrganizationPermission(
	ctx context.Context,
	orgID string,
	userID uint,
	permission string,
) error {
	allowed, err := s.CheckOrganizationPermission(ctx, orgID, userID, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return entity.NewPermissionDeniedError(permission, "podzone:organization/"+orgID)
	}
	return nil
}

func (s *interactor) ListOrganizationMembers(
	ctx context.Context,
	orgID string,
	query collection.Query,
) (collection.Page[entity.OrganizationMembership], error) {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return collection.Page[entity.OrganizationMembership]{}, entity.ErrOrganizationNotFound
	}
	return s.orgQueries.ListMemberships(ctx, orgID, query.Normalize())
}

func (s *interactor) AttachTenantToOrganization(ctx context.Context, tenantID string, orgID string) error {
	tenantID = strings.TrimSpace(tenantID)
	orgID = strings.TrimSpace(orgID)
	if tenantID == "" {
		return entity.ErrTenantNotFound
	}
	if orgID == "" {
		return entity.ErrOrganizationNotFound
	}
	if _, err := s.tenantQueries.GetByID(ctx, tenantID); err != nil {
		return err
	}
	if _, err := s.orgQueries.GetByID(ctx, orgID); err != nil {
		return err
	}
	return s.tenantCommands.AttachOrganization(ctx, tenantID, orgID)
}

func (s *interactor) DetachTenantFromOrganization(ctx context.Context, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return entity.ErrTenantNotFound
	}
	return s.tenantCommands.DetachOrganization(ctx, tenantID)
}

func (s *interactor) AttachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return entity.ErrOrganizationNotFound
	}
	policy, err := s.policyQueries.GetPolicy(ctx, entity.PolicyRef{
		Scope: entity.PolicyScopePlatform,
		Name:  strings.TrimSpace(policyName),
	})
	if err != nil {
		return err
	}
	return s.orgCommands.AttachServiceControlPolicy(ctx, orgID, policy.ID)
}

func (s *interactor) DetachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return entity.ErrOrganizationNotFound
	}
	policy, err := s.policyQueries.GetPolicy(ctx, entity.PolicyRef{
		Scope: entity.PolicyScopePlatform,
		Name:  strings.TrimSpace(policyName),
	})
	if err != nil {
		return err
	}
	return s.orgCommands.DetachServiceControlPolicy(ctx, orgID, policy.ID)
}

func (s *interactor) ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error) {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return nil, entity.ErrOrganizationNotFound
	}
	return s.orgQueries.ListServiceControlPolicies(ctx, orgID)
}
