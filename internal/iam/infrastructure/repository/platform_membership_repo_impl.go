package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type PlatformMembershipRepositoryImpl struct {
	db *sqlx.DB
}

func NewPlatformMembershipRepository(p repoParams) iamdomain.PlatformMembershipRepository {
	return &PlatformMembershipRepositoryImpl{db: p.DB}
}

func (r *PlatformMembershipRepositoryImpl) Upsert(
	ctx context.Context,
	userID uint,
	roleID uint64,
	status string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO user_platform_roles (user_id, role_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, now(), now())
		 ON CONFLICT (user_id, role_id) DO UPDATE
		 SET status = EXCLUDED.status,
		     updated_at = EXCLUDED.updated_at`,
		userID,
		roleID,
		status,
	)
	return err
}

func (r *PlatformMembershipRepositoryImpl) ListRoleIDsByUser(ctx context.Context, userID uint) ([]uint64, error) {
	var rows []platformMembershipModel
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT user_id, role_id, status, created_at, updated_at
		 FROM user_platform_roles
		 WHERE user_id = $1 AND status = 'active'
		 ORDER BY created_at ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.RoleID)
	}
	return out, nil
}

func (r *PlatformMembershipRepositoryImpl) ListByUser(
	ctx context.Context,
	userID uint,
) ([]iamdomain.PlatformMembership, error) {
	var rows []struct {
		UserID    uint      `db:"user_id"`
		RoleID    uint64    `db:"role_id"`
		RoleName  string    `db:"role_name"`
		Status    string    `db:"status"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
	if err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT upr.user_id, upr.role_id, r.name AS role_name, upr.status, upr.created_at, upr.updated_at
		 FROM user_platform_roles upr
		 JOIN iam_roles r ON r.id = upr.role_id
		 WHERE upr.user_id = $1
		 ORDER BY upr.created_at ASC`,
		userID,
	); err != nil {
		return nil, err
	}
	out := make([]iamdomain.PlatformMembership, 0, len(rows))
	for _, row := range rows {
		out = append(out, iamdomain.PlatformMembership{
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

func (r *PlatformMembershipRepositoryImpl) Delete(ctx context.Context, userID uint, roleID uint64) error {
	res, err := r.db.ExecContext(
		ctx,
		`DELETE FROM user_platform_roles WHERE user_id = $1 AND role_id = $2`,
		userID,
		roleID,
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
