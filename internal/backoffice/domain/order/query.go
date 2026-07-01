package order

import "github.com/tuannm99/podzone/pkg/collection"

type ListCustomerOrdersQuery struct {
	StoreID string
}

type ListCustomerOrderPageQuery struct {
	StoreID    string
	Collection collection.Query
}
