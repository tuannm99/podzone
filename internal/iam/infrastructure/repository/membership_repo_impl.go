package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type MembershipRepositoryImpl struct {
	db *sqlx.DB
}

func NewMembershipRepository(p repoParams) iamdomain.MembershipRepository {
	return &MembershipRepositoryImpl{db: p.DB}
}

func (r *MembershipRepositoryImpl) Upsert(ctx context.Context, membership iamdomain.Membership) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tenant_memberships (tenant_id, user_id, role_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, user_id) DO UPDATE
		 SET role_id = EXCLUDED.role_id,
		     status = EXCLUDED.status,
		     updated_at = EXCLUDED.updated_at`,
		membership.TenantID,
		membership.UserID,
		membership.RoleID,
		membership.Status,
		membership.CreatedAt,
		membership.UpdatedAt,
	)
	return err
}

func (r *MembershipRepositoryImpl) GetByTenantAndUser(
	ctx context.Context,
	tenantID string,
	userID uint,
) (*iamdomain.Membership, error) {
	var out membershipModel
	err := r.db.GetContext(
		ctx,
		&out,
		`SELECT tm.tenant_id, tm.user_id, tm.role_id, r.name AS role_name, tm.status, tm.created_at, tm.updated_at
		 FROM tenant_memberships tm
		 JOIN iam_roles r ON r.id = tm.role_id
		 WHERE tm.tenant_id = $1 AND tm.user_id = $2`,
		tenantID,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrMembershipNotFound
		}
		return nil, err
	}
	return &iamdomain.Membership{
		TenantID:  out.TenantID,
		UserID:    out.UserID,
		RoleID:    out.RoleID,
		RoleName:  out.RoleName,
		Status:    out.Status,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}

func (r *MembershipRepositoryImpl) ListByTenant(ctx context.Context, tenantID string) ([]iamdomain.Membership, error) {
	var rows []membershipModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT tm.tenant_id, tm.user_id, tm.role_id, r.name AS role_name, tm.status, tm.created_at, tm.updated_at
		 FROM tenant_memberships tm
		 JOIN iam_roles r ON r.id = tm.role_id
		 WHERE tm.tenant_id = $1
		 ORDER BY tm.created_at ASC`,
		tenantID,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.Membership, 0, len(rows))
	for _, row := range rows {
		out = append(out, iamdomain.Membership{
			TenantID:  row.TenantID,
			UserID:    row.UserID,
			RoleID:    row.RoleID,
			RoleName:  row.RoleName,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *MembershipRepositoryImpl) ListByUser(ctx context.Context, userID uint) ([]iamdomain.Membership, error) {
	var rows []membershipModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT tm.tenant_id, tm.user_id, tm.role_id, r.name AS role_name, tm.status, tm.created_at, tm.updated_at
		 FROM tenant_memberships tm
		 JOIN iam_roles r ON r.id = tm.role_id
		 WHERE tm.user_id = $1
		 ORDER BY tm.created_at ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.Membership, 0, len(rows))
	for _, row := range rows {
		out = append(out, iamdomain.Membership{
			TenantID:  row.TenantID,
			UserID:    row.UserID,
			RoleID:    row.RoleID,
			RoleName:  row.RoleName,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *MembershipRepositoryImpl) Delete(ctx context.Context, tenantID string, userID uint) error {
	res, err := r.db.ExecContext(
		ctx,
		`DELETE FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2`,
		tenantID,
		userID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return iamdomain.ErrMembershipNotFound
	}
	return nil
}
