// Package sqlspec compiles a declarative, per-entity field spec into the
// (normalized query, WHERE predicates, ORDER BY expression) shape every
// collection repository method in this codebase builds by hand today. See
// docs/07-problems/repository-query-layer-2026-07-24.md RQ2: the store and
// routing repositories' collection_query.go files were ~90% identical, and
// the one place they diverged (routing never guarded FilterContains/
// FilterStartsWith against non-text columns) was a real bug - see RQ1.
package sqlspec

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/tuannm99/podzone/pkg/collection"
)

// FieldKind gates which filter operators a field accepts. Kept minimal -
// today's repositories only ever needed to distinguish "text" (accepts
// contains/starts_with) from everything else.
type FieldKind int

const (
	KindOther FieldKind = iota
	KindText
)

type Field struct {
	Column string
	Kind   FieldKind
}

// Spec is defined once per entity, as a package-level var - see
// docs/08-adr/ADR-0008-frontend-v2-declarative-collection-field-spec.md for
// the frontend-side counterpart of this same "declare once" instinct.
type Spec struct {
	// Entity names the spec in error messages (e.g. "store", "routed order").
	Entity string
	// DefaultSort is the field key used when the query has no SortBy.
	DefaultSort string
	// SearchOrder lists the field keys included in the free-text search
	// OR-clause, in order. Fields not listed here are filterable/sortable
	// but not part of free-text search - matches the two current
	// repositories, which both search a narrower set of columns than they
	// accept as explicit filters.
	SearchOrder []string
	Fields      map[string]Field

	// CustomFilters handles field keys that aren't a plain column
	// comparison (e.g. routing's "queueView", which maps a preset name to
	// a hand-written predicate). Returning (nil, nil) means "match
	// everything, add no predicate" - preserves the field's `all`/no-op
	// values without a special case in Compile itself.
	CustomFilters map[string]func(filter collection.Filter) (sq.Sqlizer, error)
	// CustomSorts handles SortBy/DefaultSort keys that need a raw SQL
	// expression instead of a plain column name (e.g. routing's "priority"
	// CASE expression).
	CustomSorts map[string]string
}

// Compile mirrors the previous buildXCollectionQuery(query) functions'
// return shape exactly, so callers are a drop-in swap. base is prepended to
// the predicate list unconditionally (e.g. a tenant/store-scope
// sq.Eq{"store_id": storeID}) - it is the caller's responsibility, not part
// of the reusable Spec, because it is per-call data, not per-entity shape.
func (s Spec) Compile(query collection.Query, base ...sq.Sqlizer) (collection.Query, []sq.Sqlizer, string, error) {
	normalized := query.Normalize()
	predicates := make([]sq.Sqlizer, 0, len(base)+len(normalized.Filters)+1)
	predicates = append(predicates, base...)

	if search := strings.TrimSpace(normalized.Search); search != "" && len(s.SearchOrder) > 0 {
		pattern := "%" + escapeLike(search) + "%"
		ors := make([]sq.Sqlizer, 0, len(s.SearchOrder))
		for _, key := range s.SearchOrder {
			field, ok := s.Fields[key]
			if !ok {
				continue
			}
			ors = append(ors, likeExpr(field.Column, pattern))
		}
		if len(ors) > 0 {
			predicates = append(predicates, sq.Or(ors))
		}
	}

	for _, filter := range normalized.Filters {
		if handler, ok := s.CustomFilters[filter.Field]; ok {
			clause, err := handler(filter)
			if err != nil {
				return collection.Query{}, nil, "", err
			}
			if clause != nil {
				predicates = append(predicates, clause)
			}
			continue
		}
		field, ok := s.Fields[filter.Field]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported %s filter field %q",
				collection.ErrInvalidQuery, s.Entity, filter.Field,
			)
		}
		clause, err := s.filterClause(field, filter)
		if err != nil {
			return collection.Query{}, nil, "", err
		}
		predicates = append(predicates, clause)
	}

	sortKey := normalized.SortBy
	if sortKey == "" {
		sortKey = s.DefaultSort
	}
	sortExpr, ok := s.CustomSorts[sortKey]
	if !ok {
		field, fieldOK := s.Fields[sortKey]
		if !fieldOK {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported %s sort field %q",
				collection.ErrInvalidQuery, s.Entity, sortKey,
			)
		}
		sortExpr = field.Column
	}

	direction := "DESC"
	if normalized.SortDirection == collection.SortAscending {
		direction = "ASC"
	}
	return normalized, predicates, sortExpr + " " + direction, nil
}

func (s Spec) filterClause(field Field, filter collection.Filter) (sq.Sqlizer, error) {
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: %s filter %q requires a value",
			collection.ErrInvalidQuery, s.Entity, filter.Field,
		)
	}
	switch filter.Operator {
	case collection.FilterEqual:
		return sq.Eq{field.Column: filter.Values[0]}, nil
	case collection.FilterNotEqual:
		return sq.NotEq{field.Column: filter.Values[0]}, nil
	case collection.FilterContains:
		if field.Kind != KindText {
			return nil, s.unsupportedOperator(filter)
		}
		return likeExpr(field.Column, "%"+escapeLike(filter.Values[0])+"%"), nil
	case collection.FilterStartsWith:
		if field.Kind != KindText {
			return nil, s.unsupportedOperator(filter)
		}
		return likeExpr(field.Column, escapeLike(filter.Values[0])+"%"), nil
	case collection.FilterGreaterThan:
		return sq.Gt{field.Column: filter.Values[0]}, nil
	case collection.FilterGreaterThanOrEqual:
		return sq.GtOrEq{field.Column: filter.Values[0]}, nil
	case collection.FilterLessThan:
		return sq.Lt{field.Column: filter.Values[0]}, nil
	case collection.FilterLessThanOrEqual:
		return sq.LtOrEq{field.Column: filter.Values[0]}, nil
	case collection.FilterIn:
		return sq.Eq{field.Column: filter.Values}, nil
	default:
		return nil, s.unsupportedOperator(filter)
	}
}

func (s Spec) unsupportedOperator(filter collection.Filter) error {
	return fmt.Errorf(
		"%w: unsupported operator %q for %s field %q",
		collection.ErrInvalidQuery, filter.Operator, s.Entity, filter.Field,
	)
}

func likeExpr(column string, pattern string) sq.Sqlizer {
	return sq.Expr("LOWER("+column+") LIKE LOWER(?) ESCAPE '\\'", pattern)
}

func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
