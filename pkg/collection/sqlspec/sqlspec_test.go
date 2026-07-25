package sqlspec

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/collection"
)

var testSpec = Spec{
	Entity:      "widget",
	DefaultSort: "createdAt",
	SearchOrder: []string{"name"},
	Fields: map[string]Field{
		"name":      {Column: "name", Kind: KindText},
		"createdAt": {Column: "created_at"},
	},
}

func TestCompileSearchAndFilter(t *testing.T) {
	t.Parallel()

	_, predicates, orderBy, err := testSpec.Compile(collection.Query{
		Search: "acme",
		Filters: []collection.Filter{
			{Field: "name", Operator: collection.FilterEqual, Values: []string{"widget-1"}},
		},
		SortBy:        "name",
		SortDirection: collection.SortAscending,
	})

	require.NoError(t, err)
	assert.Len(t, predicates, 2) // search OR + filter
	assert.Equal(t, "name ASC", orderBy)
}

func TestCompileDefaultSortWhenSortByEmpty(t *testing.T) {
	t.Parallel()

	_, _, orderBy, err := testSpec.Compile(collection.Query{})

	require.NoError(t, err)
	assert.Equal(t, "created_at DESC", orderBy)
}

func TestCompileRejectsContainsOnNonTextField(t *testing.T) {
	t.Parallel()

	_, _, _, err := testSpec.Compile(collection.Query{
		Filters: []collection.Filter{
			{Field: "createdAt", Operator: collection.FilterContains, Values: []string{"2026"}},
		},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestCompileRejectsUnknownFilterField(t *testing.T) {
	t.Parallel()

	_, _, _, err := testSpec.Compile(collection.Query{
		Filters: []collection.Filter{{Field: "nope", Operator: collection.FilterEqual, Values: []string{"x"}}},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestCompileRejectsUnknownSortField(t *testing.T) {
	t.Parallel()

	_, _, _, err := testSpec.Compile(collection.Query{SortBy: "nope"})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestCompilePrependsBasePredicates(t *testing.T) {
	t.Parallel()

	_, predicates, _, err := testSpec.Compile(collection.Query{}, sq.Eq{"tenant_id": "t-1"})

	require.NoError(t, err)
	require.Len(t, predicates, 1)
	assert.Equal(t, sq.Eq{"tenant_id": "t-1"}, predicates[0])
}

func TestCompileCustomFilterNilClauseAddsNoPredicate(t *testing.T) {
	t.Parallel()

	spec := testSpec
	spec.CustomFilters = map[string]func(collection.Filter) (sq.Sqlizer, error){
		"scope": func(filter collection.Filter) (sq.Sqlizer, error) {
			switch filter.Values[0] {
			case "all":
				return nil, nil
			case "bogus":
				return nil, collection.ErrInvalidQuery
			default:
				return sq.Eq{"scope": filter.Values[0]}, nil
			}
		},
	}

	_, predicates, _, err := spec.Compile(collection.Query{
		Filters: []collection.Filter{{Field: "scope", Operator: collection.FilterEqual, Values: []string{"all"}}},
	})
	require.NoError(t, err)
	assert.Empty(t, predicates)

	_, predicates, _, err = spec.Compile(collection.Query{
		Filters: []collection.Filter{{Field: "scope", Operator: collection.FilterEqual, Values: []string{"mine"}}},
	})
	require.NoError(t, err)
	assert.Len(t, predicates, 1)
}

func TestCompileCustomSort(t *testing.T) {
	t.Parallel()

	spec := testSpec
	spec.CustomSorts = map[string]string{"priority": "CASE WHEN name = 'x' THEN 0 ELSE 1 END"}

	_, _, orderBy, err := spec.Compile(collection.Query{SortBy: "priority"})
	require.NoError(t, err)
	assert.Contains(t, orderBy, "CASE WHEN name")
}
