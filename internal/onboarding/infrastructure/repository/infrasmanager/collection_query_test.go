package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildConnectionCollectionScopesSearchFilterAndSort(t *testing.T) {
	query, filter, sort, err := buildConnectionCollection(
		"tenant-1",
		false,
		collection.Query{
			Page:          2,
			PageSize:      25,
			Search:        "primary.*",
			SortBy:        "name",
			SortDirection: collection.SortAscending,
			Filters: []collection.Filter{{
				Field:    "infraType",
				Operator: collection.FilterEqual,
				Values:   []string{"postgres"},
			}},
		},
	)

	require.NoError(t, err)
	require.Equal(t, 25, query.Offset())
	require.Equal(t, bson.D{{Key: "name", Value: 1}}, sort)
	clauses := filter["$and"].(bson.A)
	require.Contains(t, clauses, bson.M{"tenant_id": "tenant-1"})
	require.Contains(t, clauses, bson.M{"deleted_at": nil})
	require.Contains(t, clauses, bson.M{"infra_type": "postgres"})
	search := clauses[2].(bson.M)["$or"].(bson.A)
	require.Equal(t, primitive.Regex{Pattern: "primary\\.\\*", Options: "i"}, search[0].(bson.M)["infra_type"])
}

func TestBuildEventCollectionRejectsUnknownField(t *testing.T) {
	_, _, _, err := buildEventCollection("tenant-1", collection.Query{
		Filters: []collection.Filter{{
			Field:    "payload",
			Operator: collection.FilterEqual,
			Values:   []string{"secret"},
		}},
	})

	require.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestBuildConnectionCollectionParsesNumericFilter(t *testing.T) {
	_, filter, _, err := buildConnectionCollection("tenant-1", true, collection.Query{
		Filters: []collection.Filter{{
			Field:    "version",
			Operator: collection.FilterGreaterThan,
			Values:   []string{"3"},
		}},
	})

	require.NoError(t, err)
	clauses := filter["$and"].(bson.A)
	require.Contains(t, clauses, bson.M{"version": bson.M{"$gt": int64(3)}})
	require.NotContains(t, clauses, bson.M{"deleted_at": nil})
}

func TestBuildRuntimePoolCollectionParsesHealthFilter(t *testing.T) {
	_, filter, _, err := buildRuntimePoolCollection(collection.Query{
		Filters: []collection.Filter{{
			Field:    "healthy",
			Operator: collection.FilterEqual,
			Values:   []string{"true"},
		}},
	})

	require.NoError(t, err)
	clauses := filter["$and"].(bson.A)
	require.Contains(t, clauses, bson.M{"healthy": true})
}
