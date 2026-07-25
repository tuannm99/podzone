import type { CollectionFilterOperator } from './collection-types';

// See docs/08-adr/ADR-0008-frontend-v2-declarative-collection-field-spec.md.
// A collection page declares its queryable fields once, here, instead of
// inlining a filter-fields array per page - operator validity follows from
// `kind` automatically (operatorsFor below), closing off the RQ1 defect
// class (offering `contains` on a non-text field) at the UI layer before
// the request ever reaches the backend.
export type CollectionFieldKind = 'text' | 'number' | 'boolean' | 'timestamp' | 'enum';

export type CollectionFieldSpec<T> = {
  field: keyof T & string;
  label: string;
  kind: CollectionFieldKind;
  filterable?: boolean;
  sortable?: boolean;
};

// Deliberately conservative per kind - only operators that produce a
// well-formed query for that SQL/Mongo column type. Matches the backend's
// own kind-gated operator set (pkg/collection/sqlspec.Spec.filterClause).
export function operatorsFor(kind: CollectionFieldKind): CollectionFilterOperator[] {
  switch (kind) {
    case 'text':
      return [
        'FILTER_OPERATOR_EQ',
        'FILTER_OPERATOR_NEQ',
        'FILTER_OPERATOR_CONTAINS',
        'FILTER_OPERATOR_STARTS_WITH',
        'FILTER_OPERATOR_IN',
      ];
    case 'number':
      return [
        'FILTER_OPERATOR_EQ',
        'FILTER_OPERATOR_NEQ',
        'FILTER_OPERATOR_GT',
        'FILTER_OPERATOR_GTE',
        'FILTER_OPERATOR_LT',
        'FILTER_OPERATOR_LTE',
        'FILTER_OPERATOR_IN',
      ];
    case 'timestamp':
      return [
        'FILTER_OPERATOR_EQ',
        'FILTER_OPERATOR_NEQ',
        'FILTER_OPERATOR_GT',
        'FILTER_OPERATOR_GTE',
        'FILTER_OPERATOR_LT',
        'FILTER_OPERATOR_LTE',
      ];
    case 'boolean':
      return ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_NEQ'];
    case 'enum':
      return ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_NEQ', 'FILTER_OPERATOR_IN'];
  }
}

export const operatorLabels: Record<CollectionFilterOperator, string> = {
  FILTER_OPERATOR_EQ: 'Equals',
  FILTER_OPERATOR_NEQ: 'Not equal',
  FILTER_OPERATOR_CONTAINS: 'Contains',
  FILTER_OPERATOR_STARTS_WITH: 'Starts with',
  FILTER_OPERATOR_GT: 'Greater than',
  FILTER_OPERATOR_GTE: 'Greater or equal',
  FILTER_OPERATOR_LT: 'Less than',
  FILTER_OPERATOR_LTE: 'Less or equal',
  FILTER_OPERATOR_IN: 'In list',
};
