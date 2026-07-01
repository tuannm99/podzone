package repository

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildIAMCollectionQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		query       collection.Query
		columns     map[string]collectionColumn
		search      []string
		defaultSort string
		wantOrder   string
		wantWhere   int
	}{
		{
			name: "organization search and sort",
			query: collection.Query{
				Page:          2,
				PageSize:      10,
				Search:        `ops%_team`,
				SortBy:        "name",
				SortDirection: collection.SortAscending,
			},
			columns:     organizationCollectionColumns,
			search:      []string{"id", "slug", "name"},
			defaultSort: "created_at",
			wantOrder:   "name ASC",
			wantWhere:   1,
		},
		{
			name: "policy equality filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "scope",
					Operator: collection.FilterEqual,
					Values:   []string{"platform"},
				}},
			},
			columns:     policyCollectionColumns,
			search:      []string{"scope", "name", "description"},
			defaultSort: "created_at",
			wantOrder:   "created_at DESC",
			wantWhere:   1,
		},
		{
			name: "group name contains filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "name",
					Operator: collection.FilterContains,
					Values:   []string{"operator"},
				}},
			},
			columns:     groupCollectionColumns,
			search:      []string{"scope", "name", "description"},
			defaultSort: "created_at",
			wantOrder:   "created_at DESC",
			wantWhere:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			normalized, where, orderBy, err := buildIAMCollectionQuery(
				tt.query,
				tt.columns,
				tt.search,
				tt.defaultSort,
			)
			require.NoError(t, err)
			assert.Equal(t, tt.wantOrder, orderBy)
			assert.Len(t, where, tt.wantWhere)
			assert.GreaterOrEqual(t, normalized.Page, collection.DefaultPage)

			builder := sq.Select("*").From("items")
			for _, predicate := range where {
				builder = builder.Where(predicate)
			}
			query, args, sqlErr := builder.PlaceholderFormat(sq.Dollar).ToSql()
			require.NoError(t, sqlErr)
			assert.NotEmpty(t, query)
			if tt.query.Search != "" {
				assert.NotContains(t, query, tt.query.Search)
				require.NotEmpty(t, args)
				assert.Equal(t, `%ops\%\_team%`, args[0])
			}
		})
	}
}

func TestBuildIAMCollectionQueryRejectsUnsupportedField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildIAMCollectionQuery(
		collection.Query{
			Filters: []collection.Filter{{
				Field:    "raw_sql",
				Operator: collection.FilterEqual,
				Values:   []string{"1"},
			}},
		},
		policyCollectionColumns,
		[]string{"name"},
		"created_at",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestBuildIAMCollectionQueryRejectsTextOperatorForNonTextField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildIAMCollectionQuery(
		collection.Query{
			Filters: []collection.Filter{{
				Field:    "isSystem",
				Operator: collection.FilterContains,
				Values:   []string{"true"},
			}},
		},
		policyCollectionColumns,
		[]string{"name"},
		"created_at",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, collection.ErrInvalidQuery)
}
