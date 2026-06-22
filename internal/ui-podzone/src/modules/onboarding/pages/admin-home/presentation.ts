import type { StoreRequest, StoreRequestStatus } from '@/services/onboarding';
import type { StoreInfo } from '@/services/store';

export type StoreAttention = {
  tenantId: string;
  storeId: string;
  storeName: string;
  overdueCount: number;
  disputedCount: number;
  unassignedCount: number;
};

export type WorkspaceSummary = {
  tenantId: string;
  roleName: string;
  status: string;
  userId: number;
  stores: StoreInfo[];
  storeRequests: StoreRequest[];
  storeCount: number;
  activeStoreCount: number;
};

export const provisioningSteps: StoreRequestStatus[] = [
  'requested',
  'queued',
  'provisioning',
  'ready',
];

export function parseUserID(raw: unknown): number {
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw;
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

export function membershipStatusColor(status: string) {
  return status === 'active' ? 'green' : 'dark';
}

export function provisioningStepIndex(status: StoreRequestStatus) {
  if (status === 'pending_approval') return 0;
  if (status === 'failed') return 2;
  return provisioningSteps.indexOf(status);
}

export function provisioningStatusLabel(status: StoreRequestStatus) {
  switch (status) {
    case 'pending_approval':
      return 'Pending approval';
    case 'queued':
      return 'Queued';
    case 'provisioning':
      return 'Provisioning infrastructure';
    case 'ready':
      return 'Ready';
    case 'failed':
      return 'Provisioning failed';
    default:
      return status.charAt(0).toUpperCase() + status.slice(1);
  }
}

export function isOverdue(value?: string) {
  if (!value) {
    return false;
  }
  return new Date(value).getTime() < Date.now();
}

export function buildOrdersHref(
  tenantID: string,
  storeID: string,
  queueView: string
) {
  const params = new URLSearchParams({
    storeId: storeID,
    queueView,
    queueSort: 'priority',
  });
  return `/t/${tenantID}/orders?${params.toString()}`;
}
