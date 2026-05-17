package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type GroupRepositoryImpl struct {
	db *sqlx.DB
}

func NewGroupRepository(p repoParams) iamdomain.GroupRepository {
	return &GroupRepositoryImpl{db: p.DB}
}

func (r *GroupRepositoryImpl) CreateGroup(ctx context.Context, group iamdomain.Group) (*iamdomain.Group, error) {
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

func (r *GroupRepositoryImpl) ListGroups(ctx context.Context, scope string, tenantID string) ([]iamdomain.Group, error) {
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
	out := make([]iamdomain.Group, 0, len(rows))
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
			return iamdomain.ErrGroupNotFound
		}
		return err
	}
	if group.IsSystem {
		return iamdomain.ErrImmutableGroup
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_groups WHERE id = $1`, groupID)
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

func (r *GroupRepositoryImpl) ListPolicies(ctx context.Context, groupID uint64) ([]iamdomain.Policy, error) {
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
