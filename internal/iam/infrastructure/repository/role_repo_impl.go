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
		`SELECT id, name, description, is_system, created_at, updated_at FROM iam_roles WHERE name = $1`,
		name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrRoleNotFound
		}
		return nil, err
	}
	return &iamdomain.Role{
		ID:          out.ID,
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
