package mapper

import (
	graphqlmodel "github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	"github.com/tuannm99/podzone/pkg/collection"
)

func ToCollectionQuery(input *graphqlmodel.CollectionInput) collection.Query {
	if input == nil {
		return collection.Query{}.Normalize()
	}
	filters := make([]collection.Filter, 0, len(input.Filters))
	for _, filter := range input.Filters {
		if filter == nil {
			continue
		}
		filters = append(filters, collection.Filter{
			Field:    filter.Field,
			Operator: toFilterOperator(filter.Operator),
			Values:   append([]string(nil), filter.Values...),
		})
	}
	query := collection.Query{
		Filters: filters,
	}
	if input.Page != nil {
		query.Page = *input.Page
	}
	if input.PageSize != nil {
		query.PageSize = *input.PageSize
	}
	if input.Search != nil {
		query.Search = *input.Search
	}
	if input.SortBy != nil {
		query.SortBy = *input.SortBy
	}
	if input.SortDirection != nil {
		if *input.SortDirection == graphqlmodel.CollectionSortDirectionAsc {
			query.SortDirection = collection.SortAscending
		} else {
			query.SortDirection = collection.SortDescending
		}
	}
	return query.Normalize()
}

func toFilterOperator(operator graphqlmodel.CollectionFilterOperator) collection.FilterOperator {
	switch operator {
	case graphqlmodel.CollectionFilterOperatorEq:
		return collection.FilterEqual
	case graphqlmodel.CollectionFilterOperatorNeq:
		return collection.FilterNotEqual
	case graphqlmodel.CollectionFilterOperatorContains:
		return collection.FilterContains
	case graphqlmodel.CollectionFilterOperatorStartsWith:
		return collection.FilterStartsWith
	case graphqlmodel.CollectionFilterOperatorGt:
		return collection.FilterGreaterThan
	case graphqlmodel.CollectionFilterOperatorGte:
		return collection.FilterGreaterThanOrEqual
	case graphqlmodel.CollectionFilterOperatorLt:
		return collection.FilterLessThan
	case graphqlmodel.CollectionFilterOperatorLte:
		return collection.FilterLessThanOrEqual
	case graphqlmodel.CollectionFilterOperatorIn:
		return collection.FilterIn
	default:
		return ""
	}
}

func ToPageInfo[T any](page collection.Page[T]) *graphqlmodel.PageInfo {
	return &graphqlmodel.PageInfo{
		Total:       int(page.Total),
		Page:        page.Page,
		PageSize:    page.PageSize,
		TotalPages:  page.TotalPages,
		HasNext:     page.HasNext,
		HasPrevious: page.HasPrevious,
	}
}
