package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildStoreRequestCollection(t *testing.T) {
	t.Parallel()

	query, filter, sort, err := buildStoreRequestCollection("workspace-1", collection.Query{
		Page:     2,
		PageSize: 10,
		Search:   `urban.*`,
		Filters: []collection.Filter{{
			Field:    "status",
			Operator: collection.FilterIn,
			Values:   []string{"queued", "failed"},
		}},
		SortBy:        "createdAt",
		SortDirection: collection.SortAscending,
	})

	require.NoError(t, err)
	assert.Equal(t, 2, query.Page)
	assert.Contains(t, filter, "$and")
	require.Len(t, sort, 2)
	assert.Equal(t, "created_at", sort[0].Key)
	assert.Equal(t, 1, sort[0].Value)
	assert.NotContains(t, filter, `urban.*`)
}

func TestBuildStoreRequestCollectionRejectsUnsupportedField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildStoreRequestCollection("workspace-1", collection.Query{
		Filters: []collection.Filter{{
			Field:    "raw",
			Operator: collection.FilterEqual,
			Values:   []string{"value"},
		}},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}
