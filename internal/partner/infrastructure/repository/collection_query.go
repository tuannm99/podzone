package repository

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
	"github.com/tuannm99/podzone/pkg/collection"
)

var partnerColumns = map[string]string{
	"id":                  "id",
	"code":                "code",
	"name":                "name",
	"contactName":         "contact_name",
	"contactEmail":        "contact_email",
	"notes":               "notes",
	"partnerType":         "partner_type",
	"status":              "status",
	"slaDays":             "sla_days",
	"routingPriority":     "routing_priority",
	"baseFulfillmentCost": "base_fulfillment_cost",
	"createdAt":           "created_at",
	"updatedAt":           "updated_at",
}

func buildPartnerCollectionQuery(
	query partnerdomain.ListPartnersQuery,
) (collection.Query, []sq.Sqlizer, string, error) {
	normalized := query.Collection.Normalize()
	where := []sq.Sqlizer{sq.Eq{"tenant_id": query.TenantID}}
	if query.Status != "" {
		where = append(where, sq.Eq{"status": query.Status})
	}
	if query.PartnerType != "" {
		where = append(where, sq.Eq{"partner_type": query.PartnerType})
	}
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := "%" + escapePartnerLike(search) + "%"
		where = append(where, sq.Or{
			likePartnerColumn("code", pattern),
			likePartnerColumn("name", pattern),
			likePartnerColumn("contact_name", pattern),
			likePartnerColumn("contact_email", pattern),
			likePartnerColumn("notes", pattern),
			likePartnerColumn("partner_type", pattern),
			likePartnerColumn("status", pattern),
		})
	}
	for _, filter := range normalized.Filters {
		column, ok := partnerColumns[filter.Field]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported partner filter field %q",
				collection.ErrInvalidQuery,
				filter.Field,
			)
		}
		clause, err := partnerFilterClause(column, filter)
		if err != nil {
			return collection.Query{}, nil, "", err
		}
		where = append(where, clause)
	}
	sortColumn := "routing_priority"
	if normalized.SortBy != "" {
		var ok bool
		sortColumn, ok = partnerColumns[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported partner sort field %q",
				collection.ErrInvalidQuery,
				normalized.SortBy,
			)
		}
	}
	direction := "DESC"
	if normalized.SortDirection == collection.SortAscending {
		direction = "ASC"
	}
	return normalized, where, sortColumn + " " + direction, nil
}

func partnerFilterClause(column string, filter collection.Filter) (sq.Sqlizer, error) {
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: partner filter %q requires a value",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	switch filter.Operator {
	case collection.FilterEqual:
		return sq.Eq{column: filter.Values[0]}, nil
	case collection.FilterNotEqual:
		return sq.NotEq{column: filter.Values[0]}, nil
	case collection.FilterContains:
		return likePartnerColumn(column, "%"+escapePartnerLike(filter.Values[0])+"%"), nil
	case collection.FilterStartsWith:
		return likePartnerColumn(column, escapePartnerLike(filter.Values[0])+"%"), nil
	case collection.FilterGreaterThan:
		return sq.Gt{column: filter.Values[0]}, nil
	case collection.FilterGreaterThanOrEqual:
		return sq.GtOrEq{column: filter.Values[0]}, nil
	case collection.FilterLessThan:
		return sq.Lt{column: filter.Values[0]}, nil
	case collection.FilterLessThanOrEqual:
		return sq.LtOrEq{column: filter.Values[0]}, nil
	case collection.FilterIn:
		return sq.Eq{column: filter.Values}, nil
	default:
		return nil, fmt.Errorf(
			"%w: unsupported partner filter operator %q",
			collection.ErrInvalidQuery,
			filter.Operator,
		)
	}
}

func likePartnerColumn(column, pattern string) sq.Sqlizer {
	return sq.Expr("LOWER("+column+") LIKE LOWER(?) ESCAPE '\\'", pattern)
}

func escapePartnerLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
