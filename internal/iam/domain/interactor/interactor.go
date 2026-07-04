package interactor

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/inputport"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type interactor struct {
	tenantCommands             outputport.TenantCommandRepository
	tenantQueries              outputport.TenantQueryRepository
	roleCommands               outputport.RoleCommandRepository
	roleQueries                outputport.RoleQueryRepository
	policyCommands             outputport.PolicyCommandRepository
	policyQueries              outputport.PolicyQueryRepository
	groupCommands              outputport.GroupCommandRepository
	groupQueries               outputport.GroupQueryRepository
	orgCommands                outputport.OrganizationCommandRepository
	orgQueries                 outputport.OrganizationQueryRepository
	platformMembershipCommands outputport.PlatformMembershipCommandRepository
	platformMembershipQueries  outputport.PlatformMembershipQueryRepository
	membershipCommands         outputport.MembershipCommandRepository
	membershipQueries          outputport.MembershipQueryRepository
	inviteCommands             outputport.InviteCommandRepository
	inviteQueries              outputport.InviteQueryRepository
	userDirectory              outputport.UserDirectory
	outbox                     outputport.OutboxRepository
}

var (
	_ inputport.IAMUsecase        = (*interactor)(nil)
	_ inputport.IAMCommandUsecase = (*interactor)(nil)
	_ inputport.IAMQueryUsecase   = (*interactor)(nil)
)

func NewInteractor(
	tenantCommands outputport.TenantCommandRepository,
	tenantQueries outputport.TenantQueryRepository,
	roleCommands outputport.RoleCommandRepository,
	roleQueries outputport.RoleQueryRepository,
	policyCommands outputport.PolicyCommandRepository,
	policyQueries outputport.PolicyQueryRepository,
	groupCommands outputport.GroupCommandRepository,
	groupQueries outputport.GroupQueryRepository,
	orgCommands outputport.OrganizationCommandRepository,
	orgQueries outputport.OrganizationQueryRepository,
	platformMembershipCommands outputport.PlatformMembershipCommandRepository,
	platformMembershipQueries outputport.PlatformMembershipQueryRepository,
	membershipCommands outputport.MembershipCommandRepository,
	membershipQueries outputport.MembershipQueryRepository,
	inviteCommands outputport.InviteCommandRepository,
	inviteQueries outputport.InviteQueryRepository,
	outbox outputport.OutboxRepository,
	userDirectory outputport.UserDirectory,
) *interactor {
	return &interactor{
		tenantCommands:             tenantCommands,
		tenantQueries:              tenantQueries,
		roleCommands:               roleCommands,
		roleQueries:                roleQueries,
		policyCommands:             policyCommands,
		policyQueries:              policyQueries,
		groupCommands:              groupCommands,
		groupQueries:               groupQueries,
		orgCommands:                orgCommands,
		orgQueries:                 orgQueries,
		platformMembershipCommands: platformMembershipCommands,
		platformMembershipQueries:  platformMembershipQueries,
		membershipCommands:         membershipCommands,
		membershipQueries:          membershipQueries,
		inviteCommands:             inviteCommands,
		inviteQueries:              inviteQueries,
		userDirectory:              userDirectory,
		outbox:                     outbox,
	}
}

func NewCommandInteractor(
	tenantCommands outputport.TenantCommandRepository,
	tenantQueries outputport.TenantQueryRepository,
	roleCommands outputport.RoleCommandRepository,
	roleQueries outputport.RoleQueryRepository,
	policyCommands outputport.PolicyCommandRepository,
	policyQueries outputport.PolicyQueryRepository,
	groupCommands outputport.GroupCommandRepository,
	groupQueries outputport.GroupQueryRepository,
	orgCommands outputport.OrganizationCommandRepository,
	orgQueries outputport.OrganizationQueryRepository,
	platformMembershipCommands outputport.PlatformMembershipCommandRepository,
	platformMembershipQueries outputport.PlatformMembershipQueryRepository,
	membershipCommands outputport.MembershipCommandRepository,
	membershipQueries outputport.MembershipQueryRepository,
	inviteCommands outputport.InviteCommandRepository,
	inviteQueries outputport.InviteQueryRepository,
	outbox outputport.OutboxRepository,
) inputport.IAMCommandUsecase {
	return NewInteractor(
		tenantCommands,
		tenantQueries,
		roleCommands,
		roleQueries,
		policyCommands,
		policyQueries,
		groupCommands,
		groupQueries,
		orgCommands,
		orgQueries,
		platformMembershipCommands,
		platformMembershipQueries,
		membershipCommands,
		membershipQueries,
		inviteCommands,
		inviteQueries,
		outbox,
		nil,
	)
}

func NewQueryInteractor(
	tenantQueries outputport.TenantQueryRepository,
	roleQueries outputport.RoleQueryRepository,
	policyQueries outputport.PolicyQueryRepository,
	groupQueries outputport.GroupQueryRepository,
	orgQueries outputport.OrganizationQueryRepository,
	platformMembershipQueries outputport.PlatformMembershipQueryRepository,
	membershipQueries outputport.MembershipQueryRepository,
	inviteQueries outputport.InviteQueryRepository,
	userDirectory outputport.UserDirectory,
) inputport.IAMQueryUsecase {
	return &interactor{
		tenantQueries:             tenantQueries,
		roleQueries:               roleQueries,
		policyQueries:             policyQueries,
		groupQueries:              groupQueries,
		orgQueries:                orgQueries,
		platformMembershipQueries: platformMembershipQueries,
		membershipQueries:         membershipQueries,
		inviteQueries:             inviteQueries,
		userDirectory:             userDirectory,
	}
}

func NewIAMUsecase(
	tenants outputport.TenantRepository,
	roles outputport.RoleRepository,
	policies outputport.PolicyRepository,
	groups outputport.GroupRepository,
	orgs outputport.OrganizationRepository,
	platformMemberships outputport.PlatformMembershipRepository,
	memberships outputport.MembershipRepository,
	invites outputport.InviteRepository,
	outbox outputport.OutboxRepository,
	userDirectory outputport.UserDirectory,
) inputport.IAMUsecase {
	return NewInteractor(
		tenants,
		tenants,
		roles,
		roles,
		policies,
		policies,
		groups,
		groups,
		orgs,
		orgs,
		platformMemberships,
		platformMemberships,
		memberships,
		memberships,
		invites,
		invites,
		outbox,
		userDirectory,
	)
}

func (s *interactor) CreateTenant(
	ctx context.Context,
	ownerUserID uint,
	cmd entity.CreateTenantCmd,
) (*entity.Tenant, error) {
	if ownerUserID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, entity.ErrInvalidTenantName
	}
	slug := strings.TrimSpace(cmd.Slug)
	if slug == "" {
		return nil, entity.ErrInvalidTenantSlug
	}

	now := time.Now().UTC()
	orgID := ""
	rootOrg, orgErr := s.orgQueries.GetByRootUserID(ctx, ownerUserID)
	switch {
	case orgErr == nil:
		orgID = rootOrg.ID
	case !errors.Is(orgErr, entity.ErrOrganizationNotFound):
		return nil, orgErr
	}
	tenant, err := s.tenantCommands.Create(ctx, entity.Tenant{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		OrgID:     orgID,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, err
	}

	role, err := s.roleQueries.GetByName(ctx, entity.RoleTenantOwner)
	if err != nil {
		return nil, err
	}

	if err := s.membershipCommands.Upsert(ctx, entity.Membership{
		TenantID:  tenant.ID,
		UserID:    ownerUserID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    entity.MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return nil, err
	}

	if s.outbox != nil {
		record, err := newTenantCreatedOutboxRecord(*tenant, ownerUserID, now)
		if err != nil {
			return nil, err
		}
		if err := s.appendOutboxRecord(ctx, now, record); err != nil {
			return nil, err
		}
	}

	return tenant, nil
}

func newTenantCreatedOutboxRecord(
	tenant entity.Tenant,
	ownerUserID uint,
	now time.Time,
) (messaging.OutboxRecord, error) {
	return newIAMEventOutboxRecord(now, "tenant.created", tenant.ID, tenant.ID, tenant.ID, map[string]any{
		"tenant_id":     tenant.ID,
		"tenant_slug":   tenant.Slug,
		"tenant_name":   tenant.Name,
		"owner_user_id": ownerUserID,
	})
}

func (s *interactor) AssumeRole(ctx context.Context, input entity.AssumeRoleInput) (*entity.AssumedRole, error) {
	if input.UserID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	role, err := s.roleQueries.GetByName(ctx, strings.TrimSpace(input.RoleName))
	if err != nil {
		return nil, err
	}
	tenantID := strings.TrimSpace(input.TenantID)
	if role.Scope == entity.PolicyScopeTenant && tenantID == "" {
		return nil, entity.ErrInvalidAssumeRole
	}
	if role.Scope == entity.PolicyScopePlatform && tenantID != "" {
		return nil, entity.ErrInvalidAssumeRole
	}
	if input.ServicePrincipal != "" && !validServicePrincipal(input.ServicePrincipal) {
		return nil, entity.ErrInvalidServicePrincipal
	}
	trustStatements, err := s.roleQueries.GetTrustPolicy(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	if !s.canAssumeRole(
		ctx,
		input.UserID,
		tenantID,
		strings.TrimSpace(input.ExternalID),
		strings.TrimSpace(input.ServicePrincipal),
		trustStatements,
	) {
		return nil, entity.ErrAssumeRoleDenied
	}
	now := time.Now().UTC()
	duration := time.Duration(input.DurationSeconds) * time.Second
	if duration <= 0 {
		duration = time.Hour
	}
	if duration > 12*time.Hour {
		duration = 12 * time.Hour
	}
	return &entity.AssumedRole{
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

func normalizePolicyConditions(conditions []entity.PolicyCondition) []entity.PolicyCondition {
	if len(conditions) == 0 {
		return nil
	}
	normalized := make([]entity.PolicyCondition, 0, len(conditions))
	for _, condition := range conditions {
		operator := strings.TrimSpace(condition.Operator)
		key := strings.TrimSpace(condition.Key)
		value := strings.TrimSpace(condition.Value)
		if operator == "" || key == "" {
			continue
		}
		normalized = append(normalized, entity.PolicyCondition{
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

func normalizePolicyStatement(statement entity.PolicyStatement, createdAt time.Time) (entity.PolicyStatement, error) {
	effect := strings.ToLower(strings.TrimSpace(statement.Effect))
	if effect == "" {
		effect = entity.PolicyEffectAllow
	}
	if effect != entity.PolicyEffectAllow && effect != entity.PolicyEffectDeny {
		return entity.PolicyStatement{}, entity.ErrInvalidPolicyStatement
	}
	action := strings.TrimSpace(statement.ActionPattern)
	if action == "" || len(action) > 128 || !strings.Contains(action, ":") {
		return entity.PolicyStatement{}, entity.ErrInvalidPolicyStatement
	}
	resource := strings.TrimSpace(statement.ResourcePattern)
	if resource == "" {
		resource = "*"
	}
	if resource != "*" && len(resource) > 256 {
		return entity.PolicyStatement{}, entity.ErrInvalidPolicyStatement
	}
	conditions := normalizePolicyConditions(statement.Conditions)
	if len(conditions) > 16 {
		return entity.PolicyStatement{}, entity.ErrInvalidPolicyStatement
	}
	return entity.PolicyStatement{
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
	tags := entity.GetSessionTags(ctx)
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
