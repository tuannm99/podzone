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

// RQ1 follow-up (docs/07-problems/repository-query-layer-2026-07-24.md):
// slaDays is INTEGER; contains previously built LOWER(sla_days) LIKE ...,
// which Postgres rejects. sqlspec.Spec's text-kind guard now rejects it as
// a normal validation error instead.
func TestBuildPartnerCollectionQueryRejectsContainsOnIntegerField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildPartnerCollectionQuery(partnerdomain.ListPartnersQuery{
		TenantID: "tenant-1",
		Collection: collection.Query{
			Filters: []collection.Filter{{
				Field:    "slaDays",
				Operator: collection.FilterContains,
				Values:   []string{"3"},
			}},
		},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}

// Same defect class on a TIMESTAMPTZ column.
func TestBuildPartnerCollectionQueryRejectsStartsWithOnTimestampField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildPartnerCollectionQuery(partnerdomain.ListPartnersQuery{
		TenantID: "tenant-1",
		Collection: collection.Query{
			Filters: []collection.Filter{{
				Field:    "createdAt",
				Operator: collection.FilterStartsWith,
				Values:   []string{"2026"},
			}},
		},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestBuildPartnerCollectionQueryAppliesStatusAndPartnerTypeFixedFilters(t *testing.T) {
	t.Parallel()

	_, predicates, _, err := buildPartnerCollectionQuery(partnerdomain.ListPartnersQuery{
		TenantID:    "tenant-1",
		Status:      "active",
		PartnerType: "print",
		Collection:  collection.Query{},
	})

	require.NoError(t, err)
	// tenant_id + status + partner_type fixed predicates, no search/filters.
	assert.Len(t, predicates, 3)
}
