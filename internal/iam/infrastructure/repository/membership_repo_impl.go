package repository

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/pkg/collection"
)

type MembershipRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.MembershipRepository = (*MembershipRepositoryImpl)(nil)

func NewMembershipRepository(p repoParams) outputport.MembershipRepository {
	return &MembershipRepositoryImpl{db: p.DB}
}

func (r *MembershipRepositoryImpl) Upsert(ctx context.Context, membership entity.Membership) error {
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
) (*entity.Membership, error) {
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
			return nil, entity.ErrMembershipNotFound
		}
		return nil, err
	}
	return &entity.Membership{
		TenantID:  out.TenantID,
		UserID:    out.UserID,
		RoleID:    out.RoleID,
		RoleName:  out.RoleName,
		Status:    out.Status,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}

func (r *MembershipRepositoryImpl) ListPageByTenant(
	ctx context.Context,
	tenantID string,
	query collection.Query,
) (collection.Page[entity.Membership], error) {
	page, err := listIAMCollectionModels[membershipModel](
		ctx,
		r.db,
		query,
		"tenant_memberships tm JOIN iam_roles r ON r.id = tm.role_id",
		[]string{
			"tm.tenant_id",
			"tm.user_id",
			"tm.role_id",
			"r.name AS role_name",
			"tm.status",
			"tm.created_at",
			"tm.updated_at",
		},
		[]sq.Sqlizer{sq.Eq{"tm.tenant_id": tenantID}},
		tenantMembershipCollectionColumns,
		[]string{"CAST(tm.user_id AS TEXT)", "r.name", "tm.status"},
		"tm.created_at",
		"tm.user_id ASC",
	)
	if err != nil {
		return collection.Page[entity.Membership]{}, err
	}
	out := make([]entity.Membership, 0, len(page.Items))
	for _, row := range page.Items {
		out = append(out, entity.Membership{
			TenantID:  row.TenantID,
			UserID:    row.UserID,
			RoleID:    row.RoleID,
			RoleName:  row.RoleName,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return collection.NewPage(out, page.Total, query), nil
}

func (r *MembershipRepositoryImpl) ListByUser(ctx context.Context, userID uint) ([]entity.Membership, error) {
	var rows []membershipModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT memberships.tenant_id,
		        memberships.user_id,
		        memberships.role_id,
		        memberships.role_name,
		        memberships.status,
		        memberships.created_at,
		        memberships.updated_at
		 FROM (
			SELECT tm.tenant_id,
			       tm.user_id,
			       tm.role_id,
			       r.name AS role_name,
			       tm.status,
			       tm.created_at,
			       tm.updated_at
			FROM tenant_memberships tm
			JOIN iam_roles r ON r.id = tm.role_id
			WHERE tm.user_id = $1
			UNION ALL
			SELECT t.id,
			       $1,
			       owner_role.id,
			       owner_role.name,
			       'active',
			       t.created_at,
			       t.updated_at
			FROM iam_organizations o
			JOIN tenants t ON t.org_id = o.id
			JOIN iam_roles owner_role ON owner_role.name = 'tenant_owner'
			WHERE o.root_user_id = $1
			  AND NOT EXISTS (
				SELECT 1
				FROM tenant_memberships tm
				WHERE tm.tenant_id = t.id AND tm.user_id = $1
			  )
		 ) AS memberships
		 ORDER BY memberships.created_at ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	out := make([]entity.Membership, 0, len(rows))
	for _, row := range rows {
		out = append(out, entity.Membership{
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
		return entity.ErrMembershipNotFound
	}
	return nil
}
