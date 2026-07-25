import type { CollectionFieldSpec } from '../../../shared/services/collection-field-spec';

export type StoreInfo = {
  id: string;
  name: string;
  ownerId: string;
  isActive: boolean;
  description: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

// Matches internal/backoffice/infrastructure/repository/store/collection_query.go's
// storeSpec exactly (field keys, filterable/text kinds) - see
// docs/08-adr/ADR-0008-frontend-v2-declarative-collection-field-spec.md.
export const storeFieldSpec: CollectionFieldSpec<StoreInfo>[] = [
  { field: 'id', label: 'ID', kind: 'text', filterable: true },
  { field: 'name', label: 'Name', kind: 'text', filterable: true, sortable: true },
  { field: 'description', label: 'Description', kind: 'text', filterable: true },
  { field: 'ownerId', label: 'Owner ID', kind: 'text', filterable: true },
  { field: 'status', label: 'Status', kind: 'text', filterable: true, sortable: true },
  { field: 'createdAt', label: 'Created', kind: 'timestamp', filterable: true, sortable: true },
  { field: 'updatedAt', label: 'Updated', kind: 'timestamp', filterable: true, sortable: true },
];
