package routing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildRoutedOrderCollectionQuery(t *testing.T) {
	t.Parallel()

	query, predicates, orderBy, err := buildRoutedOrderCollectionQuery(
		"store-1",
		collection.Query{
			Page:     3,
			PageSize: 10,
			Search:   "customer",
			Filters: []collection.Filter{{
				Field:    "settlementStatus",
				Operator: collection.FilterIn,
				Values:   []string{"pending", "disputed"},
			}},
			SortBy:        "updatedAt",
			SortDirection: collection.SortAscending,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 3, query.Page)
	assert.Equal(t, 10, query.PageSize)
	assert.Len(t, predicates, 3)
	assert.Equal(t, "updated_at ASC", orderBy)
}

func TestBuildRoutedOrderCollectionQueryRejectsUnknownSort(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildRoutedOrderCollectionQuery("store-1", collection.Query{
		SortBy: "privateColumn",
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}
