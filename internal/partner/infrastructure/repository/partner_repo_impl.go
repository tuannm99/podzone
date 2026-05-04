package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
)

type repoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-partner"`
}

type PartnerRepositoryImpl struct {
	db *sqlx.DB
}

func NewPartnerRepository(p repoParams) partnerdomain.PartnerRepository {
	return &PartnerRepositoryImpl{db: p.DB}
}

func (r *PartnerRepositoryImpl) Create(
	ctx context.Context,
	partner partnerdomain.Partner,
) (*partnerdomain.Partner, error) {
	query, args, err := sq.Insert("partners").
		Columns("id", "tenant_id", "code", "name", "contact_name", "contact_email", "notes", "partner_type", "status", "created_at", "updated_at").
		Values(
			partner.ID,
			partner.TenantID,
			partner.Code,
			partner.Name,
			partner.ContactName,
			partner.ContactEmail,
			partner.Notes,
			partner.PartnerType,
			partner.Status,
			partner.CreatedAt,
			partner.UpdatedAt,
		).
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var out partnerModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, partnerdomain.ErrPartnerCodeTaken
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *PartnerRepositoryImpl) GetByID(ctx context.Context, id string) (*partnerdomain.Partner, error) {
	query, args, err := sq.Select(
		"id",
		"tenant_id",
		"code",
		"name",
		"contact_name",
		"contact_email",
		"notes",
		"partner_type",
		"status",
		"created_at",
		"updated_at",
	).From("partners").Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	var out partnerModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, partnerdomain.ErrPartnerNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *PartnerRepositoryImpl) List(
	ctx context.Context,
	queryArg partnerdomain.ListPartnersQuery,
) ([]partnerdomain.Partner, error) {
	builder := sq.Select(
		"id",
		"tenant_id",
		"code",
		"name",
		"contact_name",
		"contact_email",
		"notes",
		"partner_type",
		"status",
		"created_at",
		"updated_at",
	).From("partners").Where(sq.Eq{"tenant_id": queryArg.TenantID}).OrderBy("created_at DESC")
	if queryArg.Status != "" {
		builder = builder.Where(sq.Eq{"status": queryArg.Status})
	}
	if queryArg.PartnerType != "" {
		builder = builder.Where(sq.Eq{"partner_type": queryArg.PartnerType})
	}
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	var rows []partnerModel
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	out := make([]partnerdomain.Partner, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row.toEntity())
	}
	return out, nil
}

func (r *PartnerRepositoryImpl) Update(
	ctx context.Context,
	partner partnerdomain.Partner,
) (*partnerdomain.Partner, error) {
	query, args, err := sq.Update("partners").
		Set("name", partner.Name).
		Set("contact_name", partner.ContactName).
		Set("contact_email", partner.ContactEmail).
		Set("notes", partner.Notes).
		Set("partner_type", partner.PartnerType).
		Set("updated_at", partner.UpdatedAt).
		Where(sq.Eq{"id": partner.ID}).
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var out partnerModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, partnerdomain.ErrPartnerNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *PartnerRepositoryImpl) UpdateStatus(
	ctx context.Context,
	id, status string,
) (*partnerdomain.Partner, error) {
	query, args, err := sq.Update("partners").
		Set("status", status).
		Set("updated_at", sq.Expr("NOW() AT TIME ZONE 'UTC'")).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var out partnerModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, partnerdomain.ErrPartnerNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}
