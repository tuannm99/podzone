package collection

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const maxURLFilters = 100

func ParseURLValues(values url.Values, prefix string) (Query, error) {
	query := Query{}
	var err error
	if query.Page, err = parseOptionalInt(values.Get(prefix + "page")); err != nil {
		return Query{}, fmt.Errorf("%w: invalid page", ErrInvalidQuery)
	}
	if query.PageSize, err = parseOptionalInt(values.Get(prefix + "pageSize")); err != nil {
		return Query{}, fmt.Errorf("%w: invalid page size", ErrInvalidQuery)
	}
	query.Search = values.Get(prefix + "search")
	query.SortBy = values.Get(prefix + "sortBy")
	query.SortDirection, err = parseSortDirection(values.Get(prefix + "sortDirection"))
	if err != nil {
		return Query{}, err
	}
	query.Filters, err = parseURLFilters(values, prefix)
	if err != nil {
		return Query{}, err
	}
	return query.Normalize(), nil
}

func parseOptionalInt(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func parseSortDirection(value string) (SortDirection, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "sort_direction_desc", "desc":
		return SortDescending, nil
	case "sort_direction_asc", "asc":
		return SortAscending, nil
	default:
		return "", fmt.Errorf("%w: invalid sort direction", ErrInvalidQuery)
	}
}

func parseURLFilters(values url.Values, prefix string) ([]Filter, error) {
	filters := make([]Filter, 0)
	for index := range maxURLFilters {
		filterPrefix := fmt.Sprintf("%sfilters[%d].", prefix, index)
		field := strings.TrimSpace(values.Get(filterPrefix + "field"))
		if field == "" {
			continue
		}
		operator, err := parseFilterOperator(values.Get(filterPrefix + "operator"))
		if err != nil {
			return nil, err
		}
		filterValues := append([]string(nil), values[filterPrefix+"values"]...)
		filterValues = append(filterValues, values[filterPrefix+"values[]"]...)
		if len(filterValues) == 0 {
			return nil, fmt.Errorf("%w: filter %q requires a value", ErrInvalidQuery, field)
		}
		filters = append(filters, Filter{
			Field:    field,
			Operator: operator,
			Values:   filterValues,
		})
	}
	return filters, nil
}

func parseFilterOperator(value string) (FilterOperator, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "filter_operator_eq", "eq":
		return FilterEqual, nil
	case "filter_operator_neq", "neq":
		return FilterNotEqual, nil
	case "filter_operator_contains", "contains":
		return FilterContains, nil
	case "filter_operator_starts_with", "starts_with":
		return FilterStartsWith, nil
	case "filter_operator_gt", "gt":
		return FilterGreaterThan, nil
	case "filter_operator_gte", "gte":
		return FilterGreaterThanOrEqual, nil
	case "filter_operator_lt", "lt":
		return FilterLessThan, nil
	case "filter_operator_lte", "lte":
		return FilterLessThanOrEqual, nil
	case "filter_operator_in", "in":
		return FilterIn, nil
	default:
		return "", fmt.Errorf("%w: invalid filter operator", ErrInvalidQuery)
	}
}
