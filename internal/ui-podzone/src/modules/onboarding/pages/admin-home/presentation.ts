import type { StoreRequestStatus } from '@/services/onboarding'

export type StoreAttention = {
  tenantId: string
  storeId: string
  storeName: string
  overdueCount: number
  disputedCount: number
  unassignedCount: number
}

export type WorkspaceSummary = {
  tenantId: string
  roleName: string
  status: string
  userId: number
  storeCount: number
  activeStoreCount: number
}

export function parseUserID(raw: unknown): number {
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10)
    return Number.isFinite(parsed) ? parsed : 0
  }
  return 0
}

export function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export function membershipStatusColor(status: string) {
  return status === 'active' ? 'green' : 'dark'
}

export function provisioningStatusLabel(status: StoreRequestStatus) {
  switch (status) {
    case 'planning':
      return 'Planning placement'
    case 'planned':
      return 'Placement planned'
    case 'pending_approval':
      return 'Pending approval'
    case 'queued':
      return 'Queued'
    case 'provisioning':
      return 'Provisioning infrastructure'
    case 'ready':
      return 'Ready'
    case 'failed':
      return 'Provisioning failed'
    case 'failed_retryable':
      return 'Failed, retry available'
    case 'failed_non_retryable':
      return 'Failed, platform action required'
    case 'pending_platform_setup':
      return 'Pending platform setup'
    default:
      return status.charAt(0).toUpperCase() + status.slice(1)
  }
}

export function isOverdue(value?: string) {
  if (!value) {
    return false
  }
  return new Date(value).getTime() < Date.now()
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
  })
  return `/t/${tenantID}/orders?${params.toString()}`
}
