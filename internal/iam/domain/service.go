package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type iamService struct {
	tenants             TenantRepository
	roles               RoleRepository
	policies            PolicyRepository
	groups              GroupRepository
	orgs                OrganizationRepository
	platformMemberships PlatformMembershipRepository
	memberships         MembershipRepository
	invites             InviteRepository
}

var _ IAMUsecase = (*iamService)(nil)

func NewIAMUsecase(
	tenants TenantRepository,
	roles RoleRepository,
	policies PolicyRepository,
	groups GroupRepository,
	orgs OrganizationRepository,
	platformMemberships PlatformMembershipRepository,
	memberships MembershipRepository,
	invites InviteRepository,
) IAMUsecase {
	return &iamService{
		tenants:             tenants,
		roles:               roles,
		policies:            policies,
		groups:              groups,
		orgs:                orgs,
		platformMemberships: platformMemberships,
		memberships:         memberships,
		invites:             invites,
	}
}

func (s *iamService) CreateTenant(ctx context.Context, ownerUserID uint, cmd CreateTenantCmd) (*Tenant, error) {
	if ownerUserID == 0 {
		return nil, ErrInvalidUserID
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ErrInvalidTenantName
	}
	slug := strings.TrimSpace(cmd.Slug)
	if slug == "" {
		return nil, ErrInvalidTenantSlug
	}

	now := time.Now().UTC()
	tenant, err := s.tenants.Create(ctx, Tenant{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, err
	}

	role, err := s.roles.GetByName(ctx, RoleTenantOwner)
	if err != nil {
		return nil, err
	}

	if err := s.memberships.Upsert(ctx, Membership{
		TenantID:  tenant.ID,
		UserID:    ownerUserID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *iamService) AssumeRole(ctx context.Context, input AssumeRoleInput) (*AssumedRole, error) {
	if input.UserID == 0 {
		return nil, ErrInvalidUserID
	}
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(input.RoleName))
	if err != nil {
		return nil, err
	}
	tenantID := strings.TrimSpace(input.TenantID)
	if role.Scope == PolicyScopeTenant && tenantID == "" {
		return nil, ErrInvalidAssumeRole
	}
	if role.Scope == PolicyScopePlatform && tenantID != "" {
		return nil, ErrInvalidAssumeRole
	}
	if input.ServicePrincipal != "" && !validServicePrincipal(input.ServicePrincipal) {
		return nil, ErrInvalidServicePrincipal
	}
	trustStatements, err := s.roles.GetTrustPolicy(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	if !s.canAssumeRole(ctx, input.UserID, tenantID, strings.TrimSpace(input.ExternalID), strings.TrimSpace(input.ServicePrincipal), trustStatements) {
		return nil, ErrAssumeRoleDenied
	}
	now := time.Now().UTC()
	duration := time.Duration(input.DurationSeconds) * time.Second
	if duration <= 0 {
		duration = time.Hour
	}
	if duration > 12*time.Hour {
		duration = 12 * time.Hour
	}
	return &AssumedRole{
		RoleID:           role.ID,
		RoleScope:        role.Scope,
		RoleName:         role.Name,
		TenantID:         tenantID,
		ServicePrincipal: strings.TrimSpace(input.ServicePrincipal),
		SessionName:      strings.TrimSpace(input.SessionName),
		SourceIdentity:   strings.TrimSpace(input.SourceIdentity),
		SessionTags:      cloneStringMap(input.SessionTags),
		ExpiresAt:        now.Add(duration),
		CreatedAt:        now,
	}, nil
}


func normalizePolicyConditions(conditions []PolicyCondition) []PolicyCondition {
	if len(conditions) == 0 {
		return nil
	}
	normalized := make([]PolicyCondition, 0, len(conditions))
	for _, condition := range conditions {
		operator := strings.TrimSpace(condition.Operator)
		key := strings.TrimSpace(condition.Key)
		value := strings.TrimSpace(condition.Value)
		if operator == "" || key == "" {
			continue
		}
		normalized = append(normalized, PolicyCondition{
			Operator: operator,
			Key:      key,
			Value:    value,
		})
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizePolicyStatement(statement PolicyStatement, createdAt time.Time) (PolicyStatement, error) {
	effect := strings.ToLower(strings.TrimSpace(statement.Effect))
	if effect == "" {
		effect = PolicyEffectAllow
	}
	if effect != PolicyEffectAllow && effect != PolicyEffectDeny {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	action := strings.TrimSpace(statement.ActionPattern)
	if action == "" || len(action) > 128 || !strings.Contains(action, ":") {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	resource := strings.TrimSpace(statement.ResourcePattern)
	if resource == "" {
		resource = "*"
	}
	if resource != "*" && len(resource) > 256 {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	conditions := normalizePolicyConditions(statement.Conditions)
	if len(conditions) > 16 {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	return PolicyStatement{
		Effect:          effect,
		ActionPattern:   action,
		ResourcePattern: resource,
		Conditions:      conditions,
		CreatedAt:       createdAt,
	}, nil
}

func cloneStringMap(items map[string]string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(items))
	for k, v := range items {
		cloned[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return cloned
}

func validServicePrincipal(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && strings.Contains(value, ".")
}

func requestAttributesFromContext(ctx context.Context) map[string]string {
	tags := GetSessionTags(ctx)
	if len(tags) == 0 {
		return nil
	}
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		out["principal_tag:"+k] = v
		out["request_tag:"+k] = v
	}
	return out
}
