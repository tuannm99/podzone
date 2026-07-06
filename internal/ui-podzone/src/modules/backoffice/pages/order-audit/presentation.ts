import type { RoutedOrderActivityFeedEntry } from '@/services/orders'

export const activityFilterOptions = [
    { name: 'All', value: 'all' },
    { name: 'Notes only', value: 'notes' },
    { name: 'System', value: 'system' },
    { name: 'Shipment', value: 'shipment_note' },
    { name: 'Settlement', value: 'settlement_note' },
    { name: 'Issue', value: 'issue_note' },
] as const

export const timeWindowOptions = [
    { name: '24h', value: '24h' },
    { name: '7 days', value: '7d' },
    { name: '30 days', value: '30d' },
    { name: 'All time', value: 'all' },
] as const

export type ActivityFilter = (typeof activityFilterOptions)[number]['value']
export type TimeWindow = (typeof timeWindowOptions)[number]['value']

export function activityColor(type: string) {
    switch (type) {
        case 'shipment_note':
            return 'indigo'
        case 'settlement_note':
            return 'green'
        case 'issue_note':
            return 'red'
        default:
            return 'dark'
    }
}

export function formatActivityTime(value: string) {
    return new Date(value).toLocaleString()
}

export function formatActivityActor(actor: string) {
    return actor.trim() || 'system'
}

export function resolveSinceIso(window: TimeWindow) {
    const durations: Partial<Record<TimeWindow, number>> = {
        '24h': 24 * 60 * 60 * 1000,
        '7d': 7 * 24 * 60 * 60 * 1000,
        '30d': 30 * 24 * 60 * 60 * 1000,
    }
    const duration = durations[window]
    return duration ? new Date(Date.now() - duration).toISOString() : undefined
}

export function formatFeedSummary(storeLabel: string, entries: RoutedOrderActivityFeedEntry[]) {
    return [
        `Store audit feed for ${storeLabel}`,
        '',
        ...entries.map((entry) => {
            const details = entry.activity.details.map((detail) => `${detail.key}=${detail.value}`).join(', ')
            return [
                `[${formatActivityTime(entry.activity.createdAt)}]`,
                entry.orderId,
                `(${entry.productTitle})`,
                `[${entry.partner}]`,
                `owner ${entry.operatorAssignee || 'unassigned'}`,
                entry.activity.type,
                `by ${formatActivityActor(entry.activity.actor)}`,
                entry.activity.message,
                details ? `(${details})` : '',
            ]
                .filter(Boolean)
                .join(' ')
        }),
    ].join('\n')
}

export function buildOrdersHref(tenantID: string, storeID: string) {
    const params = new URLSearchParams()
    if (storeID) params.set('storeId', storeID)
    const query = params.toString()
    return `/t/${tenantID}/orders${query ? `?${query}` : ''}`
}
