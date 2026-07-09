import type { RoutedOrder } from '@podzone/shared/services/orders'

export type PartnerFinanceRow = {
    partner: string
    orders: number
    pending: number
    disputed: number
    paid: number
    blocked: number
    forcedReroutes: number
    realizedMargin: number
}

export function parseMoneyValue(value: string) {
    const trimmed = value.trim()
    if (!trimmed || trimmed === 'TBD') return null
    const numeric = Number.parseFloat(trimmed.replace(/[^0-9.]/g, ''))
    if (Number.isNaN(numeric)) return null
    return trimmed.includes('-') ? -numeric : numeric
}

export function formatMoney(value: number) {
    return `$${value.toFixed(2)}`
}

export function formatFlag(value: string) {
    return value.replaceAll('_', ' ')
}

export function anomalyFlagsFor(order: RoutedOrder) {
    const flags: string[] = []
    const realizedMargin = parseMoneyValue(order.realizedMargin)
    const fulfillmentCost = parseMoneyValue(order.fulfillmentCost)
    const baseCostSnapshot = parseMoneyValue(order.baseCostSnapshot)
    const shippingCost = parseMoneyValue(order.shippingCost)
    const issueCost = parseMoneyValue(order.issueCost)
    if (realizedMargin !== null && realizedMargin < 0) flags.push('negative_margin')
    if (fulfillmentCost !== null && baseCostSnapshot !== null && fulfillmentCost > baseCostSnapshot)
        flags.push('fulfillment_above_snapshot')
    if (shippingCost !== null && shippingCost >= 8) flags.push('high_shipping_cost')
    if (issueCost !== null && issueCost > 0) flags.push('issue_cost_present')
    if (order.settlementStatus === 'disputed') flags.push('settlement_disputed')
    return flags
}

export function hasFinanceAttention(order: RoutedOrder) {
    return (
        order.settlementStatus === 'pending' ||
        order.settlementStatus === 'disputed' ||
        anomalyFlagsFor(order).length > 0
    )
}

export function settlementColor(status: string) {
    switch (status) {
        case 'paid':
            return 'green'
        case 'reconciled':
            return 'blue'
        case 'disputed':
            return 'red'
        default:
            return 'yellow'
    }
}

export function buildFinanceQueueHref(tenantID: string, storeID: string) {
    const params = new URLSearchParams({
        queueView: 'finance_review',
        queueSort: 'priority',
    })
    if (storeID) params.set('storeId', storeID)
    return `/t/${tenantID}/orders?${params.toString()}`
}
