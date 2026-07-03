package collection

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURLValues(t *testing.T) {
	t.Parallel()

	values := url.Values{
		"collection.page":                      {"2"},
		"collection.pageSize":                  {"10"},
		"collection.search":                    {"urban"},
		"collection.sortBy":                    {"createdAt"},
		"collection.sortDirection":             {"SORT_DIRECTION_ASC"},
		"collection.filters[0].field":          {"status"},
		"collection.filters[0].operator":       {"FILTER_OPERATOR_IN"},
		"collection.filters[0].values[]":       {"queued", "failed"},
	}

	query, err := ParseURLValues(values, "collection.")

	require.NoError(t, err)
	assert.Equal(t, 2, query.Page)
	assert.Equal(t, 10, query.PageSize)
	assert.Equal(t, SortAscending, query.SortDirection)
	require.Len(t, query.Filters, 1)
	assert.Equal(t, []string{"queued", "failed"}, query.Filters[0].Values)
}

func TestParseURLValuesRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	_, err := ParseURLValues(url.Values{
		"collection.page": {"invalid"},
	}, "collection.")

	require.ErrorIs(t, err, ErrInvalidQuery)
}
