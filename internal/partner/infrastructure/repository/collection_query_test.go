package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildPartnerCollectionQuery(t *testing.T) {
	t.Parallel()

	query, predicates, orderBy, err := buildPartnerCollectionQuery(
		partnerdomain.ListPartnersQuery{
			TenantID: "tenant-1",
			Collection: collection.Query{
				Page:     2,
				PageSize: 8,
				Search:   "print",
				Filters: []collection.Filter{{
					Field:    "status",
					Operator: collection.FilterEqual,
					Values:   []string{"active"},
				}},
				SortBy:        "name",
				SortDirection: collection.SortAscending,
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 2, query.Page)
	assert.Equal(t, 8, query.PageSize)
	assert.Len(t, predicates, 3)
	assert.Equal(t, "name ASC", orderBy)
}

func TestBuildPartnerCollectionQueryRejectsUnknownField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildPartnerCollectionQuery(partnerdomain.ListPartnersQuery{
		TenantID: "tenant-1",
		Collection: collection.Query{
			Filters: []collection.Filter{{
				Field:    "secret",
				Operator: collection.FilterEqual,
				Values:   []string{"value"},
			}},
		},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}
