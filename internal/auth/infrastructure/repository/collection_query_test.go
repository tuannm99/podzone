package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildCollectionQuery(t *testing.T) {
	t.Parallel()

	spec := collectionSpec{
		searchColumns: []string{"id", "status"},
		filterFields: map[string]collectionField{
			"status": {
				column: "status",
				operators: operators(
					collection.FilterEqual,
					collection.FilterIn,
				),
			},
		},
		sortFields: map[string]string{
			"created_at": "created_at",
			"status":     "status",
		},
		defaultSort: "created_at",
	}

	tests := []struct {
		name      string
		query     collection.Query
		wantOrder string
		wantWhere int
		wantError bool
	}{
		{
			name: "builds search filter and ascending sort",
			query: collection.Query{
				Page:     2,
				PageSize: 10,
				Search:   "tenant_%",
				Filters: []collection.Filter{
					{
						Field:    "status",
						Operator: collection.FilterEqual,
						Values:   []string{"active"},
					},
				},
				SortBy:        "status",
				SortDirection: collection.SortAscending,
			},
			wantOrder: "status ASC",
			wantWhere: 2,
		},
		{
			name: "rejects unknown sort field",
			query: collection.Query{
				SortBy: "private_column",
			},
			wantError: true,
		},
		{
			name: "rejects unknown filter field",
			query: collection.Query{
				Filters: []collection.Filter{
					{
						Field:    "private_column",
						Operator: collection.FilterEqual,
						Values:   []string{"value"},
					},
				},
			},
			wantError: true,
		},
		{
			name: "rejects unsupported operator",
			query: collection.Query{
				Filters: []collection.Filter{
					{
						Field:    "status",
						Operator: collection.FilterContains,
						Values:   []string{"active"},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			normalized, where, orderBy, err := buildCollectionQuery(tt.query, spec)
			if tt.wantError {
				require.Error(t, err)
				assert.ErrorIs(t, err, collection.ErrInvalidQuery)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.query.Page, normalized.Page)
			assert.Len(t, where, tt.wantWhere)
			assert.Equal(t, tt.wantOrder, orderBy)
		})
	}
}

func TestEscapeLike(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `tenant\_\%\\`, escapeLike(`tenant_%\`))
}
