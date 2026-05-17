package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type PolicyRepositoryImpl struct {
	db *sqlx.DB
}

func NewPolicyRepository(p repoParams) iamdomain.PolicyRepository {
	return &PolicyRepositoryImpl{db: p.DB}
}

func (r *PolicyRepositoryImpl) CreatePolicy(
	ctx context.Context,
	policy iamdomain.Policy,
	statements []iamdomain.PolicyStatement,
) (*iamdomain.Policy, []iamdomain.PolicyStatement, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var created policyModel
	if err := tx.GetContext(
		ctx,
		&created,
		`INSERT INTO iam_policies (scope, name, description, is_system, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, scope, name, description, is_system, created_at, updated_at`,
		policy.Scope,
		policy.Name,
		policy.Description,
		policy.IsSystem,
		policy.CreatedAt,
		policy.UpdatedAt,
	); err != nil {
		return nil, nil, err
	}

	outStatements := make([]iamdomain.PolicyStatement, 0, len(statements))
	for _, statement := range statements {
		var createdStatement policyStatementModel
		if err := tx.GetContext(
			ctx,
			&createdStatement,
			`INSERT INTO iam_policy_statements (policy_id, effect, action_pattern, resource_pattern, created_at)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, policy_id, $6 AS policy_name, effect, action_pattern, resource_pattern, created_at`,
			created.ID,
			statement.Effect,
			statement.ActionPattern,
			coalescePattern(statement.ResourcePattern),
			statement.CreatedAt,
			created.Name,
		); err != nil {
			return nil, nil, err
		}
		outStatements = append(outStatements, createdStatement.toEntity())
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	entity := created.toEntity()
	return &entity, outStatements, nil
}

func (r *PolicyRepositoryImpl) GetPolicyByName(ctx context.Context, name string) (*iamdomain.Policy, error) {
	var out policyModel
	if err := r.db.GetContext(
		ctx,
		&out,
		`SELECT id, scope, name, description, is_system, created_at, updated_at
		 FROM iam_policies WHERE name = $1`,
		name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrRoleNotFound
		}
		return nil, err
	}
	entity := out.toEntity()
	return &entity, nil
}

func (r *PolicyRepositoryImpl) GetPolicyStatements(ctx context.Context, policyID uint64) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.created_at
		 FROM iam_policy_statements ps
		 JOIN iam_policies p ON p.id = ps.policy_id
		 WHERE ps.policy_id = $1
		 ORDER BY ps.id ASC`,
		policyID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) ListPolicies(ctx context.Context, scope string) ([]iamdomain.Policy, error) {
	query := `SELECT id, scope, name, description, is_system, created_at, updated_at FROM iam_policies`
	args := []any{}
	if scope != "" {
		query += ` WHERE scope = $1`
		args = append(args, scope)
	}
	query += ` ORDER BY scope ASC, name ASC`

	var rows []policyModel
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]iamdomain.Policy, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

func (r *PolicyRepositoryImpl) ListPolicyAttachments(ctx context.Context, policyID uint64) ([]iamdomain.PolicyAttachment, error) {
	var rows []policyAttachmentModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT attachment_type, scope, tenant_id, role_id, role_name, user_id, group_id, group_name, created_at
		 FROM (
		   SELECT
		     'role' AS attachment_type,
		     '' AS scope,
		     '' AS tenant_id,
		     r.id AS role_id,
		     r.name AS role_name,
		     0::bigint AS user_id,
		     0::bigint AS group_id,
		     '' AS group_name,
		     arpa.created_at AS created_at
		   FROM iam_role_policy_attachments arpa
		   JOIN iam_roles r ON r.id = arpa.role_id
		   WHERE arpa.policy_id = $1
		   UNION ALL
		   SELECT
		     'platform_user' AS attachment_type,
		     upa.scope AS scope,
		     '' AS tenant_id,
		     0::bigint AS role_id,
		     '' AS role_name,
		     upa.user_id AS user_id,
		     0::bigint AS group_id,
		     '' AS group_name,
		     upa.created_at AS created_at
		   FROM iam_user_policy_attachments upa
		   WHERE upa.policy_id = $1
		   UNION ALL
		   SELECT
		     'tenant_user' AS attachment_type,
		     'tenant' AS scope,
		     tupa.tenant_id AS tenant_id,
		     0::bigint AS role_id,
		     '' AS role_name,
		     tupa.user_id AS user_id,
		     0::bigint AS group_id,
		     '' AS group_name,
		     tupa.created_at AS created_at
		   FROM iam_tenant_user_policy_attachments tupa
		   WHERE tupa.policy_id = $1
		   UNION ALL
		   SELECT
		     'group' AS attachment_type,
		     g.scope AS scope,
		     COALESCE(g.tenant_id, '') AS tenant_id,
		     0::bigint AS role_id,
		     '' AS role_name,
		     0::bigint AS user_id,
		     g.id AS group_id,
		     g.name AS group_name,
		     gpa.created_at AS created_at
		   FROM iam_group_policy_attachments gpa
		   JOIN iam_groups g ON g.id = gpa.group_id
		   WHERE gpa.policy_id = $1
		 ) attachments
		 ORDER BY created_at ASC, attachment_type ASC`,
		policyID,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.PolicyAttachment, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

func (r *PolicyRepositoryImpl) DeletePolicy(ctx context.Context, policyID uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var policy policyModel
	if err := tx.GetContext(
		ctx,
		&policy,
		`SELECT id, scope, name, description, is_system, created_at, updated_at
		 FROM iam_policies
		 WHERE id = $1`,
		policyID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return iamdomain.ErrPolicyNotFound
		}
		return err
	}
	if policy.IsSystem {
		return iamdomain.ErrImmutablePolicy
	}

	var attachmentCount int
	if err := tx.GetContext(
		ctx,
		&attachmentCount,
		`SELECT
			(SELECT COUNT(*) FROM iam_role_policy_attachments WHERE policy_id = $1) +
			(SELECT COUNT(*) FROM iam_user_policy_attachments WHERE policy_id = $1) +
			(SELECT COUNT(*) FROM iam_tenant_user_policy_attachments WHERE policy_id = $1) +
			(SELECT COUNT(*) FROM iam_group_policy_attachments WHERE policy_id = $1)`,
		policyID,
	); err != nil {
		return err
	}
	if attachmentCount > 0 {
		return iamdomain.ErrPolicyInUse
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM iam_policies WHERE id = $1`, policyID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PolicyRepositoryImpl) ListRoleStatements(
	ctx context.Context,
	roleID uint64,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.created_at
		 FROM iam_policy_statements ps
		 JOIN iam_policies p ON p.id = ps.policy_id
		 JOIN iam_role_policy_attachments arpa ON arpa.policy_id = p.id
		 WHERE arpa.role_id = $1
		 ORDER BY ps.id ASC`,
		roleID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) ListPlatformUserStatements(
	ctx context.Context,
	userID uint,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_user_policy_attachments upa ON upa.policy_id = p.id
		   WHERE upa.user_id = $1 AND upa.scope = 'platform'
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, ups.policy_name AS policy_name, ups.effect, ups.action_pattern, ups.resource_pattern, ups.created_at
		   FROM iam_platform_user_inline_policy_statements ups
		   WHERE ups.user_id = $1
		 ) user_statements
		 ORDER BY created_at ASC, policy_name ASC, action_pattern ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) ListTenantUserStatements(
	ctx context.Context,
	tenantID string,
	userID uint,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_tenant_user_policy_attachments tupa ON tupa.policy_id = p.id
		   WHERE tupa.tenant_id = $1 AND tupa.user_id = $2
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, tups.policy_name AS policy_name, tups.effect, tups.action_pattern, tups.resource_pattern, tups.created_at
		   FROM iam_tenant_user_inline_policy_statements tups
		   WHERE tups.tenant_id = $1 AND tups.user_id = $2
		 ) user_statements
		 ORDER BY created_at ASC, policy_name ASC, action_pattern ASC`,
		tenantID,
		userID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) ListPlatformGroupStatements(
	ctx context.Context,
	userID uint,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_group_policy_attachments gpa ON gpa.policy_id = p.id
		   JOIN iam_group_members gm ON gm.group_id = gpa.group_id
		   JOIN iam_groups g ON g.id = gpa.group_id
		   WHERE gm.user_id = $1 AND g.scope = 'platform'
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, gps.policy_name AS policy_name, gps.effect, gps.action_pattern, gps.resource_pattern, gps.created_at
		   FROM iam_group_inline_policy_statements gps
		   JOIN iam_group_members gm ON gm.group_id = gps.group_id
		   JOIN iam_groups g ON g.id = gps.group_id
		   WHERE gm.user_id = $1 AND g.scope = 'platform'
		 ) group_statements
		 ORDER BY created_at ASC, policy_name ASC, action_pattern ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) ListTenantGroupStatements(
	ctx context.Context,
	tenantID string,
	userID uint,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_group_policy_attachments gpa ON gpa.policy_id = p.id
		   JOIN iam_group_members gm ON gm.group_id = gpa.group_id
		   JOIN iam_groups g ON g.id = gpa.group_id
		   WHERE gm.user_id = $1 AND g.scope = 'tenant' AND g.tenant_id = $2
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, gps.policy_name AS policy_name, gps.effect, gps.action_pattern, gps.resource_pattern, gps.created_at
		   FROM iam_group_inline_policy_statements gps
		   JOIN iam_group_members gm ON gm.group_id = gps.group_id
		   JOIN iam_groups g ON g.id = gps.group_id
		   WHERE gm.user_id = $1 AND g.scope = 'tenant' AND g.tenant_id = $2
		 ) group_statements
		 ORDER BY created_at ASC, policy_name ASC, action_pattern ASC`,
		userID,
		tenantID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) AttachPlatformUserPolicy(ctx context.Context, userID uint, policyID uint64) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_user_policy_attachments (user_id, scope, policy_id, created_at)
		 VALUES ($1, 'platform', $2, now())
		 ON CONFLICT (user_id, scope, policy_id) DO NOTHING`,
		userID,
		policyID,
	)
	return err
}

func (r *PolicyRepositoryImpl) DetachPlatformUserPolicy(ctx context.Context, userID uint, policyID uint64) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_user_policy_attachments
		 WHERE user_id = $1 AND scope = 'platform' AND policy_id = $2`,
		userID,
		policyID,
	)
	return err
}

func (r *PolicyRepositoryImpl) ListPlatformUserPolicies(ctx context.Context, userID uint) ([]iamdomain.Policy, error) {
	var rows []policyModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT p.id, p.scope, p.name, p.description, p.is_system, p.created_at, p.updated_at
		 FROM iam_policies p
		 JOIN iam_user_policy_attachments upa ON upa.policy_id = p.id
		 WHERE upa.user_id = $1 AND upa.scope = 'platform'
		 ORDER BY p.name ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	return toPolicies(rows), nil
}

func (r *PolicyRepositoryImpl) PutPlatformUserInlinePolicy(ctx context.Context, input iamdomain.PutPlatformUserInlinePolicyInput) error {
	return r.putUserInlinePolicy(ctx, userInlinePolicyModel{
		Scope:       iamdomain.PolicyScopePlatform,
		UserID:      input.UserID,
		Name:        input.Name,
		Description: input.Description,
	}, input.Statements)
}

func (r *PolicyRepositoryImpl) GetPlatformUserInlinePolicy(ctx context.Context, userID uint, name string) (*iamdomain.UserInlinePolicy, error) {
	return r.getPlatformUserInlinePolicy(ctx, userID, name)
}

func (r *PolicyRepositoryImpl) ListPlatformUserInlinePolicies(ctx context.Context, userID uint) ([]iamdomain.UserInlinePolicy, error) {
	return r.listPlatformUserInlinePolicies(ctx, userID)
}

func (r *PolicyRepositoryImpl) DeletePlatformUserInlinePolicy(ctx context.Context, userID uint, name string) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_platform_user_inline_policies WHERE user_id = $1 AND name = $2`,
		userID,
		name,
	)
	return err
}

func (r *PolicyRepositoryImpl) AttachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyID uint64) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_tenant_user_policy_attachments (tenant_id, user_id, policy_id, created_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (tenant_id, user_id, policy_id) DO NOTHING`,
		tenantID,
		userID,
		policyID,
	)
	return err
}

func (r *PolicyRepositoryImpl) DetachTenantUserPolicy(ctx context.Context, tenantID string, userID uint, policyID uint64) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_tenant_user_policy_attachments
		 WHERE tenant_id = $1 AND user_id = $2 AND policy_id = $3`,
		tenantID,
		userID,
		policyID,
	)
	return err
}

func (r *PolicyRepositoryImpl) ListTenantUserPolicies(ctx context.Context, tenantID string, userID uint) ([]iamdomain.Policy, error) {
	var rows []policyModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT p.id, p.scope, p.name, p.description, p.is_system, p.created_at, p.updated_at
		 FROM iam_policies p
		 JOIN iam_tenant_user_policy_attachments tupa ON tupa.policy_id = p.id
		 WHERE tupa.tenant_id = $1 AND tupa.user_id = $2
		 ORDER BY p.name ASC`,
		tenantID,
		userID,
	); err != nil {
		return nil, err
	}
	return toPolicies(rows), nil
}

func (r *PolicyRepositoryImpl) PutTenantUserInlinePolicy(ctx context.Context, input iamdomain.PutTenantUserInlinePolicyInput) error {
	return r.putUserInlinePolicy(ctx, userInlinePolicyModel{
		Scope:       iamdomain.PolicyScopeTenant,
		TenantID:    input.TenantID,
		UserID:      input.UserID,
		Name:        input.Name,
		Description: input.Description,
	}, input.Statements)
}

func (r *PolicyRepositoryImpl) GetTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) (*iamdomain.UserInlinePolicy, error) {
	return r.getTenantUserInlinePolicy(ctx, tenantID, userID, name)
}

func (r *PolicyRepositoryImpl) ListTenantUserInlinePolicies(ctx context.Context, tenantID string, userID uint) ([]iamdomain.UserInlinePolicy, error) {
	return r.listTenantUserInlinePolicies(ctx, tenantID, userID)
}

func (r *PolicyRepositoryImpl) DeleteTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_tenant_user_inline_policies WHERE tenant_id = $1 AND user_id = $2 AND name = $3`,
		tenantID,
		userID,
		name,
	)
	return err
}

func (r *PolicyRepositoryImpl) putUserInlinePolicy(ctx context.Context, policy userInlinePolicyModel, statements []iamdomain.PolicyStatement) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if policy.Scope == iamdomain.PolicyScopePlatform {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO iam_platform_user_inline_policies (user_id, name, description, created_at, updated_at)
			 VALUES ($1, $2, $3, now(), now())
			 ON CONFLICT (user_id, name)
			 DO UPDATE SET description = EXCLUDED.description, updated_at = now()`,
			policy.UserID,
			policy.Name,
			policy.Description,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM iam_platform_user_inline_policy_statements WHERE user_id = $1 AND policy_name = $2`,
			policy.UserID,
			policy.Name,
		); err != nil {
			return err
		}
		for i, statement := range statements {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO iam_platform_user_inline_policy_statements
				  (user_id, policy_name, statement_index, effect, action_pattern, resource_pattern, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				policy.UserID,
				policy.Name,
				i,
				statement.Effect,
				statement.ActionPattern,
				coalescePattern(statement.ResourcePattern),
				statement.CreatedAt,
			); err != nil {
				return err
			}
		}
	} else {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO iam_tenant_user_inline_policies (tenant_id, user_id, name, description, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, now(), now())
			 ON CONFLICT (tenant_id, user_id, name)
			 DO UPDATE SET description = EXCLUDED.description, updated_at = now()`,
			policy.TenantID,
			policy.UserID,
			policy.Name,
			policy.Description,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM iam_tenant_user_inline_policy_statements WHERE tenant_id = $1 AND user_id = $2 AND policy_name = $3`,
			policy.TenantID,
			policy.UserID,
			policy.Name,
		); err != nil {
			return err
		}
		for i, statement := range statements {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO iam_tenant_user_inline_policy_statements
				  (tenant_id, user_id, policy_name, statement_index, effect, action_pattern, resource_pattern, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				policy.TenantID,
				policy.UserID,
				policy.Name,
				i,
				statement.Effect,
				statement.ActionPattern,
				coalescePattern(statement.ResourcePattern),
				statement.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *PolicyRepositoryImpl) getPlatformUserInlinePolicy(ctx context.Context, userID uint, name string) (*iamdomain.UserInlinePolicy, error) {
	var policy userInlinePolicyModel
	if err := r.db.GetContext(
		ctx,
		&policy,
		`SELECT 'platform' AS scope, '' AS tenant_id, user_id, name, description, created_at, updated_at
		 FROM iam_platform_user_inline_policies
		 WHERE user_id = $1 AND name = $2`,
		userID,
		name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrPolicyNotFound
		}
		return nil, err
	}
	statements, err := r.listPlatformUserInlinePolicyStatements(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	entity := policy.toEntity(statements)
	return &entity, nil
}

func (r *PolicyRepositoryImpl) listPlatformUserInlinePolicies(ctx context.Context, userID uint) ([]iamdomain.UserInlinePolicy, error) {
	var policies []userInlinePolicyModel
	if err := r.db.SelectContext(
		ctx,
		&policies,
		`SELECT 'platform' AS scope, '' AS tenant_id, user_id, name, description, created_at, updated_at
		 FROM iam_platform_user_inline_policies
		 WHERE user_id = $1
		 ORDER BY name ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.UserInlinePolicy, 0, len(policies))
	for _, policy := range policies {
		statements, err := r.listPlatformUserInlinePolicyStatements(ctx, userID, policy.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, policy.toEntity(statements))
	}
	return out, nil
}

func (r *PolicyRepositoryImpl) getTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) (*iamdomain.UserInlinePolicy, error) {
	var policy userInlinePolicyModel
	if err := r.db.GetContext(
		ctx,
		&policy,
		`SELECT 'tenant' AS scope, tenant_id, user_id, name, description, created_at, updated_at
		 FROM iam_tenant_user_inline_policies
		 WHERE tenant_id = $1 AND user_id = $2 AND name = $3`,
		tenantID,
		userID,
		name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrPolicyNotFound
		}
		return nil, err
	}
	statements, err := r.listTenantUserInlinePolicyStatements(ctx, tenantID, userID, name)
	if err != nil {
		return nil, err
	}
	entity := policy.toEntity(statements)
	return &entity, nil
}

func (r *PolicyRepositoryImpl) listTenantUserInlinePolicies(ctx context.Context, tenantID string, userID uint) ([]iamdomain.UserInlinePolicy, error) {
	var policies []userInlinePolicyModel
	if err := r.db.SelectContext(
		ctx,
		&policies,
		`SELECT 'tenant' AS scope, tenant_id, user_id, name, description, created_at, updated_at
		 FROM iam_tenant_user_inline_policies
		 WHERE tenant_id = $1 AND user_id = $2
		 ORDER BY name ASC`,
		tenantID,
		userID,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.UserInlinePolicy, 0, len(policies))
	for _, policy := range policies {
		statements, err := r.listTenantUserInlinePolicyStatements(ctx, tenantID, userID, policy.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, policy.toEntity(statements))
	}
	return out, nil
}

func (r *PolicyRepositoryImpl) listPlatformUserInlinePolicyStatements(ctx context.Context, userID uint, name string) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT 0 AS id, 0 AS policy_id, policy_name, effect, action_pattern, resource_pattern, created_at
		 FROM iam_platform_user_inline_policy_statements
		 WHERE user_id = $1 AND policy_name = $2
		 ORDER BY statement_index ASC`,
		userID,
		name,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) listTenantUserInlinePolicyStatements(ctx context.Context, tenantID string, userID uint, name string) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT 0 AS id, 0 AS policy_id, policy_name, effect, action_pattern, resource_pattern, created_at
		 FROM iam_tenant_user_inline_policy_statements
		 WHERE tenant_id = $1 AND user_id = $2 AND policy_name = $3
		 ORDER BY statement_index ASC`,
		tenantID,
		userID,
		name,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func toPolicyStatements(rows []policyStatementModel) []iamdomain.PolicyStatement {
	out := make([]iamdomain.PolicyStatement, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out
}

func toPolicies(rows []policyModel) []iamdomain.Policy {
	out := make([]iamdomain.Policy, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out
}

func coalescePattern(pattern string) string {
	if pattern == "" {
		return "*"
	}
	return pattern
}
