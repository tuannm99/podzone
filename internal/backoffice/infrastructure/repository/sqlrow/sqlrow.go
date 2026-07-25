// Package sqlrow derives SELECT column lists from a row struct's `db` tags
// instead of hand-writing them at each call site. See docs/07-problems/
// repository-query-layer-2026-07-24.md RQ3: a struct's column list was
// previously written out by hand up to 4 times per row type across
// store/catalog/routing, with nothing keeping the copies and the struct in
// sync.
package sqlrow

import (
	"reflect"
	"sync"

	sq "github.com/Masterminds/squirrel"
)

var columnsCache sync.Map // reflect.Type -> []string

// ColumnsOf returns T's `db`-tagged struct fields, in declaration order.
// Cached by reflect.Type after the first call per T.
func ColumnsOf[T any]() []string {
	t := reflect.TypeFor[T]()
	if cached, ok := columnsCache.Load(t); ok {
		return cached.([]string)
	}

	columns := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		tag := t.Field(i).Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		columns = append(columns, tag)
	}

	columnsCache.Store(t, columns)
	return columns
}

// SelectFrom builds `psql.Select(ColumnsOf[Row]()...).From(table)` - the
// shape every plain "fetch the whole row" query in this repository layer
// needs. Queries that only need a subset of columns, joins, or aggregates
// still build their sq.SelectBuilder directly; this only replaces the
// full-row case.
func SelectFrom[Row any](psql sq.StatementBuilderType, table string) sq.SelectBuilder {
	return psql.Select(ColumnsOf[Row]()...).From(table)
}
