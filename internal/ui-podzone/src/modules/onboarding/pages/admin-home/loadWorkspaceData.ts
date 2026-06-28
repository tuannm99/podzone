import { ensureActiveTenant } from '@/services/auth'
import type { TenantMembership } from '@/services/iam'
import { getRoutedOrders } from '@/services/orders'
import { listStoreRequests } from '@/services/onboarding'
import { listStores } from '@/services/store'
import { storeStorage } from '@/services/storeStorage'
import { tenantStorage } from '@/services/tenantStorage'
import { tokenStorage } from '@/services/tokenStorage'
import {
  isOverdue,
  type StoreAttention,
  type WorkspaceSummary,
} from './presentation'

export async function collectWorkspaceData(
  memberships: TenantMembership[]
): Promise<{
  summaries: WorkspaceSummary[]
  attention: StoreAttention[]
}> {
  const activeWorkspaces = memberships.filter(
    (membership) => membership.status === 'active'
  )
  if (activeWorkspaces.length === 0) {
    return { summaries: [], attention: [] }
  }

  const originalTenantID = tokenStorage.getActiveTenantID()
  const originalStoreID = originalTenantID
    ? storeStorage.getStoreID(originalTenantID)
    : ''
  const previousStoreByTenant = new Map<string, string>()
  const summaries: WorkspaceSummary[] = []
  const attention: StoreAttention[] = []

  try {
    for (const membership of activeWorkspaces) {
      const switched = await ensureActiveTenant(membership.tenantId)
      if (!switched.success) continue

      const requestsResult = await listStoreRequests(membership.tenantId)
      const storesResult = await listStores()
      const storeRequests = requestsResult.success ? requestsResult.data : []
      const stores = storesResult.success ? storesResult.data : []
      summaries.push({
        tenantId: membership.tenantId,
        roleName: membership.roleName,
        status: membership.status,
        userId: membership.userId,
        stores,
        storeRequests,
        storeCount: stores.length,
        activeStoreCount: stores.filter((store) => store.isActive).length,
      })

      for (const store of stores) {
        if (!previousStoreByTenant.has(membership.tenantId)) {
          previousStoreByTenant.set(
            membership.tenantId,
            storeStorage.getStoreID(membership.tenantId)
          )
        }
        storeStorage.setStoreID(membership.tenantId, store.id)
        const ordersResult = await getRoutedOrders()
        if (!ordersResult.success) continue

        const orders = ordersResult.data.orders
        attention.push({
          tenantId: membership.tenantId,
          storeId: store.id,
          storeName: store.name,
          overdueCount: orders.filter(
            (order) =>
              (!!order.shipmentSlaDueAt &&
                isOverdue(order.shipmentSlaDueAt) &&
                order.shipmentStatus !== 'delivered') ||
              (!!order.issueSlaDueAt &&
                isOverdue(order.issueSlaDueAt) &&
                (order.exceptionStatus === 'open' ||
                  order.exceptionStatus === 'escalated' ||
                  order.shipmentStatus === 'delivery_issue'))
          ).length,
          disputedCount: orders.filter(
            (order) => order.settlementStatus === 'disputed'
          ).length,
          unassignedCount: orders.filter(
            (order) =>
              !order.operatorAssignee || order.operatorAssignee === 'unassigned'
          ).length,
        })
      }
    }
  } finally {
    previousStoreByTenant.forEach((storeID, tenantID) => {
      if (storeID) storeStorage.setStoreID(tenantID, storeID)
      else storeStorage.clearStoreID(tenantID)
    })
    if (originalTenantID) {
      await ensureActiveTenant(originalTenantID)
      tenantStorage.setTenantID(originalTenantID)
      if (originalStoreID) {
        storeStorage.setStoreID(originalTenantID, originalStoreID)
      } else {
        storeStorage.clearStoreID(originalTenantID)
      }
    }
  }

  return { summaries, attention }
}
