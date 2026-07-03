export type CollectionSortDirection =
  | 'SORT_DIRECTION_ASC'
  | 'SORT_DIRECTION_DESC'

export type CollectionFilterOperator =
  | 'FILTER_OPERATOR_EQ'
  | 'FILTER_OPERATOR_NEQ'
  | 'FILTER_OPERATOR_CONTAINS'
  | 'FILTER_OPERATOR_STARTS_WITH'
  | 'FILTER_OPERATOR_GT'
  | 'FILTER_OPERATOR_GTE'
  | 'FILTER_OPERATOR_LT'
  | 'FILTER_OPERATOR_LTE'
  | 'FILTER_OPERATOR_IN'

export type CollectionFilter = {
  field: string
  operator: CollectionFilterOperator
  values: string[]
}

export type CollectionQuery = {
  page: number
  pageSize: number
  search?: string
  filters?: CollectionFilter[]
  sortBy?: string
  sortDirection?: CollectionSortDirection
}

export type PageInfo = {
  total: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrevious: boolean
}

export type WirePageInfo = Omit<PageInfo, 'total'> & {
  total: number | string
}

export type CollectionPage<T> = {
  items: T[]
  pageInfo: PageInfo
}

export const emptyPageInfo = (query: CollectionQuery): PageInfo => ({
  total: 0,
  page: query.page,
  pageSize: query.pageSize,
  totalPages: 0,
  hasNext: false,
  hasPrevious: false,
})

export function normalizePageInfo(
  value: WirePageInfo | undefined,
  query: CollectionQuery
): PageInfo {
  if (!value) return emptyPageInfo(query)
  const total =
    typeof value.total === 'string'
      ? Number.parseInt(value.total, 10)
      : value.total
  return {
    ...value,
    total: Number.isFinite(total) ? total : 0,
  }
}

export function toCollectionParams(query: CollectionQuery) {
  const params: Record<string, string | number | string[]> = {
    'collection.page': query.page,
    'collection.pageSize': query.pageSize,
  }
  if (query.search?.trim()) {
    params['collection.search'] = query.search.trim()
  }
  if (query.sortBy) {
    params['collection.sortBy'] = query.sortBy
  }
  if (query.sortDirection) {
    params['collection.sortDirection'] = query.sortDirection
  }
  query.filters?.forEach((filter, index) => {
    params[`collection.filters[${index}].field`] = filter.field
    params[`collection.filters[${index}].operator`] = filter.operator
    params[`collection.filters[${index}].values`] = filter.values
  })
  return params
}

export function toGraphQLCollectionInput(query: CollectionQuery) {
  return {
    ...query,
    filters: query.filters?.map((filter) => ({
      ...filter,
      operator: filter.operator.replace('FILTER_OPERATOR_', ''),
    })),
    sortDirection:
      query.sortDirection === 'SORT_DIRECTION_ASC' ? 'ASC' : 'DESC',
  }
}
