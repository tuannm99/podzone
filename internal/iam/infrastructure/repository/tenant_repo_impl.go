package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	entity "github.com/tuannm99/podzone/internal/iam/entity"
	"github.com/tuannm99/podzone/internal/iam/outputport"
)

type repoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-iam"`
}

type TenantRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.TenantRepository = (*TenantRepositoryImpl)(nil)

func NewTenantRepository(p repoParams) outputport.TenantRepository {
	return &TenantRepositoryImpl{db: p.DB}
}

func (r *TenantRepositoryImpl) Create(ctx context.Context, tenant entity.Tenant) (*entity.Tenant, error) {
	var out tenantModel
	err := r.db.GetContext(
		ctx,
		&out,
		`INSERT INTO tenants (id, slug, name, org_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, slug, name, org_id, created_at, updated_at`,
		tenant.ID,
		tenant.Slug,
		tenant.Name,
		tenant.OrgID,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, entity.ErrTenantSlugTaken
		}
		return nil, err
	}
	return &entity.Tenant{
		ID:        out.ID,
		Slug:      out.Slug,
		Name:      out.Name,
		OrgID:     out.OrgID,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}

func (r *TenantRepositoryImpl) GetByID(ctx context.Context, tenantID string) (*entity.Tenant, error) {
	var out tenantModel
	if err := r.db.GetContext(ctx, &out, `SELECT id, slug, name, org_id, created_at, updated_at FROM tenants WHERE id = $1`, tenantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrTenantNotFound
		}
		return nil, err
	}
	return &entity.Tenant{
		ID:        out.ID,
		Slug:      out.Slug,
		Name:      out.Name,
		OrgID:     out.OrgID,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}

func (r *TenantRepositoryImpl) AttachOrganization(ctx context.Context, tenantID string, orgID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tenants SET org_id = $2, updated_at = now() WHERE id = $1`, tenantID, orgID)
	return err
}

func (r *TenantRepositoryImpl) DetachOrganization(ctx context.Context, tenantID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tenants SET org_id = '', updated_at = now() WHERE id = $1`, tenantID)
	return err
}
