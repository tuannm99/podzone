package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"go.uber.org/fx"
)

type IAMProjectionRepoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-auth"`
}

type IAMProjectionRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.IAMProjectionRepository = (*IAMProjectionRepositoryImpl)(nil)

func NewIAMProjectionRepositoryImpl(p IAMProjectionRepoParams) outputport.IAMProjectionRepository {
	return &IAMProjectionRepositoryImpl{db: p.DB}
}

func (r *IAMProjectionRepositoryImpl) UpsertTenant(
	ctx context.Context,
	tenantID string,
	slug string,
	name string,
) error {
	query, args, err := psql.
		Insert("iam_tenants_projection").
		Columns("tenant_id", "slug", "name").
		Values(tenantID, slug, name).
		Suffix(`ON CONFLICT (tenant_id) DO UPDATE SET slug = EXCLUDED.slug, name = EXCLUDED.name, updated_at = now()`).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *IAMProjectionRepositoryImpl) UpsertTenantMembership(
	ctx context.Context,
	tenantID string,
	userID uint,
	roleName string,
	status string,
) error {
	query, args, err := psql.
		Insert("iam_tenant_memberships_projection").
		Columns("tenant_id", "user_id", "role_name", "status").
		Values(tenantID, userID, roleName, status).
		Suffix(`ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_name = EXCLUDED.role_name, status = EXCLUDED.status, updated_at = now()`).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *IAMProjectionRepositoryImpl) GetTenantMembership(
	ctx context.Context,
	tenantID string,
	userID uint,
) (*outputport.TenantMembershipProjection, error) {
	var row outputport.TenantMembershipProjection
	if err := r.db.GetContext(
		ctx,
		&row,
		`SELECT tenant_id, user_id, role_name, status
		 FROM iam_tenant_memberships_projection
		 WHERE tenant_id = $1 AND user_id = $2`,
		tenantID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}
