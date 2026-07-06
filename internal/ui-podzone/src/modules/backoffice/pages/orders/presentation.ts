import type { QueueSort, QueueView, ShipmentSlaMode } from './board-context'
import type { ActivityFilter, BadgeColor } from './order-card/types'

export const routeStatuses = [
    { name: 'Routing blocked', value: 'routing_blocked' },
    { name: 'Queued', value: 'queued' },
    { name: 'In production', value: 'in_production' },
    { name: 'Shipped', value: 'shipped' },
]

export const shipmentOptions = [
    { name: 'Awaiting label', value: 'awaiting_label' },
    { name: 'Label ready', value: 'label_ready' },
    { name: 'In transit', value: 'in_transit' },
    { name: 'Delivered', value: 'delivered' },
    { name: 'Delivery issue', value: 'delivery_issue' },
]

export const settlementOptions = [
    { name: 'Pending', value: 'pending' },
    { name: 'Reconciled', value: 'reconciled' },
    { name: 'Paid', value: 'paid' },
    { name: 'Disputed', value: 'disputed' },
]

export const issueResolutionOptions = [
    { name: 'Monitor', value: 'monitor' },
    { name: 'Reprint', value: 'reprint' },
    { name: 'Refund', value: 'refund' },
    { name: 'Carrier claim', value: 'carrier_claim' },
    { name: 'Address retry', value: 'address_retry' },
]

export const activityFilterOptions = [
    { name: 'All', value: 'all' },
    { name: 'Notes only', value: 'notes' },
    { name: 'System', value: 'system' },
    { name: 'Shipment', value: 'shipment_note' },
    { name: 'Settlement', value: 'settlement_note' },
    { name: 'Issue', value: 'issue_note' },
] satisfies { name: string; value: ActivityFilter }[]

export function statusColor(status: string): BadgeColor {
    switch (status) {
        case 'routing_blocked':
            return 'red'
        case 'shipped':
            return 'green'
        case 'in_production':
            return 'blue'
        default:
            return 'yellow'
    }
}

export function exceptionColor(status: string): BadgeColor {
    switch (status) {
        case 'resolved':
            return 'green'
        case 'escalated':
            return 'red'
        case 'open':
            return 'yellow'
        default:
            return 'dark'
    }
}

export function shipmentColor(status: string): BadgeColor {
    switch (status) {
        case 'delivered':
            return 'green'
        case 'in_transit':
            return 'blue'
        case 'delivery_issue':
            return 'red'
        case 'label_ready':
            return 'indigo'
        default:
            return 'dark'
    }
}

export function settlementColor(status: string): BadgeColor {
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

export function activityColor(type: string): BadgeColor {
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

export function toLocalDateTimeValue(value?: string) {
    if (!value) {
        return ''
    }
    const date = new Date(value)
    const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
    return local.toISOString().slice(0, 16)
}

export function toIsoDateTime(value: string) {
    if (!value.trim()) {
        return ''
    }
    return new Date(value).toISOString()
}

export function isOverdue(value?: string) {
    if (!value) {
        return false
    }
    return new Date(value).getTime() < Date.now()
}

export function isQueueView(value: string): value is QueueView {
    return (
        value === 'all' ||
        value === 'my_queue' ||
        value === 'overdue' ||
        value === 'delivery_issues' ||
        value === 'settlement_pending' ||
        value === 'finance_review'
    )
}

export function isQueueSort(value: string): value is QueueSort {
    return value === 'priority' || value === 'newest'
}

export function scopedStorageKey(prefix: string, tenantID: string, storeID: string) {
    return `${prefix}:${tenantID}:${storeID || 'pending'}`
}

export function queuePresetStorageKey(tenantID: string, storeID: string) {
    return scopedStorageKey('podzone:orders:queue-presets', tenantID, storeID)
}

export function bulkTemplateStorageKey(tenantID: string, storeID: string) {
    return scopedStorageKey('podzone:orders:bulk-templates', tenantID, storeID)
}

export function resolveShipmentSla(mode: ShipmentSlaMode) {
    if (!mode) {
        return ''
    }
    const now = new Date()
    if (mode === 'plus_2h') {
        return new Date(now.getTime() + 2 * 60 * 60 * 1000).toISOString()
    }
    if (mode === 'plus_4h') {
        return new Date(now.getTime() + 4 * 60 * 60 * 1000).toISOString()
    }
    const endOfDay = new Date(now)
    endOfDay.setHours(23, 59, 0, 0)
    return endOfDay.toISOString()
}
