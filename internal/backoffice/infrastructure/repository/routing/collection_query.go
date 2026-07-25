package routing

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/collection/sqlspec"
)

// See docs/07-problems/repository-query-layer-2026-07-24.md RQ2, which
// closes RQ1: previously FilterContains/FilterStartsWith had no text-column
// guard here (unlike store's), so e.g. `createdAt contains "2026"` built
// `LOWER(created_at) LIKE ...` against a TIMESTAMPTZ column and Postgres
// rejected it as a 500. sqlspec.Spec applies the same guard store already
// had - that request now fails validation (collection.ErrInvalidQuery)
// before a query is ever built. This is a visible API behavior change
// (500 -> 400-class validation error) - see the changelog entry and
// frontend check called out in the RQ2 write-up.
var routedOrderSpec = sqlspec.Spec{
	Entity:      "routed order",
	DefaultSort: "createdAt",
	SearchOrder: []string{"id", "productTitle", "partner", "customerName", "status", "operatorAssignee", "settlementStatus"},
	Fields: map[string]sqlspec.Field{
		"id":               {Column: "id", Kind: sqlspec.KindText},
		"productTitle":     {Column: "product_title", Kind: sqlspec.KindText},
		"partner":          {Column: "partner", Kind: sqlspec.KindText},
		"customerName":     {Column: "customer_name", Kind: sqlspec.KindText},
		"status":           {Column: "status", Kind: sqlspec.KindText},
		"exceptionType":    {Column: "exception_type", Kind: sqlspec.KindText},
		"exceptionStatus":  {Column: "exception_status", Kind: sqlspec.KindText},
		"shipmentStatus":   {Column: "shipment_status", Kind: sqlspec.KindText},
		"operatorAssignee": {Column: "operator_assignee", Kind: sqlspec.KindText},
		"settlementStatus": {Column: "settlement_status", Kind: sqlspec.KindText},
		"realizedMargin":   {Column: "realized_margin", Kind: sqlspec.KindText},
		"createdAt":        {Column: "created_at"},
		"updatedAt":        {Column: "updated_at"},
		"shipmentSlaDueAt": {Column: "shipment_sla_due_at"},
		"issueSlaDueAt":    {Column: "issue_sla_due_at"},
	},
	CustomFilters: map[string]func(filter collection.Filter) (sq.Sqlizer, error){
		"queueView": routedOrderQueueViewClause,
	},
	CustomSorts: map[string]string{
		"priority": `CASE
			WHEN exception_status = 'escalated' THEN 0
			WHEN exception_status = 'open' THEN 1
			WHEN settlement_status = 'disputed' THEN 2
			WHEN shipment_status = 'delivery_issue' THEN 3
			ELSE 4
		END`,
	},
}

func buildRoutedOrderCollectionQuery(
	storeID string,
	query collection.Query,
) (collection.Query, []sq.Sqlizer, string, error) {
	return routedOrderSpec.Compile(query, sq.Eq{"store_id": storeID})
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
