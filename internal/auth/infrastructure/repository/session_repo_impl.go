package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/model"
)

var (
	_ outputport.SessionRepository      = (*SessionRepositoryImpl)(nil)
	_ outputport.RefreshTokenRepository = (*RefreshTokenRepositoryImpl)(nil)
)

type SessionRepositoryImpl struct {
	db *sqlx.DB
}

func NewSessionRepositoryImpl(p UserRepoParams) *SessionRepositoryImpl {
	return &SessionRepositoryImpl{db: p.DB}
}

func (r *SessionRepositoryImpl) Create(ctx context.Context, session entity.Session) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("auth_sessions").
		Columns("id", "user_id", "active_tenant_id", "status", "created_at",
			"updated_at", "expires_at", "revoked_at").
		Values(session.ID, session.UserID, session.ActiveTenantID, session.Status, session.CreatedAt,
			session.UpdatedAt, session.ExpiresAt, session.RevokedAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *SessionRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "user_id", "active_tenant_id", "status", "created_at",
			"updated_at", "expires_at", "revoked_at").
		From("auth_sessions").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	var out model.Session
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrSessionNotFound
		}
		return nil, err
	}
	return out.ToEntity(), nil
}

func (r *SessionRepositoryImpl) ListByUser(ctx context.Context, userID uint) ([]entity.Session, error) {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "user_id", "active_tenant_id", "status", "created_at",
			"updated_at", "expires_at", "revoked_at").
		From("auth_sessions").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []model.Session
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]entity.Session, 0, len(rows))
	for _, row := range rows {
		if rowEntity := row.ToEntity(); rowEntity != nil {
			out = append(out, *rowEntity)
		}
	}
	return out, nil
}

func (r *SessionRepositoryImpl) UpdateActiveTenant(
	ctx context.Context,
	id, tenantID string,
	updatedAt time.Time,
) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("auth_sessions").
		Set("active_tenant_id", tenantID).
		Set("updated_at", updatedAt).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *SessionRepositoryImpl) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("auth_sessions").
		Set("status", entity.SessionStatusRevoked).
		Set("revoked_at", revokedAt).
		Set("updated_at", revokedAt).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

type RefreshTokenRepositoryImpl struct {
	db *sqlx.DB
}

func NewRefreshTokenRepositoryImpl(p UserRepoParams) *RefreshTokenRepositoryImpl {
	return &RefreshTokenRepositoryImpl{db: p.DB}
}

func (r *RefreshTokenRepositoryImpl) Create(ctx context.Context, token entity.RefreshToken) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("auth_refresh_tokens").
		Columns("id", "session_id", "token_hash", "expires_at", "created_at",
			"updated_at", "revoked_at", "replaced_by_token_id").
		Values(token.ID, token.SessionID, token.TokenHash, token.ExpiresAt, token.CreatedAt,
			token.UpdatedAt, token.RevokedAt, token.ReplacedByTokenID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *RefreshTokenRepositoryImpl) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*entity.RefreshToken, error) {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "session_id", "token_hash", "expires_at", "created_at",
			"updated_at", "revoked_at", "replaced_by_token_id").
		From("auth_refresh_tokens").
		Where(sq.Eq{"token_hash": tokenHash}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	var out model.RefreshToken
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrRefreshTokenInvalid
		}
		return nil, err
	}
	return out.ToEntity(), nil
}

func (r *RefreshTokenRepositoryImpl) Revoke(
	ctx context.Context,
	id string,
	revokedAt time.Time,
	replacedByTokenID *string,
) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("auth_refresh_tokens").
		Set("revoked_at", revokedAt).
		Set("updated_at", revokedAt).
		Set("replaced_by_token_id", replacedByTokenID).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *RefreshTokenRepositoryImpl) RevokeBySession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("auth_refresh_tokens").
		Set("revoked_at", revokedAt).
		Set("updated_at", revokedAt).
		Where(sq.Eq{"session_id": sessionID}).
		Where("revoked_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}
