// Ported verbatim (types only, no service logic) from
// frontend/packages/shared/services/collection.ts — keep in sync if that
// file changes. Full service logic (toCollectionParams, normalizePageInfo,
// etc.) is not ported yet; add it when a real feature needs it.

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
