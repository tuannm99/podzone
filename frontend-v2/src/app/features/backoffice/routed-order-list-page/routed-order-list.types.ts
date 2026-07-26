import type { CollectionFieldSpec } from '../../../shared/services/collection-field-spec';

export type RoutedOrder = {
  id: string;
  productTitle: string;
  partner: string;
  customerName: string;
  status: string;
  exceptionType: string;
  exceptionStatus: string;
  shipmentStatus: string;
  operatorAssignee: string;
  settlementStatus: string;
  realizedMargin: string;
  createdAt: string;
  updatedAt: string;
  shipmentSlaDueAt?: string;
  issueSlaDueAt?: string;
};

// Matches internal/backoffice/infrastructure/repository/routing/
// collection_query.go's routedOrderCollectionColumns exactly (field keys,
// text kinds). queueView/priority (backend CustomFilters/CustomSorts hooks)
// have no equivalent here yet - see PZEP-0014 Non-Goals.
export const routedOrderFieldSpec: CollectionFieldSpec<RoutedOrder>[] = [
  { field: 'id', label: 'ID', kind: 'text', filterable: true },
  { field: 'productTitle', label: 'Product', kind: 'text', filterable: true, sortable: true },
  { field: 'partner', label: 'Partner', kind: 'text', filterable: true },
  { field: 'customerName', label: 'Customer', kind: 'text', filterable: true },
  { field: 'status', label: 'Status', kind: 'text', filterable: true, sortable: true },
  { field: 'exceptionType', label: 'Exception type', kind: 'text', filterable: true },
  { field: 'exceptionStatus', label: 'Exception status', kind: 'text', filterable: true },
  { field: 'shipmentStatus', label: 'Shipment status', kind: 'text', filterable: true },
  { field: 'operatorAssignee', label: 'Assignee', kind: 'text', filterable: true },
  { field: 'settlementStatus', label: 'Settlement status', kind: 'text', filterable: true },
  { field: 'realizedMargin', label: 'Realized margin', kind: 'text', filterable: true },
  { field: 'createdAt', label: 'Created', kind: 'timestamp', filterable: true, sortable: true },
  { field: 'updatedAt', label: 'Updated', kind: 'timestamp', filterable: true, sortable: true },
  { field: 'shipmentSlaDueAt', label: 'Shipment SLA due', kind: 'timestamp', filterable: true },
  { field: 'issueSlaDueAt', label: 'Issue SLA due', kind: 'timestamp', filterable: true },
];
