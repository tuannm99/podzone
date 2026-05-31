package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
)

type GroupRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.GroupRepository = (*GroupRepositoryImpl)(nil)

func NewGroupRepository(p repoParams) outputport.GroupRepository {
	return &GroupRepositoryImpl{db: p.DB}
}

func (r *GroupRepositoryImpl) CreateGroup(ctx context.Context, group entity.Group) (*entity.Group, error) {
	var out groupModel
	if err := r.db.GetContext(
		ctx,
		&out,
		`INSERT INTO iam_groups (scope, tenant_id, name, description, is_system, created_at, updated_at)
		 VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7)
		 RETURNING id, scope, COALESCE(tenant_id, '') AS tenant_id, name, description, is_system, created_at, updated_at`,
		group.Scope,
		group.TenantID,
		group.Name,
		group.Description,
		group.IsSystem,
		group.CreatedAt,
		group.UpdatedAt,
	); err != nil {
		return nil, err
	}
	entity := out.toEntity()
	return &entity, nil
}

func (r *GroupRepositoryImpl) GetByID(ctx context.Context, groupID uint64) (*entity.Group, error) {
	var out groupModel
	if err := r.db.GetContext(
		ctx,
		&out,
		`SELECT id, scope, COALESCE(tenant_id, '') AS tenant_id, name, description, is_system, created_at, updated_at
		 FROM iam_groups
		 WHERE id = $1`,
		groupID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrGroupNotFound
		}
		return nil, err
	}
	group := out.toEntity()
	return &group, nil
}

func (r *GroupRepositoryImpl) ListGroups(ctx context.Context, scope string, tenantID string) ([]entity.Group, error) {
	query := `SELECT id, scope, COALESCE(tenant_id, '') AS tenant_id, name, description, is_system, created_at, updated_at FROM iam_groups WHERE 1=1`
	args := []any{}
	if scope != "" {
		query += ` AND scope = $1`
		args = append(args, scope)
	}
	if tenantID != "" {
		query += fmt.Sprintf(` AND tenant_id = $%d`, len(args)+1)
		args = append(args, tenantID)
	}
	query += ` ORDER BY scope ASC, name ASC`

	var rows []groupModel
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]entity.Group, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

func (r *GroupRepositoryImpl) DeleteGroup(ctx context.Context, groupID uint64) error {
	var group groupModel
	if err := r.db.GetContext(
		ctx,
		&group,
		`SELECT id, scope, COALESCE(tenant_id, '') AS tenant_id, name, description, is_system, created_at, updated_at
		 FROM iam_groups
		 WHERE id = $1`,
		groupID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrGroupNotFound
		}
		return err
	}
	if group.IsSystem {
		return entity.ErrImmutableGroup
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_groups WHERE id = $1`, groupID)
	return err
}

func (r *GroupRepositoryImpl) PutInlinePolicy(ctx context.Context, input entity.PutGroupInlinePolicyInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO iam_group_inline_policies (group_id, name, description, created_at, updated_at)
		 VALUES ($1, $2, $3, now(), now())
		 ON CONFLICT (group_id, name)
		 DO UPDATE SET description = EXCLUDED.description, updated_at = now()`,
		input.GroupID,
		input.Name,
		input.Description,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM iam_group_inline_policy_statements WHERE group_id = $1 AND policy_name = $2`,
		input.GroupID,
		input.Name,
	); err != nil {
		return err
	}

	for i, statement := range input.Statements {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO iam_group_inline_policy_statements
			  (group_id, policy_name, statement_index, effect, action_pattern, resource_pattern, conditions_json, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, '[]', $7)`,
			input.GroupID,
			input.Name,
			i,
			statement.Effect,
			statement.ActionPattern,
			coalescePattern(statement.ResourcePattern),
			statement.CreatedAt,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *GroupRepositoryImpl) GetInlinePolicy(
	ctx context.Context,
	groupID uint64,
	name string,
) (*entity.GroupInlinePolicy, error) {
	var policy groupInlinePolicyModel
	if err := r.db.GetContext(
		ctx,
		&policy,
		`SELECT group_id, name, description, created_at, updated_at
		 FROM iam_group_inline_policies
		 WHERE group_id = $1 AND name = $2`,
		groupID,
		name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrPolicyNotFound
		}
		return nil, err
	}
	statements, err := r.listInlinePolicyStatements(ctx, groupID, name)
	if err != nil {
		return nil, err
	}
	entity := policy.toEntity(statements)
	return &entity, nil
}

func (r *GroupRepositoryImpl) ListInlinePolicies(
	ctx context.Context,
	groupID uint64,
) ([]entity.GroupInlinePolicy, error) {
	var policies []groupInlinePolicyModel
	if err := r.db.SelectContext(
		ctx,
		&policies,
		`SELECT group_id, name, description, created_at, updated_at
		 FROM iam_group_inline_policies
		 WHERE group_id = $1
		 ORDER BY name ASC`,
		groupID,
	); err != nil {
		return nil, err
	}
	out := make([]entity.GroupInlinePolicy, 0, len(policies))
	for _, policy := range policies {
		statements, err := r.listInlinePolicyStatements(ctx, groupID, policy.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, policy.toEntity(statements))
	}
	return out, nil
}

func (r *GroupRepositoryImpl) DeleteInlinePolicy(ctx context.Context, groupID uint64, name string) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_group_inline_policies WHERE group_id = $1 AND name = $2`,
		groupID,
		name,
	)
	return err
}

func (r *GroupRepositoryImpl) AddMember(ctx context.Context, groupID uint64, userID uint) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_group_members (group_id, user_id, created_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (group_id, user_id) DO NOTHING`,
		groupID,
		userID,
	)
	return err
}

func (r *GroupRepositoryImpl) RemoveMember(ctx context.Context, groupID uint64, userID uint) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_group_members WHERE group_id = $1 AND user_id = $2`,
		groupID,
		userID,
	)
	return err
}

func (r *GroupRepositoryImpl) ListMembers(ctx context.Context, groupID uint64) ([]uint, error) {
	var rows []uint
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT user_id FROM iam_group_members WHERE group_id = $1 ORDER BY user_id ASC`,
		groupID,
	); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GroupRepositoryImpl) AttachPolicy(ctx context.Context, groupID uint64, policyID uint64) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_group_policy_attachments (group_id, policy_id, created_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (group_id, policy_id) DO NOTHING`,
		groupID,
		policyID,
	)
	return err
}

func (r *GroupRepositoryImpl) DetachPolicy(ctx context.Context, groupID uint64, policyID uint64) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_group_policy_attachments WHERE group_id = $1 AND policy_id = $2`,
		groupID,
		policyID,
	)
	return err
}

func (r *GroupRepositoryImpl) ListPolicies(ctx context.Context, groupID uint64) ([]entity.Policy, error) {
	var rows []policyModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT p.id, p.scope, p.name, p.description, p.is_system, p.created_at, p.updated_at
		 FROM iam_policies p
		 JOIN iam_group_policy_attachments gpa ON gpa.policy_id = p.id
		 WHERE gpa.group_id = $1
		 ORDER BY p.name ASC`,
		groupID,
	); err != nil {
		return nil, err
	}
	return toPolicies(rows), nil
}

func (r *GroupRepositoryImpl) listInlinePolicyStatements(
	ctx context.Context,
	groupID uint64,
	name string,
) ([]entity.PolicyStatement, error) {
	var rows []struct {
		Effect          string    `db:"effect"`
		ActionPattern   string    `db:"action_pattern"`
		ResourcePattern string    `db:"resource_pattern"`
		ConditionsJSON  string    `db:"conditions_json"`
		CreatedAt       time.Time `db:"created_at"`
	}
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT effect, action_pattern, resource_pattern, '[]' AS conditions_json, created_at
		 FROM iam_group_inline_policy_statements
		 WHERE group_id = $1 AND policy_name = $2
		 ORDER BY statement_index ASC`,
		groupID,
		name,
	); err != nil {
		return nil, err
	}
	out := make([]entity.PolicyStatement, 0, len(rows))
	for _, row := range rows {
		out = append(out, entity.PolicyStatement{
			PolicyName:      name,
			Effect:          row.Effect,
			ActionPattern:   row.ActionPattern,
			ResourcePattern: row.ResourcePattern,
			Conditions:      nil,
			CreatedAt:       row.CreatedAt,
		})
	}
	return out, nil
}
