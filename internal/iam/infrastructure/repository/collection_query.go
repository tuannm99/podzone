package repository

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/tuannm99/podzone/pkg/collection"
)

type collectionColumn struct {
	sql  string
	text bool
}

var organizationCollectionColumns = map[string]collectionColumn{
	"id":        {sql: "id", text: true},
	"slug":      {sql: "slug", text: true},
	"name":      {sql: "name", text: true},
	"createdAt": {sql: "created_at"},
	"updatedAt": {sql: "updated_at"},
}

var policyCollectionColumns = map[string]collectionColumn{
	"id":             {sql: "id"},
	"scope":          {sql: "scope", text: true},
	"name":           {sql: "name", text: true},
	"description":    {sql: "description", text: true},
	"isSystem":       {sql: "is_system"},
	"defaultVersion": {sql: "default_version", text: true},
	"createdAt":      {sql: "created_at"},
	"updatedAt":      {sql: "updated_at"},
}

var groupCollectionColumns = map[string]collectionColumn{
	"id":          {sql: "id"},
	"scope":       {sql: "scope", text: true},
	"tenantId":    {sql: "tenant_id", text: true},
	"name":        {sql: "name", text: true},
	"description": {sql: "description", text: true},
	"isSystem":    {sql: "is_system"},
	"createdAt":   {sql: "created_at"},
	"updatedAt":   {sql: "updated_at"},
}

func buildIAMCollectionQuery(
	query collection.Query,
	columns map[string]collectionColumn,
	searchColumns []string,
	defaultSort string,
) (collection.Query, []sq.Sqlizer, string, error) {
	normalized := query.Normalize()
	where := make([]sq.Sqlizer, 0, len(normalized.Filters)+1)
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := "%" + escapeIAMLike(search) + "%"
		searchClauses := make(sq.Or, 0, len(searchColumns))
		for _, column := range searchColumns {
			searchClauses = append(searchClauses, likeIAMColumn(column, pattern))
		}
		where = append(where, searchClauses)
	}
	for _, filter := range normalized.Filters {
		column, ok := columns[filter.Field]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported IAM filter field %q",
				collection.ErrInvalidQuery,
				filter.Field,
			)
		}
		clause, err := iamFilterClause(column, filter)
		if err != nil {
			return collection.Query{}, nil, "", err
		}
		where = append(where, clause)
	}
	sortColumn := defaultSort
	if normalized.SortBy != "" {
		column, ok := columns[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported IAM sort field %q",
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
	return normalized, where, sortColumn + " " + direction, nil
}

func iamFilterClause(column collectionColumn, filter collection.Filter) (sq.Sqlizer, error) {
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: IAM filter %q requires a value",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	switch filter.Operator {
	case collection.FilterEqual:
		return sq.Eq{column.sql: filter.Values[0]}, nil
	case collection.FilterNotEqual:
		return sq.NotEq{column.sql: filter.Values[0]}, nil
	case collection.FilterContains:
		if !column.text {
			return nil, unsupportedIAMFilterOperator(filter)
		}
		return likeIAMColumn(column.sql, "%"+escapeIAMLike(filter.Values[0])+"%"), nil
	case collection.FilterStartsWith:
		if !column.text {
			return nil, unsupportedIAMFilterOperator(filter)
		}
		return likeIAMColumn(column.sql, escapeIAMLike(filter.Values[0])+"%"), nil
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
		return nil, unsupportedIAMFilterOperator(filter)
	}
}

func unsupportedIAMFilterOperator(filter collection.Filter) error {
	return fmt.Errorf(
		"%w: unsupported IAM filter operator %q for field %q",
		collection.ErrInvalidQuery,
		filter.Operator,
		filter.Field,
	)
}

func likeIAMColumn(column string, pattern string) sq.Sqlizer {
	return sq.Expr("LOWER("+column+") LIKE LOWER(?) ESCAPE '\\'", pattern)
}

func escapeIAMLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
