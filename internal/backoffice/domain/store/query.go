package store

import "github.com/tuannm99/podzone/pkg/collection"

type ListStoresQuery struct {
	Collection collection.Query
}

type GetStoreQuery struct {
	ID string
}
