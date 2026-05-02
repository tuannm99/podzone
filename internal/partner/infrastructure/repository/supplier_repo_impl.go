package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	supplierdomain "github.com/tuannm99/podzone/internal/partner/domain"
)

type repoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-partner"`
}

type SupplierRepositoryImpl struct {
	db *sqlx.DB
}

func NewSupplierRepository(p repoParams) supplierdomain.SupplierRepository {
	return &SupplierRepositoryImpl{db: p.DB}
}

func (r *SupplierRepositoryImpl) Create(
	ctx context.Context,
	supplier supplierdomain.Supplier,
) (*supplierdomain.Supplier, error) {
	query, args, err := sq.Insert("partners").
		Columns("id", "tenant_id", "code", "name", "contact_name", "contact_email", "notes", "partner_type", "status", "created_at", "updated_at").
		Values(
			supplier.ID,
			supplier.TenantID,
			supplier.Code,
			supplier.Name,
			supplier.ContactName,
			supplier.ContactEmail,
			supplier.Notes,
			supplier.PartnerType,
			supplier.Status,
			supplier.CreatedAt,
			supplier.UpdatedAt,
		).
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var out supplierModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, supplierdomain.ErrSupplierCodeTaken
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *SupplierRepositoryImpl) GetByID(ctx context.Context, id string) (*supplierdomain.Supplier, error) {
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

	var out supplierModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, supplierdomain.ErrSupplierNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *SupplierRepositoryImpl) List(
	ctx context.Context,
	queryArg supplierdomain.ListSuppliersQuery,
) ([]supplierdomain.Supplier, error) {
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

	var rows []supplierModel
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	out := make([]supplierdomain.Supplier, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row.toEntity())
	}
	return out, nil
}

func (r *SupplierRepositoryImpl) Update(
	ctx context.Context,
	supplier supplierdomain.Supplier,
) (*supplierdomain.Supplier, error) {
	query, args, err := sq.Update("partners").
		Set("name", supplier.Name).
		Set("contact_name", supplier.ContactName).
		Set("contact_email", supplier.ContactEmail).
		Set("notes", supplier.Notes).
		Set("partner_type", supplier.PartnerType).
		Set("updated_at", supplier.UpdatedAt).
		Where(sq.Eq{"id": supplier.ID}).
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var out supplierModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, supplierdomain.ErrSupplierNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}

func (r *SupplierRepositoryImpl) UpdateStatus(
	ctx context.Context,
	id, status string,
) (*supplierdomain.Supplier, error) {
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

	var out supplierModel
	if err := r.db.GetContext(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, supplierdomain.ErrSupplierNotFound
		}
		return nil, err
	}
	return out.toEntity(), nil
}
