package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		query    Query
		expected Query
	}{
		{
			name:  "uses defaults",
			query: Query{},
			expected: Query{
				Page:          DefaultPage,
				PageSize:      DefaultPageSize,
				SortDirection: SortDescending,
			},
		},
		{
			name: "caps page size and keeps ascending sort",
			query: Query{
				Page:          2,
				PageSize:      MaxPageSize + 1,
				SortDirection: SortAscending,
			},
			expected: Query{
				Page:          2,
				PageSize:      MaxPageSize,
				SortDirection: SortAscending,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.query.Normalize())
		})
	}
}

func TestNewPage(t *testing.T) {
	t.Parallel()

	page := NewPage([]string{"one", "two"}, 21, Query{Page: 2, PageSize: 10})

	assert.Equal(t, []string{"one", "two"}, page.Items)
	assert.EqualValues(t, 21, page.Total)
	assert.Equal(t, 2, page.Page)
	assert.Equal(t, 10, page.PageSize)
	assert.Equal(t, 3, page.TotalPages)
	assert.True(t, page.HasNext)
	assert.True(t, page.HasPrevious)
}

func TestNewPageReturnsEmptyItems(t *testing.T) {
	t.Parallel()

	page := NewPage[string](nil, 0, Query{})

	assert.Empty(t, page.Items)
	assert.NotNil(t, page.Items)
	assert.Zero(t, page.TotalPages)
	assert.False(t, page.HasNext)
	assert.False(t, page.HasPrevious)
}
