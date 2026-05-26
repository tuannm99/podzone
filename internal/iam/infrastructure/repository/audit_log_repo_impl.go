package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/iam/entity"
	"github.com/tuannm99/podzone/internal/iam/outputport"
	"go.uber.org/fx"
)

var _ outputport.AuditLogRepository = (*AuditLogRepositoryImpl)(nil)

type AuditLogRepoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-iam"`
}

type AuditLogRepositoryImpl struct {
	db *sqlx.DB
}

func NewAuditLogRepository(p AuditLogRepoParams) *AuditLogRepositoryImpl {
	return &AuditLogRepositoryImpl{db: p.DB}
}

func (r *AuditLogRepositoryImpl) Create(ctx context.Context, log entity.AuditLog) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("iam_audit_logs").
		Columns("id", "actor_user_id", "action", "resource_type", "resource_id",
			"tenant_id", "status", "payload_json", "created_at").
		Values(log.ID, log.ActorUserID, log.Action, log.ResourceType, log.ResourceID,
			log.TenantID, log.Status, log.PayloadJSON, log.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}
