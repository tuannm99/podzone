package mapper

import (
	pbcommonv1 "github.com/tuannm99/podzone/pkg/api/proto/common/v1"
	"github.com/tuannm99/podzone/pkg/collection"
)

func ToCollectionQuery(request *pbcommonv1.CollectionRequest) collection.Query {
	if request == nil {
		return collection.Query{}.Normalize()
	}
	filters := make([]collection.Filter, 0, len(request.Filters))
	for _, filter := range request.Filters {
		if filter == nil {
			continue
		}
		filters = append(filters, collection.Filter{
			Field:    filter.Field,
			Operator: toFilterOperator(filter.Operator),
			Values:   append([]string(nil), filter.Values...),
		})
	}
	return collection.Query{
		Page:          int(request.Page),
		PageSize:      int(request.PageSize),
		Search:        request.Search,
		Filters:       filters,
		SortBy:        request.SortBy,
		SortDirection: toSortDirection(request.SortDirection),
	}.Normalize()
}

func ToPBPageInfo[T any](page collection.Page[T]) *pbcommonv1.PageInfo {
	return &pbcommonv1.PageInfo{
		Total:       page.Total,
		Page:        int32(page.Page),
		PageSize:    int32(page.PageSize),
		TotalPages:  int32(page.TotalPages),
		HasNext:     page.HasNext,
		HasPrevious: page.HasPrevious,
	}
}

func toSortDirection(direction pbcommonv1.SortDirection) collection.SortDirection {
	if direction == pbcommonv1.SortDirection_SORT_DIRECTION_ASC {
		return collection.SortAscending
	}
	return collection.SortDescending
}

func toFilterOperator(operator pbcommonv1.FilterOperator) collection.FilterOperator {
	switch operator {
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_EQ:
		return collection.FilterEqual
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_NEQ:
		return collection.FilterNotEqual
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_CONTAINS:
		return collection.FilterContains
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_STARTS_WITH:
		return collection.FilterStartsWith
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_GT:
		return collection.FilterGreaterThan
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_GTE:
		return collection.FilterGreaterThanOrEqual
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_LT:
		return collection.FilterLessThan
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_LTE:
		return collection.FilterLessThanOrEqual
	case pbcommonv1.FilterOperator_FILTER_OPERATOR_IN:
		return collection.FilterIn
	default:
		return ""
	}
}
