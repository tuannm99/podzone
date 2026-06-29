package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/model"
	"github.com/tuannm99/podzone/pkg/collection"
)

var _ outputport.AuditLogRepository = (*AuditLogRepositoryImpl)(nil)

type AuditLogRepositoryImpl struct {
	db *sqlx.DB
}

func NewAuditLogRepositoryImpl(p UserRepoParams) *AuditLogRepositoryImpl {
	return &AuditLogRepositoryImpl{db: p.DB}
}

func NewIAMAuditLogRepositoryImpl(p IAMUserRepoParams) *AuditLogRepositoryImpl {
	return &AuditLogRepositoryImpl{db: p.DB}
}

func (r *AuditLogRepositoryImpl) Create(ctx context.Context, log entity.AuditLog) error {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("auth_audit_logs").
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

func (r *AuditLogRepositoryImpl) ListByActor(
	ctx context.Context,
	actorUserID uint,
	listQuery collection.Query,
) (collection.Page[entity.AuditLog], error) {
	spec := collectionSpec{
		searchColumns: []string{
			"action",
			"resource_type",
			"resource_id",
			"tenant_id",
			"status",
		},
		filterFields: map[string]collectionField{
			"action": {
				column: "action",
				operators: operators(
					collection.FilterEqual,
					collection.FilterContains,
					collection.FilterStartsWith,
					collection.FilterIn,
				),
			},
			"resource_type": {
				column:    "resource_type",
				operators: operators(collection.FilterEqual, collection.FilterIn),
			},
			"resource_id": {
				column:    "resource_id",
				operators: operators(collection.FilterEqual, collection.FilterContains),
			},
			"tenant_id": {
				column:    "tenant_id",
				operators: operators(collection.FilterEqual, collection.FilterContains, collection.FilterIn),
			},
			"status": {
				column:    "status",
				operators: operators(collection.FilterEqual, collection.FilterNotEqual, collection.FilterIn),
			},
			"created_at": {
				column: "created_at",
				operators: operators(
					collection.FilterGreaterThan,
					collection.FilterGreaterThanOrEqual,
					collection.FilterLessThan,
					collection.FilterLessThanOrEqual,
				),
			},
		},
		sortFields: map[string]string{
			"action":        "action",
			"resource_type": "resource_type",
			"status":        "status",
			"tenant_id":     "tenant_id",
			"created_at":    "created_at",
		},
		defaultSort: "created_at",
	}
	normalized, where, orderBy, err := buildCollectionQuery(listQuery, spec)
	if err != nil {
		return collection.Page[entity.AuditLog]{}, err
	}

	countBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("COUNT(*)").
		From("auth_audit_logs").
		Where(sq.Eq{"actor_user_id": actorUserID})
	for _, clause := range where {
		countBuilder = countBuilder.Where(clause)
	}
	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return collection.Page[entity.AuditLog]{}, err
	}
	var total int64
	if err := r.db.GetContext(ctx, &total, countSQL, countArgs...); err != nil {
		return collection.Page[entity.AuditLog]{}, err
	}

	queryBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "actor_user_id", "action", "resource_type", "resource_id",
			"tenant_id", "status", "payload_json", "created_at").
		From("auth_audit_logs").
		Where(sq.Eq{"actor_user_id": actorUserID}).
		OrderBy(orderBy).
		Limit(uint64(normalized.PageSize)).
		Offset(uint64(normalized.Offset()))
	for _, clause := range where {
		queryBuilder = queryBuilder.Where(clause)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return collection.Page[entity.AuditLog]{}, err
	}
	var rows []model.AuditLog
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return collection.Page[entity.AuditLog]{}, err
	}
	out := make([]entity.AuditLog, 0, len(rows))
	for _, row := range rows {
		if e := row.ToEntity(); e != nil {
			out = append(out, *e)
		}
	}
	return collection.NewPage(out, total, normalized), nil
}
