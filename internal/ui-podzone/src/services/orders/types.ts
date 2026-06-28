export type RoutedOrder = {
  id: string
  candidateId: string
  productTitle: string
  partner: string
  quantity: number
  total: string
  customerName: string
  status: string
  timeline: string[]
  activityLog: {
    type: string
    actor: string
    message: string
    details: {
      key: string
      value: string
    }[]
    createdAt: string
  }[]
  exceptionType: string
  exceptionStatus: string
  shipmentStatus: string
  shipmentCarrier: string
  shipmentTrackingNumber: string
  shipmentTrackingUrl: string
  shipmentNotes: string
  operatorAssignee: string
  shipmentSlaDueAt?: string
  issueSlaDueAt?: string
  routingBlockCode: string
  routingBlockReason: string
  baseCostSnapshot: string
  fulfillmentCost: string
  shippingCost: string
  issueCost: string
  issueResolution: string
  issueNotes: string
  realizedMargin: string
  settlementStatus: string
  settlementNotes: string
  shippedAt?: string
  deliveredAt?: string
  createdAt?: string
  updatedAt?: string
}

export type RoutedOrderActivityFeedEntry = {
  orderId: string
  productTitle: string
  partner: string
  operatorAssignee: string
  activity: RoutedOrder['activityLog'][number]
}

export type RoutedOrderActivityFeedPage = {
  entries: RoutedOrderActivityFeedEntry[]
  total: number
  nextCursor?: string
}

export type OrdersResult<T> =
  | { success: true; data: T }
  | { success: false; message: string }

export type CreateRoutedOrderPayload = {
  candidateId: string
  customerName: string
  quantity: number
  productType: string
  shipRegion: string
  preferredPartner?: string
}

export type PartnerRoutingProfile = {
  id: string
  code: string
  name: string
  partnerType: string
  status: string
  supportedProductTypes: string[]
  supportedRegions: string[]
  slaDays: number
  routingPriority: number
  baseFulfillmentCost: string
  shippingCostRules: {
    region: string
    cost: string
  }[]
}

export type RoutingPartnerOption = {
  partner: PartnerRoutingProfile
  eligible: boolean
  reason: string
  estimatedFulfillmentCost: string
  estimatedShippingCost: string
  estimatedUnitMargin: string
}

export type RoutedOrderRecommendation = {
  candidateId: string
  productTitle: string
  candidatePartner: string
  productType: string
  shipRegion: string
  selectedPartner: string
  blockedReasonCode: string
  blockedReason: string
  summary: string
  options: RoutingPartnerOption[]
}

export type UpdateRoutedOrderShipmentPayload = {
  shipmentStatus: string
  carrier: string
  trackingNumber: string
  trackingUrl: string
  notes: string
}

export type UpdateRoutedOrderSettlementPayload = {
  fulfillmentCost: string
  shippingCost: string
  settlementStatus: string
  notes: string
}

export type UpdateRoutedOrderIssueHandlingPayload = {
  issueCost: string
  issueResolution: string
  notes: string
}

export type UpdateRoutedOrderQueueControlPayload = {
  operatorAssignee: string
  shipmentSlaDueAt?: string
  issueSlaDueAt?: string
}

export type ForceRerouteBlockedOrderPayload = {
  orderId: string
  preferredPartner: string
}

export type BulkUpdateRoutedOrdersPayload = {
  orderIds: string[]
  operatorAssignee?: string
  shipmentSlaDueAt?: string
  settlementStatus?: string
}

export type RoutedOrderActivityFeedQuery = {
  activityType?: string
  actorContains?: string
  orderId?: string
  partner?: string
  assignee?: string
  since?: string
  limit?: number
  after?: string
  includeSystem?: boolean
}

export type RoutedOrderRecommendationQuery = {
  candidateId: string
  productType: string
  shipRegion: string
  preferredPartner?: string
}
