package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
	"github.com/tuannm99/podzone/pkg/collection"
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
	shippingRulesJSON, err := json.Marshal(partner.ShippingCostRules)
	if err != nil {
		return nil, err
	}
	query, args, err := sq.Insert("partners").
		Columns("id", "tenant_id", "code", "name", "contact_name", "contact_email", "notes", "partner_type", "status", "supported_product_types", "supported_regions", "sla_days", "routing_priority", "base_fulfillment_cost", "shipping_cost_rules_json", "created_at", "updated_at").
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
			partner.SupportedProductTypes,
			partner.SupportedRegions,
			partner.SLADays,
			partner.RoutingPriority,
			partner.BaseFulfillmentCost,
			shippingRulesJSON,
			partner.CreatedAt,
			partner.UpdatedAt,
		).
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, supported_product_types, supported_regions, sla_days, routing_priority, base_fulfillment_cost, shipping_cost_rules_json, created_at, updated_at").
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
		"supported_product_types",
		"supported_regions",
		"sla_days",
		"routing_priority",
		"base_fulfillment_cost",
		"shipping_cost_rules_json",
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
) (collection.Page[partnerdomain.Partner], error) {
	normalized, where, orderBy, err := buildPartnerCollectionQuery(queryArg)
	if err != nil {
		return collection.Page[partnerdomain.Partner]{}, err
	}
	countBuilder := sq.Select("COUNT(*)").From("partners")
	for _, predicate := range where {
		countBuilder = countBuilder.Where(predicate)
	}
	countSQL, countArgs, err := countBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return collection.Page[partnerdomain.Partner]{}, err
	}
	var total int64
	if err := r.db.GetContext(ctx, &total, countSQL, countArgs...); err != nil {
		return collection.Page[partnerdomain.Partner]{}, err
	}

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
		"supported_product_types",
		"supported_regions",
		"sla_days",
		"routing_priority",
		"base_fulfillment_cost",
		"shipping_cost_rules_json",
		"created_at",
		"updated_at",
	).From("partners")
	for _, predicate := range where {
		builder = builder.Where(predicate)
	}
	builder = builder.
		OrderBy(orderBy, "created_at DESC").
		Limit(uint64(normalized.PageSize)).
		Offset(uint64(normalized.Offset()))
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return collection.Page[partnerdomain.Partner]{}, err
	}

	var rows []partnerModel
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return collection.Page[partnerdomain.Partner]{}, err
	}

	out := make([]partnerdomain.Partner, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row.toEntity())
	}
	return collection.NewPage(out, total, normalized), nil
}

func (r *PartnerRepositoryImpl) Update(
	ctx context.Context,
	partner partnerdomain.Partner,
) (*partnerdomain.Partner, error) {
	shippingRulesJSON, err := json.Marshal(partner.ShippingCostRules)
	if err != nil {
		return nil, err
	}
	query, args, err := sq.Update("partners").
		Set("name", partner.Name).
		Set("contact_name", partner.ContactName).
		Set("contact_email", partner.ContactEmail).
		Set("notes", partner.Notes).
		Set("partner_type", partner.PartnerType).
		Set("supported_product_types", partner.SupportedProductTypes).
		Set("supported_regions", partner.SupportedRegions).
		Set("sla_days", partner.SLADays).
		Set("routing_priority", partner.RoutingPriority).
		Set("base_fulfillment_cost", partner.BaseFulfillmentCost).
		Set("shipping_cost_rules_json", shippingRulesJSON).
		Set("updated_at", partner.UpdatedAt).
		Where(sq.Eq{"id": partner.ID}).
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, supported_product_types, supported_regions, sla_days, routing_priority, base_fulfillment_cost, shipping_cost_rules_json, created_at, updated_at").
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
		Suffix("RETURNING id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, supported_product_types, supported_regions, sla_days, routing_priority, base_fulfillment_cost, shipping_cost_rules_json, created_at, updated_at").
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
