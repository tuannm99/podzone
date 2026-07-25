// Ported from frontend/packages/shared/services/onboarding.ts — keep both in
// sync if the backend contract changes (docs/03-architecture-detail-design/
// services/onboarding/api-design.md).
export type StoreRequestStatus =
  | 'requested'
  | 'planning'
  | 'planned'
  | 'pending_approval'
  | 'queued'
  | 'provisioning'
  | 'ready'
  | 'failed'
  | 'failed_retryable'
  | 'failed_non_retryable'
  | 'pending_platform_setup'
  | 'rejected'
  | 'suspended'
  | 'archived'
  | 'cancelled';

export type StoreRequest = {
  id: string;
  workspace_id: string;
  name: string;
  subdomain: string;
  requested_by: string;
  owner_id: string;
  status: StoreRequestStatus;
  store_id?: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
};

// `storeStatusTone`/`storeStatusLabel` render the request's raw `status`
// (used as a fallback before readiness has loaded, and for terminal
// non-provisioning statuses the readiness endpoint doesn't model). Whether a
// store is actually safe to open is decided by `isReadyToOpen` below, driven
// by the readiness endpoint's `ui_state`, not this status alone.
export type StoreStatusTone = 'success' | 'danger' | 'warning' | 'neutral';

const failedStatuses: StoreRequestStatus[] = ['failed', 'failed_retryable', 'failed_non_retryable'];
const inactiveStatuses: StoreRequestStatus[] = ['suspended', 'archived', 'cancelled', 'rejected'];

export function storeStatusTone(status: StoreRequestStatus): StoreStatusTone {
  if (status === 'ready') return 'success';
  if (failedStatuses.includes(status)) return 'danger';
  if (inactiveStatuses.includes(status)) return 'neutral';
  return 'warning';
}

export function storeStatusLabel(status: StoreRequestStatus): string {
  return status
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

export function isRetryableStatus(status: StoreRequestStatus): boolean {
  return failedStatuses.includes(status);
}

// Statuses where the backend may still transition the request without user
// action - matches Solid's createAdminHomeViewModel.ts poll-trigger list.
export const activeProvisioningStatuses: StoreRequestStatus[] = [
  'requested',
  'planning',
  'planned',
  'pending_approval',
  'queued',
  'provisioning',
  'pending_platform_setup',
];

// Ported from frontend/packages/shared/services/onboarding.ts - keep both in
// sync if the backend contract changes (docs/03-architecture-detail-design/
// services/onboarding/api-design.md).
export type StoreReadinessUiState = 'pending' | 'provisioning' | 'blocked' | 'failed' | 'ready';

export type StoreReadiness = {
  request_id: string;
  request_status: StoreRequestStatus;
  readiness: {
    store_ready: boolean;
    placement_allocation_ready: boolean;
    route_ready: boolean;
  };
  failure_reason?: string;
  ui_state: StoreReadinessUiState;
};

// A request's raw `status` can say "ready" while readiness (placement/route)
// isn't - the readiness endpoint's `ui_state` is authoritative for whether a
// store can actually be opened. See docs/03-architecture-detail-design/
// 05-transport-contracts.md "Slice 0.3: Store Readiness Contract".
export function readinessTone(
  status: StoreRequestStatus,
  uiState: StoreReadinessUiState | undefined,
): StoreStatusTone {
  if (uiState === 'ready') return 'success';
  if (uiState === 'failed') return 'danger';
  if (uiState === 'pending' || uiState === 'provisioning' || uiState === 'blocked')
    return 'warning';
  return storeStatusTone(status);
}

export function isReadyToOpen(
  store: { store_id?: string },
  uiState: StoreReadinessUiState | undefined,
): boolean {
  return uiState === 'ready' && !!store.store_id;
}
