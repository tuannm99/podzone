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

// Simplified from Solid's provisioningStatusLabel/readinessBadgeColor
// (presentation.ts) — driven by the request's own `status`, not a separate
// per-request readiness/ui_state lookup. Good enough to unblock "list
// stores, open the ready ones" now; swap for the full ui_state polling flow
// (PZEP-0008 M2) if a store needs a more granular in-progress breakdown.
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
