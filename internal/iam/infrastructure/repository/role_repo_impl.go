package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	entity "github.com/tuannm99/podzone/internal/iam/entity"
	"github.com/tuannm99/podzone/internal/iam/outputport"
)

type RoleRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.RoleRepository = (*RoleRepositoryImpl)(nil)

func NewRoleRepository(p repoParams) outputport.RoleRepository {
	return &RoleRepositoryImpl{db: p.DB}
}

func (r *RoleRepositoryImpl) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var out roleModel
	if err := r.db.GetContext(
		ctx,
		&out,
		`SELECT id, scope, name, description, is_system, created_at, updated_at FROM iam_roles WHERE name = $1`,
		name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrRoleNotFound
		}
		return nil, err
	}
	return &entity.Role{
		ID:          out.ID,
		Scope:       out.Scope,
		Name:        out.Name,
		Description: out.Description,
		IsSystem:    out.IsSystem,
		CreatedAt:   out.CreatedAt,
		UpdatedAt:   out.UpdatedAt,
	}, nil
}

func (r *RoleRepositoryImpl) RoleHasPermission(ctx context.Context, roleID uint64, permission string) (bool, error) {
	var exists bool
	err := r.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS (
			SELECT 1
			FROM iam_role_permissions rp
			JOIN iam_permissions p ON p.id = rp.permission_id
			WHERE rp.role_id = $1 AND p.name = $2
		)`,
		roleID,
		permission,
	)
	return exists, err
}

func (r *RoleRepositoryImpl) PutTrustPolicy(
	ctx context.Context,
	roleID uint64,
	statements []entity.RoleTrustStatement,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM iam_role_trust_statements WHERE role_id = $1`, roleID); err != nil {
		return err
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO iam_role_trust_statements
			 (role_id, effect, principal_type, principal_pattern, tenant_pattern, external_id_pattern, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			roleID,
			statement.Effect,
			statement.PrincipalType,
			statement.PrincipalPattern,
			statement.TenantPattern,
			statement.ExternalIDPattern,
			statement.CreatedAt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RoleRepositoryImpl) GetTrustPolicy(ctx context.Context, roleID uint64) ([]entity.RoleTrustStatement, error) {
	rows := make([]roleTrustStatementModel, 0)
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, role_id, effect, principal_type, principal_pattern, tenant_pattern, external_id_pattern, created_at
		 FROM iam_role_trust_statements
		 WHERE role_id = $1
		 ORDER BY id ASC`,
		roleID,
	); err != nil {
		return nil, err
	}
	out := make([]entity.RoleTrustStatement, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

func (r *RoleRepositoryImpl) DeleteTrustPolicy(ctx context.Context, roleID uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_role_trust_statements WHERE role_id = $1`, roleID)
	return err
}

func (r *RoleRepositoryImpl) PutPermissionBoundary(ctx context.Context, roleID uint64, policyID uint64) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_role_permission_boundaries (role_id, policy_id, created_at, updated_at)
		 VALUES ($1, $2, now(), now())
		 ON CONFLICT (role_id)
		 DO UPDATE SET policy_id = EXCLUDED.policy_id, updated_at = now()`,
		roleID,
		policyID,
	)
	return err
}

func (r *RoleRepositoryImpl) GetPermissionBoundary(ctx context.Context, roleID uint64) (*entity.RolePermissionBoundary, error) {
	var row struct {
		RoleID     uint64    `db:"role_id"`
		RoleName   string    `db:"role_name"`
		PolicyID   uint64    `db:"policy_id"`
		PolicyName string    `db:"policy_name"`
		CreatedAt  time.Time `db:"created_at"`
	}
	if err := r.db.GetContext(
		ctx,
		&row,
		`SELECT rpb.role_id, r.name AS role_name, rpb.policy_id, p.name AS policy_name, rpb.created_at
		 FROM iam_role_permission_boundaries rpb
		 JOIN iam_roles r ON r.id = rpb.role_id
		 JOIN iam_policies p ON p.id = rpb.policy_id
		 WHERE rpb.role_id = $1`,
		roleID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &entity.RolePermissionBoundary{
		RoleID:     row.RoleID,
		RoleName:   row.RoleName,
		PolicyID:   row.PolicyID,
		PolicyName: row.PolicyName,
		CreatedAt:  row.CreatedAt,
	}, nil
}

func (r *RoleRepositoryImpl) GetPermissionBoundaryStatements(ctx context.Context, roleID uint64) ([]entity.PolicyStatement, error) {
	rows := make([]policyStatementModel, 0)
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		 FROM iam_role_permission_boundaries rpb
		 JOIN iam_policies p ON p.id = rpb.policy_id
		 JOIN iam_policy_statements ps ON ps.policy_id = p.id
		 WHERE rpb.role_id = $1
		 ORDER BY ps.id ASC`,
		roleID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}

func (r *RoleRepositoryImpl) DeletePermissionBoundary(ctx context.Context, roleID uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_role_permission_boundaries WHERE role_id = $1`, roleID)
	return err
}
