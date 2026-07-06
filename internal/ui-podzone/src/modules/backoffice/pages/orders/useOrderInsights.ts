import type { Accessor } from 'solid-js'
import type { RoutedOrder } from '@/services/orders'
import type { QueueSort, QueueView } from './board-context'
import type { ActivityFilter } from './order-card/types'
import {
    formatActivitySummary,
    formatStoreActivitySummary,
    hasFinanceAttention,
    parseMoneyValue,
    anomalyFlagsFor,
    type PartnerFinanceSummaryItem,
    type StoreActivityFeedEntry,
} from './utils'
import { isOverdue } from './presentation'

type OrderInsightsParams = {
    orders: Accessor<RoutedOrder[]>
    storeLabel: Accessor<string>
    activeQueueView: Accessor<QueueView>
    activeQueueSort: Accessor<QueueSort>
    operatorLens: Accessor<string>
    activityFilter: Accessor<ActivityFilter>
    hideSystemActivity: Accessor<boolean>
    setMessage: (value: string) => void
    setError: (value: string) => void
}

export function useOrderInsights(params: OrderInsightsParams) {
    const matchesQueueView = (order: RoutedOrder, view: QueueView) => {
        switch (view) {
            case 'my_queue':
                return (
                    !!params.operatorLens().trim() &&
                    order.operatorAssignee.toLowerCase() === params.operatorLens().trim().toLowerCase()
                )
            case 'overdue':
                return (
                    (order.shipmentSlaDueAt &&
                        isOverdue(order.shipmentSlaDueAt) &&
                        order.shipmentStatus !== 'delivered') ||
                    (order.issueSlaDueAt &&
                        isOverdue(order.issueSlaDueAt) &&
                        (order.exceptionStatus === 'open' ||
                            order.exceptionStatus === 'escalated' ||
                            order.shipmentStatus === 'delivery_issue'))
                )
            case 'delivery_issues':
                return order.shipmentStatus === 'delivery_issue' || order.issueResolution === 'carrier_claim'
            case 'settlement_pending':
                return order.settlementStatus === 'pending' || order.settlementStatus === 'disputed'
            case 'finance_review':
                return hasFinanceAttention(order)
            default:
                return true
        }
    }

    const filteredOrders = () => params.orders().filter((order) => matchesQueueView(order, params.activeQueueView()))

    const priorityScoreFor = (order: RoutedOrder) => {
        const shipmentOverdue =
            !!order.shipmentSlaDueAt && isOverdue(order.shipmentSlaDueAt) && order.shipmentStatus !== 'delivered'
        const issueOverdue =
            !!order.issueSlaDueAt &&
            isOverdue(order.issueSlaDueAt) &&
            (order.exceptionStatus === 'open' ||
                order.exceptionStatus === 'escalated' ||
                order.shipmentStatus === 'delivery_issue')

        if (shipmentOverdue || issueOverdue) {
            return 0
        }
        if (order.status === 'routing_blocked') {
            return 1
        }
        if (order.shipmentStatus === 'delivery_issue') {
            return 2
        }
        if (order.settlementStatus === 'disputed') {
            return 3
        }
        if (order.exceptionStatus === 'open' || order.exceptionStatus === 'escalated') {
            return 4
        }
        if (order.status === 'in_production') {
            return 5
        }
        if (order.settlementStatus === 'pending') {
            return 6
        }
        return 7
    }

    const sortedOrders = () => {
        const ranked = [...filteredOrders()]
        if (params.activeQueueSort() === 'newest') {
            return ranked.sort((a, b) => new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime())
        }
        return ranked.sort((a, b) => {
            const priorityDelta = priorityScoreFor(a) - priorityScoreFor(b)
            if (priorityDelta !== 0) {
                return priorityDelta
            }
            return new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime()
        })
    }

    const queueViewCount = (view: QueueView) => params.orders().filter((order) => matchesQueueView(order, view)).length

    const blockedOrders = () => params.orders().filter((order) => order.status === 'routing_blocked')

    const blockedReasonSummary = () => {
        const counts = new Map<string, number>()
        for (const order of blockedOrders()) {
            const key = order.routingBlockCode || 'unknown'
            counts.set(key, (counts.get(key) || 0) + 1)
        }
        return [...counts.entries()]
            .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
            .map(([code, count]) => ({ code, count }))
    }

    const forcedRerouteSummary = () => {
        const counts = new Map<string, number>()
        for (const order of params.orders()) {
            for (const activity of order.activityLog) {
                const manualReroute = activity.details.some(
                    (detail) => detail.key === 'manual_reroute' && detail.value === 'true'
                )
                if (!manualReroute) {
                    continue
                }
                const partner =
                    activity.details.find((detail) => detail.key === 'partner')?.value || order.partner || 'unknown'
                counts.set(partner, (counts.get(partner) || 0) + 1)
            }
        }
        return [...counts.entries()]
            .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
            .map(([partner, count]) => ({ partner, count }))
    }

    const reconciliationOrders = () =>
        [...params.orders()].filter(hasFinanceAttention).sort((a, b) => {
            const aDisputed = a.settlementStatus === 'disputed' ? 0 : 1
            const bDisputed = b.settlementStatus === 'disputed' ? 0 : 1
            if (aDisputed !== bDisputed) {
                return aDisputed - bDisputed
            }
            const aAnomaly = anomalyFlagsFor(a).length
            const bAnomaly = anomalyFlagsFor(b).length
            if (aAnomaly !== bAnomaly) {
                return bAnomaly - aAnomaly
            }
            return new Date(a.createdAt || 0).getTime() - new Date(b.createdAt || 0).getTime()
        })

    const partnerFinanceSummary = () => {
        const summary = new Map<string, PartnerFinanceSummaryItem>()

        for (const order of params.orders()) {
            const partner = order.partner || 'partner pending'
            const current = summary.get(partner) || {
                partner,
                orders: 0,
                pending: 0,
                disputed: 0,
                paid: 0,
                blocked: 0,
                forcedReroutes: 0,
                realizedMargin: 0,
            }
            current.orders += 1
            current.pending += order.settlementStatus === 'pending' ? 1 : 0
            current.disputed += order.settlementStatus === 'disputed' ? 1 : 0
            current.paid += order.settlementStatus === 'paid' ? 1 : 0
            current.blocked += order.status === 'routing_blocked' ? 1 : 0
            const margin = parseMoneyValue(order.realizedMargin)
            if (margin !== null) {
                current.realizedMargin += margin
            }
            for (const activity of order.activityLog) {
                const manualReroute = activity.details.some(
                    (detail) => detail.key === 'manual_reroute' && detail.value === 'true'
                )
                current.forcedReroutes += manualReroute ? 1 : 0
            }
            summary.set(partner, current)
        }

        return [...summary.values()].sort((a, b) => {
            if (b.disputed !== a.disputed) {
                return b.disputed - a.disputed
            }
            if (b.pending !== a.pending) {
                return b.pending - a.pending
            }
            return a.partner.localeCompare(b.partner)
        })
    }

    const matchesActivityFilter = (activity: RoutedOrder['activityLog'][number]) => {
        const selectedFilter = params.activityFilter()
        if (selectedFilter === 'notes') {
            return activity.type !== 'system'
        }
        if (selectedFilter !== 'all' && activity.type !== selectedFilter) {
            return false
        }
        return !(params.hideSystemActivity() && selectedFilter === 'all' && activity.type === 'system')
    }

    const filteredActivityLogFor = (order: RoutedOrder) =>
        order.activityLog.filter(matchesActivityFilter).slice().reverse().slice(0, 8)

    const hiddenSystemActivityCountFor = (order: RoutedOrder) => {
        if (params.activityFilter() !== 'all' || !params.hideSystemActivity()) {
            return 0
        }
        return order.activityLog.filter((activity) => activity.type === 'system').length
    }

    const storeActivityFeed = () =>
        sortedOrders()
            .flatMap((order) =>
                order.activityLog.filter(matchesActivityFilter).map(
                    (activity) =>
                        ({
                            orderId: order.id,
                            productTitle: order.productTitle,
                            operatorAssignee: order.operatorAssignee,
                            activity,
                        }) satisfies StoreActivityFeedEntry
                )
            )
            .sort((a, b) => new Date(b.activity.createdAt).getTime() - new Date(a.activity.createdAt).getTime())
            .slice(0, 14)

    const copyActivitySummary = async (order: RoutedOrder) => {
        const summary = formatActivitySummary(order, filteredActivityLogFor(order))
        try {
            await navigator.clipboard.writeText(summary)
            params.setMessage(`Copied activity summary for ${order.id}.`)
        } catch {
            params.setError('Could not copy activity summary to clipboard.')
        }
    }

    const copyStoreActivityFeed = async () => {
        const summary = formatStoreActivitySummary(params.storeLabel(), storeActivityFeed())
        try {
            await navigator.clipboard.writeText(summary)
            params.setMessage(`Copied store activity feed for ${params.storeLabel()}.`)
        } catch {
            params.setError('Could not copy store activity feed to clipboard.')
        }
    }

    return {
        sortedOrders,
        queueViewCount,
        blockedOrders,
        blockedReasonSummary,
        forcedRerouteSummary,
        reconciliationOrders,
        partnerFinanceSummary,
        storeActivityFeed,
        priorityScoreFor,
        filteredActivityLogFor,
        hiddenSystemActivityCountFor,
        copyActivitySummary,
        copyStoreActivityFeed,
    }
}
