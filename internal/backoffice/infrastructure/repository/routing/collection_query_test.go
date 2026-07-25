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

// RQ1 (docs/07-problems/repository-query-layer-2026-07-24.md): contains on
// a TIMESTAMPTZ column used to build LOWER(created_at) LIKE ... and fail as
// a Postgres 500 at query time. sqlspec.Spec's text-kind guard now rejects
// it up front as a normal validation error, matching store's existing
// behavior for the same operator/kind mismatch.
func TestBuildRoutedOrderCollectionQueryRejectsContainsOnTimestampColumn(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildRoutedOrderCollectionQuery("store-1", collection.Query{
		Filters: []collection.Filter{{
			Field:    "createdAt",
			Operator: collection.FilterContains,
			Values:   []string{"2026"},
		}},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestBuildRoutedOrderCollectionQueryQueueViewOverdue(t *testing.T) {
	t.Parallel()

	_, predicates, _, err := buildRoutedOrderCollectionQuery("store-1", collection.Query{
		Filters: []collection.Filter{{
			Field:    "queueView",
			Operator: collection.FilterEqual,
			Values:   []string{"overdue"},
		}},
	})

	require.NoError(t, err)
	// base store_id predicate + the overdue clause.
	assert.Len(t, predicates, 2)
}

func TestBuildRoutedOrderCollectionQueryQueueViewAllAddsNoPredicate(t *testing.T) {
	t.Parallel()

	_, predicates, _, err := buildRoutedOrderCollectionQuery("store-1", collection.Query{
		Filters: []collection.Filter{{
			Field:    "queueView",
			Operator: collection.FilterEqual,
			Values:   []string{"all"},
		}},
	})

	require.NoError(t, err)
	// base store_id predicate only - "all" adds no predicate of its own.
	assert.Len(t, predicates, 1)
}

func TestBuildRoutedOrderCollectionQueryPrioritySort(t *testing.T) {
	t.Parallel()

	_, _, orderBy, err := buildRoutedOrderCollectionQuery("store-1", collection.Query{
		SortBy:        "priority",
		SortDirection: collection.SortAscending,
	})

	require.NoError(t, err)
	assert.Contains(t, orderBy, "WHEN exception_status = 'escalated' THEN 0")
	assert.Contains(t, orderBy, "ASC")
}
