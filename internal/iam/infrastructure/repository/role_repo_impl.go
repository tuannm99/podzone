package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type RoleRepositoryImpl struct {
	db *sqlx.DB
}

func NewRoleRepository(p repoParams) iamdomain.RoleRepository {
	return &RoleRepositoryImpl{db: p.DB}
}

func (r *RoleRepositoryImpl) GetByName(ctx context.Context, name string) (*iamdomain.Role, error) {
	var out roleModel
	if err := r.db.GetContext(
		ctx,
		&out,
		`SELECT id, scope, name, description, is_system, created_at, updated_at FROM iam_roles WHERE name = $1`,
		name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrRoleNotFound
		}
		return nil, err
	}
	return &iamdomain.Role{
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
	statements []iamdomain.RoleTrustStatement,
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
			 (role_id, effect, principal_type, principal_pattern, tenant_pattern, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			roleID,
			statement.Effect,
			statement.PrincipalType,
			statement.PrincipalPattern,
			statement.TenantPattern,
			statement.CreatedAt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RoleRepositoryImpl) GetTrustPolicy(ctx context.Context, roleID uint64) ([]iamdomain.RoleTrustStatement, error) {
	rows := make([]roleTrustStatementModel, 0)
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, role_id, effect, principal_type, principal_pattern, tenant_pattern, created_at
		 FROM iam_role_trust_statements
		 WHERE role_id = $1
		 ORDER BY id ASC`,
		roleID,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.RoleTrustStatement, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

func (r *RoleRepositoryImpl) DeleteTrustPolicy(ctx context.Context, roleID uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_role_trust_statements WHERE role_id = $1`, roleID)
	return err
}
