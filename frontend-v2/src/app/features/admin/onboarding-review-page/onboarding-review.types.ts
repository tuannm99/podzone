import type { CollectionFieldSpec } from '../../../shared/services/collection-field-spec';
import type { StoreRequest } from '../../onboarding/store/store.types';

// Presentational shape for the admin cross-tenant review screen —
// camelCase field keys matching internal/onboarding/infrastructure/
// repository/store/collection_query.go's requestCollectionFields exactly
// (see docs/08-adr/ADR-0008), mapped from the wire DTO's snake_case JSON
// (StoreRequest, ported from frontend/'s onboarding contract).
export type AdminStoreRequest = {
  id: string;
  workspaceId: string;
  name: string;
  subdomain: string;
  requestedBy: string;
  ownerId: string;
  status: StoreRequest['status'];
  storeId?: string;
  lastError?: string;
  createdAt: string;
  updatedAt: string;
};

export function toAdminStoreRequest(wire: StoreRequest): AdminStoreRequest {
  return {
    id: wire.id,
    workspaceId: wire.workspace_id,
    name: wire.name,
    subdomain: wire.subdomain,
    requestedBy: wire.requested_by,
    ownerId: wire.owner_id,
    status: wire.status,
    storeId: wire.store_id,
    lastError: wire.last_error,
    createdAt: wire.created_at,
    updatedAt: wire.updated_at,
  };
}

// Mirrors internal/onboarding/domain/store.isValidStatusTransition exactly
// (Approve: requested|pending_approval -> queued; Reject: pending_approval
// -> rejected only) - client-side gating is UX only, the backend is the
// actual authority and re-validates on every call.
export function canApprove(status: AdminStoreRequest['status']): boolean {
  return status === 'requested' || status === 'pending_approval';
}

export function canReject(status: AdminStoreRequest['status']): boolean {
  return status === 'pending_approval';
}

export const adminStoreRequestFieldSpec: CollectionFieldSpec<AdminStoreRequest>[] = [
  { field: 'id', label: 'ID', kind: 'text', filterable: true },
  { field: 'workspaceId', label: 'Tenant', kind: 'text', filterable: true, sortable: true },
  { field: 'name', label: 'Name', kind: 'text', filterable: true, sortable: true },
  { field: 'subdomain', label: 'Subdomain', kind: 'text', filterable: true, sortable: true },
  { field: 'requestedBy', label: 'Requested By', kind: 'text', filterable: true, sortable: true },
  { field: 'status', label: 'Status', kind: 'enum', filterable: true, sortable: true },
  { field: 'storeId', label: 'Store ID', kind: 'text', filterable: true },
  { field: 'createdAt', label: 'Created', kind: 'timestamp', filterable: true, sortable: true },
  { field: 'updatedAt', label: 'Updated', kind: 'timestamp', filterable: true, sortable: true },
];
