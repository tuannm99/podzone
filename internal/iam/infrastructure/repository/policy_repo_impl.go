package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/jmoiron/sqlx"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type PolicyRepositoryImpl struct {
	db *sqlx.DB
}

var _ iamdomain.PolicyRepository = (*PolicyRepositoryImpl)(nil)

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
		`INSERT INTO iam_policies (scope, name, description, is_system, default_version, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'v1', $5, $6)
		 RETURNING id, scope, name, description, is_system, default_version, created_at, updated_at`,
		policy.Scope,
		policy.Name,
		policy.Description,
		policy.IsSystem,
		policy.CreatedAt,
		policy.UpdatedAt,
	); err != nil {
		return nil, nil, err
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO iam_policy_versions (policy_id, version, is_default, created_at)
		 VALUES ($1, 'v1', TRUE, $2)`,
		created.ID,
		policy.CreatedAt,
	); err != nil {
		return nil, nil, err
	}

	outStatements := make([]iamdomain.PolicyStatement, 0, len(statements))
	for _, statement := range statements {
		var createdStatement policyStatementModel
		if err := tx.GetContext(
			ctx,
			&createdStatement,
			`INSERT INTO iam_policy_statements (policy_id, effect, action_pattern, resource_pattern, conditions_json, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING id, policy_id, $7 AS policy_name, effect, action_pattern, resource_pattern, conditions_json, created_at`,
			created.ID,
			statement.Effect,
			statement.ActionPattern,
			coalescePattern(statement.ResourcePattern),
			mustMarshalPolicyConditions(statement.Conditions),
			statement.CreatedAt,
			created.Name,
		); err != nil {
			return nil, nil, err
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO iam_policy_version_statements (policy_id, version, statement_index, effect, action_pattern, resource_pattern, conditions_json, created_at)
			 VALUES ($1, 'v1', $2, $3, $4, $5, $6, $7)`,
			created.ID,
			len(outStatements),
			statement.Effect,
			coalescePattern(statement.ActionPattern),
			coalescePattern(statement.ResourcePattern),
			mustMarshalPolicyConditions(statement.Conditions),
			statement.CreatedAt,
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
		`SELECT id, scope, name, description, is_system, default_version, created_at, updated_at
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
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
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
	query = `SELECT id, scope, name, description, is_system, default_version, created_at, updated_at FROM iam_policies`
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
		   UNION ALL
		   SELECT
		     'role_boundary' AS attachment_type,
		     r.scope AS scope,
		     '' AS tenant_id,
		     r.id AS role_id,
		     r.name AS role_name,
		     0::bigint AS user_id,
		     0::bigint AS group_id,
		     '' AS group_name,
		     rpb.created_at AS created_at
		   FROM iam_role_permission_boundaries rpb
		   JOIN iam_roles r ON r.id = rpb.role_id
		   WHERE rpb.policy_id = $1
		   UNION ALL
		   SELECT
		     'platform_user_boundary' AS attachment_type,
		     'platform' AS scope,
		     '' AS tenant_id,
		     0::bigint AS role_id,
		     '' AS role_name,
		     pub.user_id AS user_id,
		     0::bigint AS group_id,
		     '' AS group_name,
		     pub.created_at AS created_at
		   FROM iam_platform_user_permission_boundaries pub
		   WHERE pub.policy_id = $1
		   UNION ALL
		   SELECT
		     'tenant_user_boundary' AS attachment_type,
		     'tenant' AS scope,
		     tub.tenant_id AS tenant_id,
		     0::bigint AS role_id,
		     '' AS role_name,
		     tub.user_id AS user_id,
		     0::bigint AS group_id,
		     '' AS group_name,
		     tub.created_at AS created_at
		   FROM iam_tenant_user_permission_boundaries tub
		   WHERE tub.policy_id = $1
		   UNION ALL
		   SELECT
		     'service_control_policy' AS attachment_type,
		     'organization' AS scope,
		     osp.org_id AS tenant_id,
		     0::bigint AS role_id,
		     '' AS role_name,
		     0::bigint AS user_id,
		     0::bigint AS group_id,
		     '' AS group_name,
		     osp.created_at AS created_at
		   FROM iam_org_service_control_policies osp
		   WHERE osp.policy_id = $1
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
		`SELECT id, scope, name, description, is_system, default_version, created_at, updated_at
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
			(SELECT COUNT(*) FROM iam_group_policy_attachments WHERE policy_id = $1) +
			(SELECT COUNT(*) FROM iam_role_permission_boundaries WHERE policy_id = $1) +
			(SELECT COUNT(*) FROM iam_platform_user_permission_boundaries WHERE policy_id = $1) +
			(SELECT COUNT(*) FROM iam_tenant_user_permission_boundaries WHERE policy_id = $1) +
			(SELECT COUNT(*) FROM iam_org_service_control_policies WHERE policy_id = $1)`,
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

func (r *PolicyRepositoryImpl) CreatePolicyVersion(
	ctx context.Context,
	policyID uint64,
	policyName string,
	statements []iamdomain.PolicyStatement,
	setAsDefault bool,
) (*iamdomain.PolicyVersion, []iamdomain.PolicyStatement, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var nextVersionNum int
	if err := tx.GetContext(ctx, &nextVersionNum,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(version FROM 2) AS INT)), 0) + 1
		 FROM iam_policy_versions
		 WHERE policy_id = $1`,
		policyID,
	); err != nil {
		return nil, nil, err
	}
	versionLabel := "v" + itoa(nextVersionNum)
	if setAsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE iam_policy_versions SET is_default = FALSE WHERE policy_id = $1`, policyID); err != nil {
			return nil, nil, err
		}
	}
	var versionRow policyVersionModel
	if err := tx.GetContext(
		ctx,
		&versionRow,
		`INSERT INTO iam_policy_versions (policy_id, version, is_default, created_at)
		 VALUES ($1, $2, $3, now())
		 RETURNING id, policy_id, $4 AS policy_name, version, is_default, created_at`,
		policyID,
		versionLabel,
		setAsDefault,
		policyName,
	); err != nil {
		return nil, nil, err
	}

	outStatements := make([]iamdomain.PolicyStatement, 0, len(statements))
	for i, statement := range statements {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO iam_policy_version_statements (policy_id, version, statement_index, effect, action_pattern, resource_pattern, conditions_json, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			policyID,
			versionLabel,
			i,
			statement.Effect,
			coalescePattern(statement.ActionPattern),
			coalescePattern(statement.ResourcePattern),
			mustMarshalPolicyConditions(statement.Conditions),
			statement.CreatedAt,
		); err != nil {
			return nil, nil, err
		}
		outStatements = append(outStatements, iamdomain.PolicyStatement{
			PolicyID:        policyID,
			PolicyName:      policyName,
			Effect:          statement.Effect,
			ActionPattern:   statement.ActionPattern,
			ResourcePattern: coalescePattern(statement.ResourcePattern),
			CreatedAt:       statement.CreatedAt,
		})
	}
	if setAsDefault {
		if err := r.syncDefaultPolicyVersionTx(ctx, tx, policyID, versionLabel); err != nil {
			return nil, nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	entity := versionRow.toEntity()
	return &entity, outStatements, nil
}

func (r *PolicyRepositoryImpl) ListPolicyVersions(ctx context.Context, policyID uint64, policyName string) ([]iamdomain.PolicyVersion, error) {
	var rows []policyVersionModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, policy_id, $2 AS policy_name, version, is_default, created_at
		 FROM iam_policy_versions
		 WHERE policy_id = $1
		 ORDER BY created_at ASC, id ASC`,
		policyID,
		policyName,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.PolicyVersion, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

func (r *PolicyRepositoryImpl) DeletePolicyVersion(ctx context.Context, policyID uint64, version string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var row struct {
		Version   string `db:"version"`
		IsDefault bool   `db:"is_default"`
	}
	if err := tx.GetContext(
		ctx,
		&row,
		`SELECT version, is_default
		 FROM iam_policy_versions
		 WHERE policy_id = $1 AND version = $2`,
		policyID,
		version,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return iamdomain.ErrPolicyVersionNotFound
		}
		return err
	}
	if row.IsDefault {
		return iamdomain.ErrDefaultPolicyVersion
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM iam_policy_versions WHERE policy_id = $1 AND version = $2`, policyID, version); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PolicyRepositoryImpl) SetDefaultPolicyVersion(ctx context.Context, policyID uint64, version string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE iam_policy_versions SET is_default = FALSE WHERE policy_id = $1`, policyID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `UPDATE iam_policy_versions SET is_default = TRUE WHERE policy_id = $1 AND version = $2`, policyID, version)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return iamdomain.ErrPolicyNotFound
	}
	if err := r.syncDefaultPolicyVersionTx(ctx, tx, policyID, version); err != nil {
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
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
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
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, conditions_json, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_user_policy_attachments upa ON upa.policy_id = p.id
		   WHERE upa.user_id = $1 AND upa.scope = 'platform'
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, ups.policy_name AS policy_name, ups.effect, ups.action_pattern, ups.resource_pattern, '[]' AS conditions_json, ups.created_at
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

func (r *PolicyRepositoryImpl) PutPlatformUserPermissionBoundary(
	ctx context.Context,
	userID uint,
	policyID uint64,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_platform_user_permission_boundaries (user_id, policy_id, created_at, updated_at)
		 VALUES ($1, $2, now(), now())
		 ON CONFLICT (user_id)
		 DO UPDATE SET policy_id = EXCLUDED.policy_id, updated_at = now()`,
		userID,
		policyID,
	)
	return err
}

func (r *PolicyRepositoryImpl) GetPlatformUserPermissionBoundary(
	ctx context.Context,
	userID uint,
) (*iamdomain.PermissionBoundary, error) {
	var row permissionBoundaryModel
	if err := r.db.GetContext(
		ctx,
		&row,
		`SELECT 'platform' AS scope, '' AS tenant_id, pub.user_id, pub.policy_id, p.name AS policy_name, pub.created_at
		 FROM iam_platform_user_permission_boundaries pub
		 JOIN iam_policies p ON p.id = pub.policy_id
		 WHERE pub.user_id = $1`,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return row.toEntity(), nil
}

func (r *PolicyRepositoryImpl) GetPlatformUserPermissionBoundaryStatements(
	ctx context.Context,
	userID uint,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		 FROM iam_platform_user_permission_boundaries pub
		 JOIN iam_policies p ON p.id = pub.policy_id
		 JOIN iam_policy_statements ps ON ps.policy_id = p.id
		 WHERE pub.user_id = $1
		 ORDER BY ps.id ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) DeletePlatformUserPermissionBoundary(ctx context.Context, userID uint) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_platform_user_permission_boundaries WHERE user_id = $1`, userID)
	return err
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
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, conditions_json, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_tenant_user_policy_attachments tupa ON tupa.policy_id = p.id
		   WHERE tupa.tenant_id = $1 AND tupa.user_id = $2
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, tups.policy_name AS policy_name, tups.effect, tups.action_pattern, tups.resource_pattern, '[]' AS conditions_json, tups.created_at
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

func (r *PolicyRepositoryImpl) PutTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
	policyID uint64,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_tenant_user_permission_boundaries (tenant_id, user_id, policy_id, created_at, updated_at)
		 VALUES ($1, $2, $3, now(), now())
		 ON CONFLICT (tenant_id, user_id)
		 DO UPDATE SET policy_id = EXCLUDED.policy_id, updated_at = now()`,
		tenantID,
		userID,
		policyID,
	)
	return err
}

func (r *PolicyRepositoryImpl) GetTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
) (*iamdomain.PermissionBoundary, error) {
	var row permissionBoundaryModel
	if err := r.db.GetContext(
		ctx,
		&row,
		`SELECT 'tenant' AS scope, tub.tenant_id, tub.user_id, tub.policy_id, p.name AS policy_name, tub.created_at
		 FROM iam_tenant_user_permission_boundaries tub
		 JOIN iam_policies p ON p.id = tub.policy_id
		 WHERE tub.tenant_id = $1 AND tub.user_id = $2`,
		tenantID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return row.toEntity(), nil
}

func (r *PolicyRepositoryImpl) GetTenantUserPermissionBoundaryStatements(
	ctx context.Context,
	tenantID string,
	userID uint,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		 FROM iam_tenant_user_permission_boundaries tub
		 JOIN iam_policies p ON p.id = tub.policy_id
		 JOIN iam_policy_statements ps ON ps.policy_id = p.id
		 WHERE tub.tenant_id = $1 AND tub.user_id = $2
		 ORDER BY ps.id ASC`,
		tenantID,
		userID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *PolicyRepositoryImpl) DeleteTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_tenant_user_permission_boundaries WHERE tenant_id = $1 AND user_id = $2`, tenantID, userID)
	return err
}

func (r *PolicyRepositoryImpl) syncDefaultPolicyVersionTx(
	ctx context.Context,
	tx *sqlx.Tx,
	policyID uint64,
	version string,
) error {
	if _, err := tx.ExecContext(ctx, `UPDATE iam_policies SET default_version = $2, updated_at = now() WHERE id = $1`, policyID, version); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM iam_policy_statements WHERE policy_id = $1`, policyID); err != nil {
		return err
	}
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO iam_policy_statements (policy_id, effect, action_pattern, resource_pattern, conditions_json, created_at)
		 SELECT policy_id, effect, action_pattern, resource_pattern, conditions_json, created_at
		 FROM iam_policy_version_statements
		 WHERE policy_id = $1 AND version = $2
		 ORDER BY statement_index ASC`,
		policyID,
		version,
	)
	return err
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func (r *PolicyRepositoryImpl) ListPlatformGroupStatements(
	ctx context.Context,
	userID uint,
) ([]iamdomain.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, conditions_json, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_group_policy_attachments gpa ON gpa.policy_id = p.id
		   JOIN iam_group_members gm ON gm.group_id = gpa.group_id
		   JOIN iam_groups g ON g.id = gpa.group_id
		   WHERE gm.user_id = $1 AND g.scope = 'platform'
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, gps.policy_name AS policy_name, gps.effect, gps.action_pattern, gps.resource_pattern, '[]' AS conditions_json, gps.created_at
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
		`SELECT id, policy_id, policy_name, effect, action_pattern, resource_pattern, conditions_json, created_at
		 FROM (
		   SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		   FROM iam_policy_statements ps
		   JOIN iam_policies p ON p.id = ps.policy_id
		   JOIN iam_group_policy_attachments gpa ON gpa.policy_id = p.id
		   JOIN iam_group_members gm ON gm.group_id = gpa.group_id
		   JOIN iam_groups g ON g.id = gpa.group_id
		   WHERE gm.user_id = $1 AND g.scope = 'tenant' AND g.tenant_id = $2
		   UNION ALL
		   SELECT 0 AS id, 0 AS policy_id, gps.policy_name AS policy_name, gps.effect, gps.action_pattern, gps.resource_pattern, '[]' AS conditions_json, gps.created_at
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
		`SELECT 0 AS id, 0 AS policy_id, policy_name, effect, action_pattern, resource_pattern, '[]' AS conditions_json, created_at
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
		`SELECT 0 AS id, 0 AS policy_id, policy_name, effect, action_pattern, resource_pattern, '[]' AS conditions_json, created_at
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

func mustMarshalPolicyConditions(items []iamdomain.PolicyCondition) string {
	if len(items) == 0 {
		return "[]"
	}
	data, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func parsePolicyConditionsJSON(raw string) []iamdomain.PolicyCondition {
	if raw == "" {
		return nil
	}
	out := make([]iamdomain.PolicyCondition, 0)
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
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
