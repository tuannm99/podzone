package collection

import "errors"

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

var ErrInvalidQuery = errors.New("invalid collection query")

type SortDirection string

const (
	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

type FilterOperator string

const (
	FilterEqual              FilterOperator = "eq"
	FilterNotEqual           FilterOperator = "neq"
	FilterContains           FilterOperator = "contains"
	FilterStartsWith         FilterOperator = "starts_with"
	FilterGreaterThan        FilterOperator = "gt"
	FilterGreaterThanOrEqual FilterOperator = "gte"
	FilterLessThan           FilterOperator = "lt"
	FilterLessThanOrEqual    FilterOperator = "lte"
	FilterIn                 FilterOperator = "in"
)

type Filter struct {
	Field    string
	Operator FilterOperator
	Values   []string
}

type Query struct {
	Page          int
	PageSize      int
	Search        string
	Filters       []Filter
	SortBy        string
	SortDirection SortDirection
}

func (q Query) Normalize() Query {
	if q.Page < DefaultPage {
		q.Page = DefaultPage
	}
	if q.PageSize <= 0 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}
	if q.SortDirection != SortAscending && q.SortDirection != SortDescending {
		q.SortDirection = SortDescending
	}
	return q
}

func (q Query) Offset() int {
	normalized := q.Normalize()
	return (normalized.Page - 1) * normalized.PageSize
}

type Page[T any] struct {
	Items       []T
	Total       int64
	Page        int
	PageSize    int
	TotalPages  int
	HasNext     bool
	HasPrevious bool
}

func NewPage[T any](items []T, total int64, query Query) Page[T] {
	normalized := query.Normalize()
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(normalized.PageSize) - 1) / int64(normalized.PageSize))
	}
	if items == nil {
		items = []T{}
	}
	return Page[T]{
		Items:       items,
		Total:       total,
		Page:        normalized.Page,
		PageSize:    normalized.PageSize,
		TotalPages:  totalPages,
		HasNext:     normalized.Page < totalPages,
		HasPrevious: normalized.Page > DefaultPage && totalPages > 0,
	}
}
