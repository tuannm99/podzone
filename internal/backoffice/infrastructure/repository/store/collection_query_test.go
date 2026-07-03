package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildStoreCollectionQuery(t *testing.T) {
	t.Parallel()

	query, predicates, orderBy, err := buildStoreCollectionQuery(collection.Query{
		Page:     2,
		PageSize: 10,
		Search:   `urban%`,
		Filters: []collection.Filter{{
			Field:    "status",
			Operator: collection.FilterEqual,
			Values:   []string{"active"},
		}},
		SortBy:        "name",
		SortDirection: collection.SortAscending,
	})

	require.NoError(t, err)
	assert.Equal(t, 2, query.Page)
	assert.Len(t, predicates, 2)
	assert.Equal(t, "name ASC", orderBy)
}

func TestBuildStoreCollectionQueryRejectsUnsupportedField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildStoreCollectionQuery(collection.Query{
		Filters: []collection.Filter{{
			Field:    "raw",
			Operator: collection.FilterEqual,
			Values:   []string{"value"},
		}},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}
