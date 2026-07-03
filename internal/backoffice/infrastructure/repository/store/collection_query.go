package store

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/tuannm99/podzone/pkg/collection"
)

type storeCollectionColumn struct {
	sql  string
	text bool
}

var storeCollectionColumns = map[string]storeCollectionColumn{
	"id":          {sql: "id", text: true},
	"name":        {sql: "name", text: true},
	"description": {sql: "description", text: true},
	"ownerId":     {sql: "owner_id", text: true},
	"status":      {sql: "status", text: true},
	"createdAt":   {sql: "created_at"},
	"updatedAt":   {sql: "updated_at"},
}

func buildStoreCollectionQuery(
	query collection.Query,
) (collection.Query, []sq.Sqlizer, string, error) {
	normalized := query.Normalize()
	predicates := make([]sq.Sqlizer, 0, len(normalized.Filters)+1)
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := "%" + escapeStoreLike(search) + "%"
		predicates = append(predicates, sq.Or{
			storeLike("id", pattern),
			storeLike("name", pattern),
			storeLike("description", pattern),
			storeLike("status", pattern),
		})
	}
	for _, filter := range normalized.Filters {
		column, ok := storeCollectionColumns[filter.Field]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported store filter field %q",
				collection.ErrInvalidQuery,
				filter.Field,
			)
		}
		predicate, err := storeFilter(column, filter)
		if err != nil {
			return collection.Query{}, nil, "", err
		}
		predicates = append(predicates, predicate)
	}
	sortColumn := storeCollectionColumns["createdAt"].sql
	if normalized.SortBy != "" {
		column, ok := storeCollectionColumns[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported store sort field %q",
				collection.ErrInvalidQuery,
				normalized.SortBy,
			)
		}
		sortColumn = column.sql
	}
	direction := "DESC"
	if normalized.SortDirection == collection.SortAscending {
		direction = "ASC"
	}
	return normalized, predicates, sortColumn + " " + direction, nil
}

func storeFilter(column storeCollectionColumn, filter collection.Filter) (sq.Sqlizer, error) {
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf("%w: store filter %q requires a value", collection.ErrInvalidQuery, filter.Field)
	}
	switch filter.Operator {
	case collection.FilterEqual:
		return sq.Eq{column.sql: filter.Values[0]}, nil
	case collection.FilterNotEqual:
		return sq.NotEq{column.sql: filter.Values[0]}, nil
	case collection.FilterContains:
		if !column.text {
			return nil, unsupportedStoreFilter(filter)
		}
		return storeLike(column.sql, "%"+escapeStoreLike(filter.Values[0])+"%"), nil
	case collection.FilterStartsWith:
		if !column.text {
			return nil, unsupportedStoreFilter(filter)
		}
		return storeLike(column.sql, escapeStoreLike(filter.Values[0])+"%"), nil
	case collection.FilterGreaterThan:
		return sq.Gt{column.sql: filter.Values[0]}, nil
	case collection.FilterGreaterThanOrEqual:
		return sq.GtOrEq{column.sql: filter.Values[0]}, nil
	case collection.FilterLessThan:
		return sq.Lt{column.sql: filter.Values[0]}, nil
	case collection.FilterLessThanOrEqual:
		return sq.LtOrEq{column.sql: filter.Values[0]}, nil
	case collection.FilterIn:
		return sq.Eq{column.sql: filter.Values}, nil
	default:
		return nil, unsupportedStoreFilter(filter)
	}
}

func unsupportedStoreFilter(filter collection.Filter) error {
	return fmt.Errorf(
		"%w: unsupported operator %q for store field %q",
		collection.ErrInvalidQuery,
		filter.Operator,
		filter.Field,
	)
}

func storeLike(column string, pattern string) sq.Sqlizer {
	return sq.Expr("LOWER("+column+") LIKE LOWER(?) ESCAPE '\\'", pattern)
}

func escapeStoreLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
