import { createEffect, createSignal, type Accessor } from 'solid-js'
import { getRoutedOrders, type RoutedOrder } from '@/services/orders'
import { tenantStorage } from '@/services/tenantStorage'
import {
    anomalyFlagsFor,
    formatMoney,
    hasFinanceAttention,
    parseMoneyValue,
    type PartnerFinanceRow,
} from './presentation'

interface OrderFinanceViewModelOptions {
    tenantID: Accessor<string>
    storeID: Accessor<string>
    storeLabel: Accessor<string>
    workspaceReady: Accessor<boolean>
}

export function createOrderFinanceViewModel(options: OrderFinanceViewModelOptions) {
    const [orders, setOrders] = createSignal<RoutedOrder[]>([])
    const [message, setMessage] = createSignal('')
    const [error, setError] = createSignal('')

    const reload = async () => {
        const result = await getRoutedOrders()
        if (!result.success) {
            setError(result.message)
            setOrders([])
            return
        }
        setError('')
        setOrders(result.data.orders)
    }
    const financeOrders = () =>
        [...orders()].filter(hasFinanceAttention).sort((left, right) => {
            const disputedOrder =
                Number(left.settlementStatus !== 'disputed') - Number(right.settlementStatus !== 'disputed')
            return (
                disputedOrder ||
                anomalyFlagsFor(right).length - anomalyFlagsFor(left).length ||
                new Date(left.createdAt || 0).getTime() - new Date(right.createdAt || 0).getTime()
            )
        })
    const partnerFinanceSummary = () => {
        const summary = new Map<string, PartnerFinanceRow>()
        for (const order of orders()) {
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
            if (order.settlementStatus === 'pending') current.pending += 1
            if (order.settlementStatus === 'disputed') current.disputed += 1
            if (order.settlementStatus === 'paid') current.paid += 1
            if (order.status === 'routing_blocked') current.blocked += 1
            current.realizedMargin += parseMoneyValue(order.realizedMargin) || 0
            current.forcedReroutes += order.activityLog.filter((activity) =>
                activity.details.some((detail) => detail.key === 'manual_reroute' && detail.value === 'true')
            ).length
            summary.set(partner, current)
        }
        return [...summary.values()].sort(
            (left, right) =>
                right.disputed - left.disputed ||
                right.pending - left.pending ||
                left.partner.localeCompare(right.partner)
        )
    }
    const totalRealizedMargin = () =>
        formatMoney(orders().reduce((sum, order) => sum + (parseMoneyValue(order.realizedMargin) || 0), 0))
    const issueExposure = () =>
        formatMoney(orders().reduce((sum, order) => sum + (parseMoneyValue(order.issueCost) || 0), 0))
    const copySummary = async () => {
        const lines = [
            `Finance review for ${options.tenantID()}`,
            `Store: ${options.storeLabel()} (${options.storeID() || 'pending'})`,
            `Orders needing attention: ${financeOrders().length}`,
            `Realized margin: ${totalRealizedMargin()}`,
            `Issue exposure: ${issueExposure()}`,
            '',
            ...partnerFinanceSummary()
                .slice(0, 8)
                .map((item) =>
                    [
                        item.partner,
                        `orders=${item.orders}`,
                        `pending=${item.pending}`,
                        `disputed=${item.disputed}`,
                        `paid=${item.paid}`,
                        `blocked=${item.blocked}`,
                        `forced_reroutes=${item.forcedReroutes}`,
                        `margin=${formatMoney(item.realizedMargin)}`,
                    ].join(' ')
                ),
        ].join('\n')
        try {
            await navigator.clipboard.writeText(lines)
            setMessage(`Copied finance summary for ${options.storeLabel()}.`)
        } catch {
            setError('Could not copy finance summary to clipboard.')
        }
    }

    createEffect(() => {
        tenantStorage.setTenantID(options.tenantID())
        if (options.workspaceReady()) void reload()
    })

    return {
        orders,
        message,
        error,
        financeOrders,
        partnerFinanceSummary,
        totalRealizedMargin,
        issueExposure,
        copySummary,
        reload,
    }
}
