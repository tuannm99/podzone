import { postBackofficeGraphQL } from '../backofficeGraphql'
import { routedOrderFields } from './graphql'
import type {
  BulkUpdateRoutedOrdersPayload,
  CreateRoutedOrderPayload,
  ForceRerouteBlockedOrderPayload,
  OrdersResult,
  RoutedOrder,
  UpdateRoutedOrderIssueHandlingPayload,
  UpdateRoutedOrderQueueControlPayload,
  UpdateRoutedOrderSettlementPayload,
  UpdateRoutedOrderShipmentPayload,
} from './types'

export async function createRoutedOrder(
  payload: CreateRoutedOrderPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    createRoutedOrder: RoutedOrder
  }>(
    `
      mutation CreateRoutedOrder($input: CreateRoutedOrderInput!) {
        createRoutedOrder(input: $input) {
${routedOrderFields}
        }
      }
    `,
    { input: payload }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.createRoutedOrder }
}

export async function advanceRoutedOrder(
  id: string
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    advanceRoutedOrder: RoutedOrder
  }>(
    `
      mutation AdvanceRoutedOrder($id: ID!) {
        advanceRoutedOrder(id: $id) {
${routedOrderFields}
        }
      }
    `,
    { id }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.advanceRoutedOrder }
}

export async function forceRerouteBlockedOrder(
  payload: ForceRerouteBlockedOrderPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    forceRerouteBlockedOrder: RoutedOrder
  }>(
    `
      mutation ForceRerouteBlockedOrder($input: ForceRerouteBlockedOrderInput!) {
        forceRerouteBlockedOrder(input: $input) {
${routedOrderFields}
        }
      }
    `,
    { input: payload }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.forceRerouteBlockedOrder }
}

export async function openRoutedOrderException(
  id: string,
  exceptionType: string
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    openOrderException: RoutedOrder
  }>(
    `
      mutation OpenOrderException($input: OpenOrderExceptionInput!) {
        openOrderException(input: $input) {
${routedOrderFields}
        }
      }
    `,
    { input: { orderId: id, exceptionType } }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.openOrderException }
}

export async function updateRoutedOrderExceptionStatus(
  id: string,
  status: string
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderExceptionStatus: RoutedOrder
  }>(
    `
      mutation UpdateOrderExceptionStatus(
        $input: UpdateOrderExceptionStatusInput!
      ) {
        updateOrderExceptionStatus(input: $input) {
${routedOrderFields}
        }
      }
    `,
    { input: { orderId: id, status } }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.updateOrderExceptionStatus }
}

export async function updateRoutedOrderShipment(
  id: string,
  payload: UpdateRoutedOrderShipmentPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderShipment: RoutedOrder
  }>(
    `
      mutation UpdateOrderShipment($input: UpdateOrderShipmentInput!) {
        updateOrderShipment(input: $input) {
${routedOrderFields}
        }
      }
    `,
    {
      input: {
        orderId: id,
        shipmentStatus: payload.shipmentStatus,
        carrier: payload.carrier,
        trackingNumber: payload.trackingNumber,
        trackingUrl: payload.trackingUrl,
        notes: payload.notes,
      },
    }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.updateOrderShipment }
}

export async function updateRoutedOrderSettlement(
  id: string,
  payload: UpdateRoutedOrderSettlementPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderSettlement: RoutedOrder
  }>(
    `
      mutation UpdateOrderSettlement($input: UpdateOrderSettlementInput!) {
        updateOrderSettlement(input: $input) {
${routedOrderFields}
        }
      }
    `,
    {
      input: {
        orderId: id,
        fulfillmentCost: payload.fulfillmentCost,
        shippingCost: payload.shippingCost,
        settlementStatus: payload.settlementStatus,
        notes: payload.notes,
      },
    }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.updateOrderSettlement }
}

export async function updateRoutedOrderIssueHandling(
  id: string,
  payload: UpdateRoutedOrderIssueHandlingPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderIssueHandling: RoutedOrder
  }>(
    `
      mutation UpdateOrderIssueHandling($input: UpdateOrderIssueHandlingInput!) {
        updateOrderIssueHandling(input: $input) {
${routedOrderFields}
        }
      }
    `,
    {
      input: {
        orderId: id,
        issueCost: payload.issueCost,
        issueResolution: payload.issueResolution,
        notes: payload.notes,
      },
    }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.updateOrderIssueHandling }
}

export async function updateRoutedOrderQueueControl(
  id: string,
  payload: UpdateRoutedOrderQueueControlPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderQueueControl: RoutedOrder
  }>(
    `
      mutation UpdateOrderQueueControl($input: UpdateOrderQueueControlInput!) {
        updateOrderQueueControl(input: $input) {
${routedOrderFields}
        }
      }
    `,
    {
      input: {
        orderId: id,
        operatorAssignee: payload.operatorAssignee,
        shipmentSlaDueAt: payload.shipmentSlaDueAt || null,
        issueSlaDueAt: payload.issueSlaDueAt || null,
      },
    }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.updateOrderQueueControl }
}

export async function bulkUpdateRoutedOrders(
  payload: BulkUpdateRoutedOrdersPayload
): Promise<OrdersResult<RoutedOrder[]>> {
  const result = await postBackofficeGraphQL<{
    bulkUpdateRoutedOrders: RoutedOrder[]
  }>(
    `
      mutation BulkUpdateRoutedOrders($input: BulkUpdateRoutedOrdersInput!) {
        bulkUpdateRoutedOrders(input: $input) {
${routedOrderFields}
        }
      }
    `,
    {
      input: {
        orderIds: payload.orderIds,
        operatorAssignee:
          payload.operatorAssignee && payload.operatorAssignee.trim()
            ? payload.operatorAssignee
            : null,
        shipmentSlaDueAt: payload.shipmentSlaDueAt || null,
        settlementStatus:
          payload.settlementStatus && payload.settlementStatus.trim()
            ? payload.settlementStatus
            : null,
      },
    }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.bulkUpdateRoutedOrders || [] }
}
