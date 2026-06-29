package repository

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/tuannm99/podzone/pkg/collection"
)

type collectionField struct {
	column    string
	operators map[collection.FilterOperator]struct{}
}

type collectionSpec struct {
	searchColumns []string
	filterFields  map[string]collectionField
	sortFields    map[string]string
	defaultSort   string
}

func buildCollectionQuery(
	query collection.Query,
	spec collectionSpec,
) (collection.Query, []sq.Sqlizer, string, error) {
	normalized := query.Normalize()
	where := make([]sq.Sqlizer, 0, len(normalized.Filters)+1)
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := "%" + escapeLike(search) + "%"
		searchClauses := make(sq.Or, 0, len(spec.searchColumns))
		for _, column := range spec.searchColumns {
			searchClauses = append(
				searchClauses,
				sq.Expr("LOWER("+column+") LIKE LOWER(?) ESCAPE '\\'", pattern),
			)
		}
		where = append(where, searchClauses)
	}
	for _, filter := range normalized.Filters {
		clause, err := buildFilterClause(filter, spec.filterFields)
		if err != nil {
			return collection.Query{}, nil, "", err
		}
		where = append(where, clause)
	}

	sortColumn := spec.defaultSort
	if normalized.SortBy != "" {
		var ok bool
		sortColumn, ok = spec.sortFields[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported sort field %q",
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

func buildFilterClause(
	filter collection.Filter,
	fields map[string]collectionField,
) (sq.Sqlizer, error) {
	field, ok := fields[filter.Field]
	if !ok {
		return nil, fmt.Errorf(
			"%w: unsupported filter field %q",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	if _, ok := field.operators[filter.Operator]; !ok {
		return nil, fmt.Errorf(
			"%w: unsupported operator %q for field %q",
			collection.ErrInvalidQuery,
			filter.Operator,
			filter.Field,
		)
	}
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: filter %q requires a value",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}

	switch filter.Operator {
	case collection.FilterEqual:
		return sq.Eq{field.column: filter.Values[0]}, nil
	case collection.FilterNotEqual:
		return sq.NotEq{field.column: filter.Values[0]}, nil
	case collection.FilterContains:
		return sq.Expr(
			"LOWER("+field.column+") LIKE LOWER(?) ESCAPE '\\'",
			"%"+escapeLike(filter.Values[0])+"%",
		), nil
	case collection.FilterStartsWith:
		return sq.Expr(
			"LOWER("+field.column+") LIKE LOWER(?) ESCAPE '\\'",
			escapeLike(filter.Values[0])+"%",
		), nil
	case collection.FilterGreaterThan:
		return sq.Gt{field.column: filter.Values[0]}, nil
	case collection.FilterGreaterThanOrEqual:
		return sq.GtOrEq{field.column: filter.Values[0]}, nil
	case collection.FilterLessThan:
		return sq.Lt{field.column: filter.Values[0]}, nil
	case collection.FilterLessThanOrEqual:
		return sq.LtOrEq{field.column: filter.Values[0]}, nil
	case collection.FilterIn:
		return sq.Eq{field.column: filter.Values}, nil
	default:
		return nil, fmt.Errorf(
			"%w: unsupported filter operator %q",
			collection.ErrInvalidQuery,
			filter.Operator,
		)
	}
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	)
	return replacer.Replace(value)
}

func operators(values ...collection.FilterOperator) map[collection.FilterOperator]struct{} {
	out := make(map[collection.FilterOperator]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}
