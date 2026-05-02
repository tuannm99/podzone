package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type InviteRepositoryImpl struct {
	db *sqlx.DB
}

func NewInviteRepository(p repoParams) iamdomain.InviteRepository {
	return &InviteRepositoryImpl{db: p.DB}
}

func (r *InviteRepositoryImpl) Create(ctx context.Context, invite iamdomain.TenantInvite) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tenant_invites
		 (id, tenant_id, email, role_id, status, invited_by_user_id, token_hash,
			created_at, updated_at, expires_at, accepted_by_user_id, accepted_at,
			revoked_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		invite.ID,
		invite.TenantID,
		invite.Email,
		invite.RoleID,
		invite.Status,
		invite.InvitedByUserID,
		invite.TokenHash,
		invite.CreatedAt,
		invite.UpdatedAt,
		invite.ExpiresAt,
		invite.AcceptedByUserID,
		invite.AcceptedAt,
		invite.RevokedAt,
	)
	return err
}

func (r *InviteRepositoryImpl) GetByID(ctx context.Context, inviteID string) (*iamdomain.TenantInvite, error) {
	var out inviteModel
	if err := r.db.GetContext(ctx, &out, `
		SELECT ti.id, ti.tenant_id, ti.email, ti.role_id, r.name AS role_name, 
			ti.status, ti.invited_by_user_id, ti.accepted_by_user_id, 
			ti.token_hash, ti.created_at, ti.updated_at, ti.expires_at, 
			ti.accepted_at, ti.revoked_at
		FROM tenant_invites ti
		JOIN iam_roles r ON r.id = ti.role_id
		WHERE ti.id = $1
	`, inviteID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrInviteNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *InviteRepositoryImpl) GetByTokenHash(ctx context.Context, tokenHash string) (*iamdomain.TenantInvite, error) {
	var out inviteModel
	if err := r.db.GetContext(ctx, &out, `
		SELECT ti.id, ti.tenant_id, ti.email, ti.role_id, r.name AS role_name, ti.status, ti.invited_by_user_id,
		       ti.accepted_by_user_id, ti.token_hash, ti.created_at, ti.updated_at, ti.expires_at, ti.accepted_at, ti.revoked_at
		FROM tenant_invites ti
		JOIN iam_roles r ON r.id = ti.role_id
		WHERE ti.token_hash = $1
	`, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, iamdomain.ErrInviteNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *InviteRepositoryImpl) ListByTenant(ctx context.Context, tenantID string) ([]iamdomain.TenantInvite, error) {
	var rows []inviteModel
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT ti.id, ti.tenant_id, ti.email, ti.role_id, r.name AS role_name, ti.status, ti.invited_by_user_id,
		       ti.accepted_by_user_id, ti.token_hash, ti.created_at, ti.updated_at, ti.expires_at, ti.accepted_at, ti.revoked_at
		FROM tenant_invites ti
		JOIN iam_roles r ON r.id = ti.role_id
		WHERE ti.tenant_id = $1
		ORDER BY ti.created_at DESC
	`, tenantID); err != nil {
		return nil, err
	}
	out := make([]iamdomain.TenantInvite, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row.toEntity())
	}
	return out, nil
}

func (r *InviteRepositoryImpl) MarkAccepted(
	ctx context.Context,
	inviteID string,
	acceptedByUserID uint,
	acceptedAt time.Time,
) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_invites
		SET status = $2, accepted_by_user_id = $3, accepted_at = $4, updated_at = $4
		WHERE id = $1
	`, inviteID, iamdomain.InviteStatusAccepted, acceptedByUserID, acceptedAt)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return iamdomain.ErrInviteNotFound
	}
	return nil
}

func (r *InviteRepositoryImpl) MarkRevoked(ctx context.Context, inviteID string, revokedAt time.Time) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_invites
		SET status = $2, revoked_at = $3, updated_at = $3
		WHERE id = $1
	`, inviteID, iamdomain.InviteStatusRevoked, revokedAt)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return iamdomain.ErrInviteNotFound
	}
	return nil
}
