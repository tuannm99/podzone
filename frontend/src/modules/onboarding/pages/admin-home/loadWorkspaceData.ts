import { ensureActiveTenant } from '@/services/auth'
import type { TenantMembership } from '@/services/iam'
import { getRoutedOrders } from '@/services/orders'
import { listAllStores } from '@/services/store'
import type { AuthContext } from '@/modules/shell/auth-context'
import { isOverdue, type StoreAttention, type WorkspaceSummary } from './presentation'

export async function collectWorkspaceData(
    memberships: TenantMembership[],
    auth: AuthContext
): Promise<{
    summaries: WorkspaceSummary[]
    attention: StoreAttention[]
}> {
    const activeWorkspaces = memberships.filter((membership) => membership.status === 'active')
    if (activeWorkspaces.length === 0) {
        return { summaries: [], attention: [] }
    }

    const originalTenantID = auth.getActiveTenantId()
    const originalStoreID = originalTenantID ? auth.getStoreId(originalTenantID) : ''
    const previousStoreByTenant = new Map<string, string>()
    const summaries: WorkspaceSummary[] = []
    const attention: StoreAttention[] = []

    try {
        for (const membership of activeWorkspaces) {
            const switched = await ensureActiveTenant(membership.tenantId)
            if (!switched.success) continue

            const storesResult = await listAllStores()
            const stores = storesResult.success ? storesResult.data : []
            summaries.push({
                tenantId: membership.tenantId,
                roleName: membership.roleName,
                status: membership.status,
                userId: membership.userId,
                storeCount: stores.length,
                activeStoreCount: stores.filter((store) => store.isActive).length,
            })

            for (const store of stores) {
                if (!previousStoreByTenant.has(membership.tenantId)) {
                    previousStoreByTenant.set(membership.tenantId, auth.getStoreId(membership.tenantId))
                }
                auth.setStoreId(membership.tenantId, store.id)
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
                    disputedCount: orders.filter((order) => order.settlementStatus === 'disputed').length,
                    unassignedCount: orders.filter(
                        (order) => !order.operatorAssignee || order.operatorAssignee === 'unassigned'
                    ).length,
                })
            }
        }
    } finally {
        previousStoreByTenant.forEach((storeID, tenantID) => {
            if (storeID) auth.setStoreId(tenantID, storeID)
            else auth.clearStoreId(tenantID)
        })
        if (originalTenantID) {
            await ensureActiveTenant(originalTenantID)
            auth.setActiveTenantId(originalTenantID)
            if (originalStoreID) {
                auth.setStoreId(originalTenantID, originalStoreID)
            } else {
                auth.clearStoreId(originalTenantID)
            }
        }
    }

    return { summaries, attention }
}
