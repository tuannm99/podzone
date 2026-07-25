// Ported from frontend/packages/shared/services/collection.ts — keep in
// sync if that file changes. `toCollectionParams` (REST query-string form)
// is still not ported; add it when a REST (non-GraphQL) collection screen
// needs it.

export type CollectionSortDirection = 'SORT_DIRECTION_ASC' | 'SORT_DIRECTION_DESC';

export type CollectionFilterOperator =
  | 'FILTER_OPERATOR_EQ'
  | 'FILTER_OPERATOR_NEQ'
  | 'FILTER_OPERATOR_CONTAINS'
  | 'FILTER_OPERATOR_STARTS_WITH'
  | 'FILTER_OPERATOR_GT'
  | 'FILTER_OPERATOR_GTE'
  | 'FILTER_OPERATOR_LT'
  | 'FILTER_OPERATOR_LTE'
  | 'FILTER_OPERATOR_IN';

export type CollectionFilter = {
  field: string;
  operator: CollectionFilterOperator;
  values: string[];
};

export type CollectionQuery = {
  page: number;
  pageSize: number;
  search?: string;
  filters?: CollectionFilter[];
  sortBy?: string;
  sortDirection?: CollectionSortDirection;
};

export type PageInfo = {
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
};

export type CollectionPage<T> = {
  items: T[];
  pageInfo: PageInfo;
};

export type WirePageInfo = Omit<PageInfo, 'total'> & { total: number | string };

export const emptyPageInfo = (query: CollectionQuery): PageInfo => ({
  total: 0,
  page: query.page,
  pageSize: query.pageSize,
  totalPages: 0,
  hasNext: false,
  hasPrevious: false,
});

// GraphQL responses carry `total` as a string when the resolver returns an
// int64 count - matches frontend/'s normalizePageInfo exactly.
export function normalizePageInfo(
  value: WirePageInfo | undefined,
  query: CollectionQuery,
): PageInfo {
  if (!value) return emptyPageInfo(query);
  const total = typeof value.total === 'string' ? Number.parseInt(value.total, 10) : value.total;
  return { ...value, total: Number.isFinite(total) ? total : 0 };
}

export function toGraphQLCollectionInput(query: CollectionQuery) {
  return {
    ...query,
    filters: query.filters?.map((filter) => ({
      ...filter,
      operator: filter.operator.replace('FILTER_OPERATOR_', ''),
    })),
    sortDirection: query.sortDirection === 'SORT_DIRECTION_ASC' ? 'ASC' : 'DESC',
  };
}
