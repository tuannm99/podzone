package routing

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/tuannm99/podzone/pkg/collection"
)

var routedOrderCollectionColumns = map[string]string{
	"id":               "id",
	"productTitle":     "product_title",
	"partner":          "partner",
	"customerName":     "customer_name",
	"status":           "status",
	"exceptionType":    "exception_type",
	"exceptionStatus":  "exception_status",
	"shipmentStatus":   "shipment_status",
	"operatorAssignee": "operator_assignee",
	"settlementStatus": "settlement_status",
	"realizedMargin":   "realized_margin",
	"createdAt":        "created_at",
	"updatedAt":        "updated_at",
	"shipmentSlaDueAt": "shipment_sla_due_at",
	"issueSlaDueAt":    "issue_sla_due_at",
}

func buildRoutedOrderCollectionQuery(
	storeID string,
	query collection.Query,
) (collection.Query, []sq.Sqlizer, string, error) {
	normalized := query.Normalize()
	where := []sq.Sqlizer{sq.Eq{"store_id": storeID}}
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := "%" + escapeRoutedOrderLike(search) + "%"
		where = append(where, sq.Or{
			likeRoutedOrderColumn("id", pattern),
			likeRoutedOrderColumn("product_title", pattern),
			likeRoutedOrderColumn("partner", pattern),
			likeRoutedOrderColumn("customer_name", pattern),
			likeRoutedOrderColumn("status", pattern),
			likeRoutedOrderColumn("operator_assignee", pattern),
			likeRoutedOrderColumn("settlement_status", pattern),
		})
	}
	for _, filter := range normalized.Filters {
		if filter.Field == "queueView" {
			clause, err := routedOrderQueueViewClause(filter)
			if err != nil {
				return collection.Query{}, nil, "", err
			}
			if clause != nil {
				where = append(where, clause)
			}
			continue
		}
		column, ok := routedOrderCollectionColumns[filter.Field]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported routed order filter field %q",
				collection.ErrInvalidQuery,
				filter.Field,
			)
		}
		clause, err := routedOrderFilterClause(column, filter)
		if err != nil {
			return collection.Query{}, nil, "", err
		}
		where = append(where, clause)
	}
	sortColumn := "created_at"
	if normalized.SortBy == "priority" {
		sortColumn = `CASE
			WHEN exception_status = 'escalated' THEN 0
			WHEN exception_status = 'open' THEN 1
			WHEN settlement_status = 'disputed' THEN 2
			WHEN shipment_status = 'delivery_issue' THEN 3
			ELSE 4
		END`
	} else if normalized.SortBy != "" {
		var ok bool
		sortColumn, ok = routedOrderCollectionColumns[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported routed order sort field %q",
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

func routedOrderQueueViewClause(filter collection.Filter) (sq.Sqlizer, error) {
	if filter.Operator != collection.FilterEqual || len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: queueView requires one EQ value",
			collection.ErrInvalidQuery,
		)
	}
	switch filter.Values[0] {
	case "", "all", "my_queue":
		return nil, nil
	case "overdue":
		return sq.Or{
			sq.And{
				sq.Expr("shipment_sla_due_at < CURRENT_TIMESTAMP"),
				sq.NotEq{"shipment_status": "delivered"},
			},
			sq.And{
				sq.Expr("issue_sla_due_at < CURRENT_TIMESTAMP"),
				sq.Eq{"exception_status": []string{"open", "escalated"}},
			},
		}, nil
	case "delivery_issues":
		return sq.Eq{"shipment_status": "delivery_issue"}, nil
	case "settlement_pending":
		return sq.Eq{"settlement_status": "pending"}, nil
	case "finance_review":
		return sq.Or{
			sq.Eq{"settlement_status": []string{"pending", "disputed"}},
			sq.Expr("realized_margin LIKE '%-%'"),
			sq.NotEq{"issue_cost": []string{"", "$0.00", "0", "TBD"}},
		}, nil
	default:
		return nil, fmt.Errorf(
			"%w: unsupported queue view %q",
			collection.ErrInvalidQuery,
			filter.Values[0],
		)
	}
}

func routedOrderFilterClause(column string, filter collection.Filter) (sq.Sqlizer, error) {
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: routed order filter %q requires a value",
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
		return likeRoutedOrderColumn(column, "%"+escapeRoutedOrderLike(filter.Values[0])+"%"), nil
	case collection.FilterStartsWith:
		return likeRoutedOrderColumn(column, escapeRoutedOrderLike(filter.Values[0])+"%"), nil
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
			"%w: unsupported routed order filter operator %q",
			collection.ErrInvalidQuery,
			filter.Operator,
		)
	}
}

func likeRoutedOrderColumn(column, pattern string) sq.Sqlizer {
	return sq.Expr("LOWER("+column+") LIKE LOWER(?) ESCAPE '\\'", pattern)
}

func escapeRoutedOrderLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
