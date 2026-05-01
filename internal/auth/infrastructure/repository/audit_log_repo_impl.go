package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/model"
)

var _ outputport.AuditLogRepository = (*AuditLogRepositoryImpl)(nil)

type AuditLogRepositoryImpl struct {
	db *sqlx.DB
}

func NewAuditLogRepositoryImpl(p UserRepoParams) *AuditLogRepositoryImpl {
	return &AuditLogRepositoryImpl{db: p.DB}
}

func (r *AuditLogRepositoryImpl) Create(ctx context.Context, log entity.AuditLog) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("auth_audit_logs").
		Columns("id", "actor_user_id", "action", "resource_type", "resource_id", "tenant_id", "status", "payload_json", "created_at").
		Values(log.ID, log.ActorUserID, log.Action, log.ResourceType, log.ResourceID, log.TenantID, log.Status, log.PayloadJSON, log.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *AuditLogRepositoryImpl) ListByActor(ctx context.Context, actorUserID uint, limit int) ([]entity.AuditLog, error) {
	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "actor_user_id", "action", "resource_type", "resource_id", "tenant_id", "status", "payload_json", "created_at").
		From("auth_audit_logs").
		Where(sq.Eq{"actor_user_id": actorUserID}).
		OrderBy("created_at DESC")
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}
	var rows []model.AuditLog
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]entity.AuditLog, 0, len(rows))
	for _, row := range rows {
		if e := row.ToEntity(); e != nil {
			out = append(out, *e)
		}
	}
	return out, nil
}
