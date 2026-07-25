package store

import (
	sq "github.com/Masterminds/squirrel"

	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/collection/sqlspec"
)

// See docs/07-problems/repository-query-layer-2026-07-24.md RQ2. Filter/sort
// behavior is unchanged from the previous hand-written version - proven by
// the existing collection_query_test.go passing unmodified against this
// implementation.
var storeSpec = sqlspec.Spec{
	Entity:      "store",
	DefaultSort: "createdAt",
	SearchOrder: []string{"id", "name", "description", "status"},
	Fields: map[string]sqlspec.Field{
		"id":          {Column: "id", Kind: sqlspec.KindText},
		"name":        {Column: "name", Kind: sqlspec.KindText},
		"description": {Column: "description", Kind: sqlspec.KindText},
		"ownerId":     {Column: "owner_id"},
		"status":      {Column: "status", Kind: sqlspec.KindText},
		"createdAt":   {Column: "created_at"},
		"updatedAt":   {Column: "updated_at"},
	},
}

func buildStoreCollectionQuery(
	query collection.Query,
) (collection.Query, []sq.Sqlizer, string, error) {
	return storeSpec.Compile(query)
}
