import type { RoutedOrder } from '@/services/orders'

export type StoreActivityFeedEntry = {
  orderId: string
  productTitle: string
  operatorAssignee: string
  activity: RoutedOrder['activityLog'][number]
}

export type PartnerFinanceSummaryItem = {
  partner: string
  orders: number
  pending: number
  disputed: number
  paid: number
  blocked: number
  forcedReroutes: number
  realizedMargin: number
}

export function formatActivityTime(value: string) {
  return new Date(value).toLocaleString()
}

export function formatActivityActor(actor: string) {
  const normalized = actor.trim()
  if (!normalized) {
    return 'system'
  }
  return normalized
}

export function formatActivitySummary(
  order: RoutedOrder,
  activities: RoutedOrder['activityLog']
) {
  const header = [
    `Order ${order.id}`,
    `Product: ${order.productTitle}`,
    `Operator: ${order.operatorAssignee || 'unassigned'}`,
    `Route status: ${order.status}`,
    `Shipment: ${order.shipmentStatus}`,
    `Settlement: ${order.settlementStatus}`,
  ]

  const activityLines = activities.map((activity) => {
    const details = activity.details
      .map((detail) => `${detail.key}=${detail.value}`)
      .join(', ')
    return [
      `[${formatActivityTime(activity.createdAt)}]`,
      activity.type,
      `by ${formatActivityActor(activity.actor)}`,
      activity.message,
      details ? `(${details})` : '',
    ]
      .filter(Boolean)
      .join(' ')
  })

  return [...header, '', 'Recent activity:', ...activityLines].join('\n')
}

export function formatStoreActivitySummary(
  tenantId: string,
  entries: StoreActivityFeedEntry[]
) {
  const lines = entries.map((entry) => {
    const details = entry.activity.details
      .map((detail) => `${detail.key}=${detail.value}`)
      .join(', ')
    return [
      `[${formatActivityTime(entry.activity.createdAt)}]`,
      entry.orderId,
      `(${entry.productTitle})`,
      `owner ${entry.operatorAssignee || 'unassigned'}`,
      entry.activity.type,
      `by ${formatActivityActor(entry.activity.actor)}`,
      entry.activity.message,
      details ? `(${details})` : '',
    ]
      .filter(Boolean)
      .join(' ')
  })

  return [`Store activity feed for ${tenantId}`, '', ...lines].join('\n')
}

export function formatBlockLabel(value: string) {
  return value.replaceAll('_', ' ')
}

export function parseMoneyValue(value: string) {
  const trimmed = value.trim()
  if (!trimmed || trimmed === 'TBD') {
    return null
  }
  const negative = trimmed.includes('-')
  const numeric = Number.parseFloat(trimmed.replace(/[^0-9.]/g, ''))
  if (Number.isNaN(numeric)) {
    return null
  }
  return negative ? -numeric : numeric
}

export function formatAnomalyLabel(value: string) {
  return value.replaceAll('_', ' ')
}

export function anomalyFlagsFor(order: RoutedOrder) {
  const flags: string[] = []
  const realizedMargin = parseMoneyValue(order.realizedMargin)
  const fulfillmentCost = parseMoneyValue(order.fulfillmentCost)
  const baseCostSnapshot = parseMoneyValue(order.baseCostSnapshot)
  const shippingCost = parseMoneyValue(order.shippingCost)
  const issueCost = parseMoneyValue(order.issueCost)

  if (realizedMargin !== null && realizedMargin < 0) {
    flags.push('negative_margin')
  }
  if (
    fulfillmentCost !== null &&
    baseCostSnapshot !== null &&
    fulfillmentCost > baseCostSnapshot
  ) {
    flags.push('fulfillment_above_snapshot')
  }
  if (shippingCost !== null && shippingCost >= 8) {
    flags.push('high_shipping_cost')
  }
  if (issueCost !== null && issueCost > 0) {
    flags.push('issue_cost_present')
  }
  if (order.settlementStatus === 'disputed') {
    flags.push('settlement_disputed')
  }
  return flags
}

export function hasFinanceAttention(order: RoutedOrder) {
  return (
    order.settlementStatus === 'pending' ||
    order.settlementStatus === 'disputed' ||
    anomalyFlagsFor(order).length > 0
  )
}
