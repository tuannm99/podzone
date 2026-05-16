import { postBackofficeGraphQL } from './backofficeGraphql';

export type RoutedOrder = {
  id: string;
  candidateId: string;
  productTitle: string;
  partner: string;
  quantity: number;
  total: string;
  customerName: string;
  status: string;
  timeline: string[];
  activityLog: {
    type: string;
    actor: string;
    message: string;
    details: {
      key: string;
      value: string;
    }[];
    createdAt: string;
  }[];
  exceptionType: string;
  exceptionStatus: string;
  shipmentStatus: string;
  shipmentCarrier: string;
  shipmentTrackingNumber: string;
  shipmentTrackingUrl: string;
  shipmentNotes: string;
  operatorAssignee: string;
  shipmentSlaDueAt?: string;
  issueSlaDueAt?: string;
  baseCostSnapshot: string;
  fulfillmentCost: string;
  shippingCost: string;
  issueCost: string;
  issueResolution: string;
  issueNotes: string;
  realizedMargin: string;
  settlementStatus: string;
  settlementNotes: string;
  shippedAt?: string;
  deliveredAt?: string;
  createdAt?: string;
  updatedAt?: string;
};

export type RoutedOrderActivityFeedEntry = {
  orderId: string;
  productTitle: string;
  partner: string;
  operatorAssignee: string;
  activity: RoutedOrder['activityLog'][number];
};

export type RoutedOrderActivityFeedPage = {
  entries: RoutedOrderActivityFeedEntry[];
  total: number;
  nextCursor?: string;
};

type OrdersResult<T> =
  | { success: true; data: T }
  | { success: false; message: string };

type CreateRoutedOrderPayload = {
  candidateId: string;
  customerName: string;
  quantity: number;
};

type UpdateRoutedOrderShipmentPayload = {
  shipmentStatus: string;
  carrier: string;
  trackingNumber: string;
  trackingUrl: string;
  notes: string;
};

type UpdateRoutedOrderSettlementPayload = {
  fulfillmentCost: string;
  shippingCost: string;
  settlementStatus: string;
  notes: string;
};

type UpdateRoutedOrderIssueHandlingPayload = {
  issueCost: string;
  issueResolution: string;
  notes: string;
};

type UpdateRoutedOrderQueueControlPayload = {
  operatorAssignee: string;
  shipmentSlaDueAt?: string;
  issueSlaDueAt?: string;
};

type BulkUpdateRoutedOrdersPayload = {
  orderIds: string[];
  operatorAssignee?: string;
  shipmentSlaDueAt?: string;
  settlementStatus?: string;
};

type RoutedOrderActivityFeedQuery = {
  activityType?: string;
  actorContains?: string;
  orderId?: string;
  partner?: string;
  assignee?: string;
  since?: string;
  limit?: number;
  after?: string;
  includeSystem?: boolean;
};

const routedOrderFields = `
  id
  candidateId
  productTitle
  partner
  quantity
  total
  customerName
  status
  timeline
  activityLog {
    type
    actor
    message
    details {
      key
      value
    }
    createdAt
  }
  exceptionType
  exceptionStatus
  shipmentStatus
  shipmentCarrier
  shipmentTrackingNumber
  shipmentTrackingUrl
  shipmentNotes
  operatorAssignee
  shipmentSlaDueAt
  issueSlaDueAt
  baseCostSnapshot
  fulfillmentCost
  shippingCost
  issueCost
  issueResolution
  issueNotes
  realizedMargin
  settlementStatus
  settlementNotes
  shippedAt
  deliveredAt
  createdAt
  updatedAt
`;

export async function getRoutedOrders(): Promise<
  OrdersResult<{ orders: RoutedOrder[] }>
> {
  const result = await postBackofficeGraphQL<{ routedOrders: RoutedOrder[] }>(`
    query RoutedOrders {
      routedOrders {
${routedOrderFields}
      }
    }
  `);
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: { orders: result.data.routedOrders || [] } };
}

export async function getRoutedOrderActivities(
  input: RoutedOrderActivityFeedQuery
) : Promise<OrdersResult<RoutedOrderActivityFeedPage>> {
  const result = await postBackofficeGraphQL<{
    routedOrderActivities: RoutedOrderActivityFeedPage;
  }>(
    `
      query RoutedOrderActivities($input: RoutedOrderActivityFeedInput) {
        routedOrderActivities(input: $input) {
          entries {
            orderId
            productTitle
            partner
            operatorAssignee
            activity {
              type
              actor
              message
              details {
                key
                value
              }
              createdAt
            }
          }
          total
          nextCursor
        }
      }
    `,
    { input }
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.routedOrderActivities };
}

export async function createRoutedOrder(
  payload: CreateRoutedOrderPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{ createRoutedOrder: RoutedOrder }>(
    `
      mutation CreateRoutedOrder($input: CreateRoutedOrderInput!) {
        createRoutedOrder(input: $input) {
${routedOrderFields}
        }
      }
    `,
    { input: payload }
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.createRoutedOrder };
}

export async function advanceRoutedOrder(
  id: string
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{ advanceRoutedOrder: RoutedOrder }>(
    `
      mutation AdvanceRoutedOrder($id: ID!) {
        advanceRoutedOrder(id: $id) {
${routedOrderFields}
        }
      }
    `,
    { id }
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.advanceRoutedOrder };
}

export async function openRoutedOrderException(
  id: string,
  exceptionType: string
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{ openOrderException: RoutedOrder }>(
    `
      mutation OpenOrderException($input: OpenOrderExceptionInput!) {
        openOrderException(input: $input) {
${routedOrderFields}
        }
      }
    `,
    { input: { orderId: id, exceptionType } }
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.openOrderException };
}

export async function updateRoutedOrderExceptionStatus(
  id: string,
  status: string
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderExceptionStatus: RoutedOrder;
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
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.updateOrderExceptionStatus };
}

export async function updateRoutedOrderShipment(
  id: string,
  payload: UpdateRoutedOrderShipmentPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderShipment: RoutedOrder;
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
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.updateOrderShipment };
}

export async function updateRoutedOrderSettlement(
  id: string,
  payload: UpdateRoutedOrderSettlementPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderSettlement: RoutedOrder;
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
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.updateOrderSettlement };
}

export async function updateRoutedOrderIssueHandling(
  id: string,
  payload: UpdateRoutedOrderIssueHandlingPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderIssueHandling: RoutedOrder;
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
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.updateOrderIssueHandling };
}

export async function updateRoutedOrderQueueControl(
  id: string,
  payload: UpdateRoutedOrderQueueControlPayload
): Promise<OrdersResult<RoutedOrder>> {
  const result = await postBackofficeGraphQL<{
    updateOrderQueueControl: RoutedOrder;
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
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.updateOrderQueueControl };
}

export async function bulkUpdateRoutedOrders(
  payload: BulkUpdateRoutedOrdersPayload
): Promise<OrdersResult<RoutedOrder[]>> {
  const result = await postBackofficeGraphQL<{
    bulkUpdateRoutedOrders: RoutedOrder[];
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
  );
  if (!result.success) {
    return { success: false, message: result.message };
  }
  return { success: true, data: result.data.bulkUpdateRoutedOrders || [] };
}
