import type { Setter } from 'solid-js';
import {
  advanceRoutedOrder,
  bulkUpdateRoutedOrders,
  createRoutedOrder,
  forceRerouteBlockedOrder,
  openRoutedOrderException,
  type RoutedOrder,
  updateRoutedOrderExceptionStatus,
  updateRoutedOrderIssueHandling,
  updateRoutedOrderQueueControl,
  updateRoutedOrderSettlement,
  updateRoutedOrderShipment,
} from '@/services/orders';
import type { CatalogCandidate } from '@/services/productSetup';
import type { BulkDraft } from './board-context';
import type { useOrderDrafts } from './useOrderDrafts';
import {
  resolveShipmentSla,
  toIsoDateTime,
  toLocalDateTimeValue,
} from './presentation';

type Drafts = ReturnType<typeof useOrderDrafts>;

type OrderActionsParams = {
  orders: () => RoutedOrder[];
  setOrders: Setter<RoutedOrder[]>;
  availableCandidates: () => CatalogCandidate[];
  selectedCandidateId: () => string;
  customerName: () => string;
  setCustomerName: Setter<string>;
  quantity: () => string;
  setQuantity: Setter<string>;
  selectedProductType: () => string;
  selectedShipRegion: () => string;
  effectivePreferredPartner: () => string;
  resetPreferredPartnerOverride: () => void;
  selectedExceptionType: () => string;
  selectedOrderIDs: () => string[];
  setSelectedOrderIDs: Setter<string[]>;
  bulkDraft: () => BulkDraft;
  setBulkDraft: Setter<BulkDraft>;
  drafts: Drafts;
  setMessage: (value: string) => void;
  setError: (value: string) => void;
};

export function useOrderActions(params: OrderActionsParams) {
  const createMockOrder = async (event: SubmitEvent) => {
    event.preventDefault();
    const candidate = params.availableCandidates().find(
      (item) => item.id === params.selectedCandidateId()
    );
    if (!candidate) {
      params.setMessage(
        'Publish a mock product candidate first before routing orders.'
      );
      return;
    }

    params.setError('');
    const result = await createRoutedOrder({
      candidateId: candidate.id,
      customerName: params.customerName().trim(),
      quantity: Math.max(1, Number.parseInt(params.quantity(), 10) || 1),
      productType: params.selectedProductType(),
      shipRegion: params.selectedShipRegion(),
      preferredPartner: params.effectivePreferredPartner() || undefined,
    });
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) => [result.data, ...current]);
    params.setCustomerName('');
    params.setQuantity('1');
    params.resetPreferredPartnerOverride();
    params.setMessage(`Created routed order ${result.data.id}.`);
  };

  const advanceOrder = async (orderId: string) => {
    params.setError('');
    const result = await advanceRoutedOrder(orderId);
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    params.setMessage(`Advanced order ${orderId} to the next routing stage.`);
  };

  const raiseException = async (orderId: string) => {
    params.setError('');
    const result = await openRoutedOrderException(
      orderId,
      params.selectedExceptionType()
    );
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    params.setMessage(
      `Raised ${params
        .selectedExceptionType()
        .replaceAll('_', ' ')} on ${orderId}.`
    );
  };

  const updateExceptionStatus = async (orderId: string, nextStatus: string) => {
    params.setError('');
    const result = await updateRoutedOrderExceptionStatus(orderId, nextStatus);
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    params.setMessage(`Updated exception on ${orderId} to ${nextStatus}.`);
  };

  const saveShipment = async (order: RoutedOrder) => {
    params.setError('');
    const draft = params.drafts.shipmentDraftFor(order);
    const result = await updateRoutedOrderShipment(order.id, {
      shipmentStatus: draft.shipmentStatus,
      carrier: draft.shipmentCarrier.trim(),
      trackingNumber: draft.shipmentTrackingNumber.trim(),
      trackingUrl: draft.shipmentTrackingUrl.trim(),
      notes: draft.shipmentNotes.trim(),
    });
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    params.drafts.setShipmentDrafts((current) => ({
      ...current,
      [order.id]: {
        shipmentStatus: result.data.shipmentStatus,
        shipmentCarrier: result.data.shipmentCarrier,
        shipmentTrackingNumber: result.data.shipmentTrackingNumber,
        shipmentTrackingUrl: result.data.shipmentTrackingUrl,
        shipmentNotes: result.data.shipmentNotes,
      },
    }));
    params.setMessage(`Updated manual shipment control on ${order.id}.`);
  };

  const saveSettlement = async (order: RoutedOrder) => {
    params.setError('');
    const draft = params.drafts.settlementDraftFor(order);
    const result = await updateRoutedOrderSettlement(order.id, {
      fulfillmentCost: draft.fulfillmentCost.trim(),
      shippingCost: draft.shippingCost.trim(),
      settlementStatus: draft.settlementStatus,
      notes: draft.settlementNotes.trim(),
    });
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    params.drafts.setSettlementDrafts((current) => ({
      ...current,
      [order.id]: {
        fulfillmentCost: result.data.fulfillmentCost,
        shippingCost: result.data.shippingCost,
        settlementStatus: result.data.settlementStatus,
        settlementNotes: result.data.settlementNotes,
      },
    }));
    params.setMessage(`Updated settlement readiness on ${order.id}.`);
  };

  const saveIssueHandling = async (order: RoutedOrder) => {
    params.setError('');
    const draft = params.drafts.issueDraftFor(order);
    const result = await updateRoutedOrderIssueHandling(order.id, {
      issueCost: draft.issueCost.trim(),
      issueResolution: draft.issueResolution,
      notes: draft.issueNotes.trim(),
    });
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    params.drafts.setIssueDrafts((current) => ({
      ...current,
      [order.id]: {
        issueCost: result.data.issueCost,
        issueResolution: result.data.issueResolution,
        issueNotes: result.data.issueNotes,
      },
    }));
    params.setMessage(`Updated issue cost handling on ${order.id}.`);
  };

  const saveQueueControl = async (order: RoutedOrder) => {
    params.setError('');
    const draft = params.drafts.queueDraftFor(order);
    const result = await updateRoutedOrderQueueControl(order.id, {
      operatorAssignee: draft.operatorAssignee.trim() || 'unassigned',
      shipmentSlaDueAt: toIsoDateTime(draft.shipmentSlaDueAt),
      issueSlaDueAt: toIsoDateTime(draft.issueSlaDueAt),
    });
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    params.drafts.setQueueDrafts((current) => ({
      ...current,
      [order.id]: {
        operatorAssignee: result.data.operatorAssignee,
        shipmentSlaDueAt: toLocalDateTimeValue(result.data.shipmentSlaDueAt),
        issueSlaDueAt: toLocalDateTimeValue(result.data.issueSlaDueAt),
      },
    }));
    params.setMessage(`Updated queue ownership on ${order.id}.`);
  };

  const rerouteBlockedOrder = async (order: RoutedOrder) => {
    params.setError('');
    const preferredPartner = params
      .drafts
      .rerouteDraftFor(order)
      .preferredPartner.trim();
    if (!preferredPartner) {
      params.setMessage('Choose a partner before rerouting a blocked order.');
      return;
    }
    const result = await forceRerouteBlockedOrder({
      orderId: order.id,
      preferredPartner,
    });
    if (!result.success) {
      params.setError(result.message);
      return;
    }
    params.setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    params.drafts.setRerouteDrafts((current) => ({
      ...current,
      [order.id]: {
        preferredPartner: result.data.partner,
      },
    }));
    params.setMessage(`Forced reroute for ${order.id} to ${result.data.partner}.`);
  };

  const applyBulkUpdate = async () => {
    params.setError('');
    if (params.selectedOrderIDs().length === 0) {
      params.setMessage('Select at least one routed order first.');
      return;
    }
    const draft = params.bulkDraft();
    if (
      !draft.operatorAssignee.trim() &&
      !draft.shipmentSlaDueAt.trim() &&
      !draft.settlementStatus.trim()
    ) {
      params.setMessage('Choose at least one bulk field before applying.');
      return;
    }

    const result = await bulkUpdateRoutedOrders({
      orderIds: params.selectedOrderIDs(),
      operatorAssignee: draft.operatorAssignee.trim(),
      shipmentSlaDueAt:
        resolveShipmentSla(draft.shipmentSlaMode) ||
        toIsoDateTime(draft.shipmentSlaDueAt),
      settlementStatus: draft.settlementStatus.trim(),
    });
    if (!result.success) {
      params.setError(result.message);
      return;
    }

    const byID = new Map(result.data.map((order) => [order.id, order]));
    params.setOrders((current) =>
      current.map((order) => byID.get(order.id) || order)
    );
    params.setMessage(`Applied bulk update to ${result.data.length} routed orders.`);
    params.setSelectedOrderIDs([]);
    params.setBulkDraft({
      operatorAssignee: '',
      shipmentSlaDueAt: '',
      shipmentSlaMode: '',
      settlementStatus: '',
    });
  };

  return {
    createMockOrder,
    advanceOrder,
    raiseException,
    updateExceptionStatus,
    saveShipment,
    saveSettlement,
    saveIssueHandling,
    saveQueueControl,
    rerouteBlockedOrder,
    applyBulkUpdate,
  };
}
