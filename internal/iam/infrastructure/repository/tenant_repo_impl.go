package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type repoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-auth"`
}

type TenantRepositoryImpl struct {
	db *sqlx.DB
}

func NewTenantRepository(p repoParams) iamdomain.TenantRepository {
	return &TenantRepositoryImpl{db: p.DB}
}

func (r *TenantRepositoryImpl) Create(ctx context.Context, tenant iamdomain.Tenant) (*iamdomain.Tenant, error) {
	var out tenantModel
	err := r.db.GetContext(
		ctx,
		&out,
		`INSERT INTO tenants (id, slug, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, slug, name, created_at, updated_at`,
		tenant.ID,
		tenant.Slug,
		tenant.Name,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, iamdomain.ErrTenantSlugTaken
		}
		return nil, err
	}
	return &iamdomain.Tenant{
		ID:        out.ID,
		Slug:      out.Slug,
		Name:      out.Name,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}

func (r *TenantRepositoryImpl) GetByID(ctx context.Context, tenantID string) (*iamdomain.Tenant, error) {
	var out tenantModel
	if err := r.db.GetContext(ctx, &out, `SELECT id, slug, name, created_at, updated_at FROM tenants WHERE id = $1`, tenantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrTenantNotFound
		}
		return nil, err
	}
	return &iamdomain.Tenant{
		ID:        out.ID,
		Slug:      out.Slug,
		Name:      out.Name,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}
