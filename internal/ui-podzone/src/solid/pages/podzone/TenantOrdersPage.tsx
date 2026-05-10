import { useParams } from '@tanstack/solid-router';
import { For, Show, createEffect, createSignal } from 'solid-js';
import {
  advanceRoutedOrder,
  createRoutedOrder,
  getRoutedOrders,
  openRoutedOrderException,
  type RoutedOrder,
  updateRoutedOrderExceptionStatus,
  updateRoutedOrderIssueHandling,
  updateRoutedOrderSettlement,
  updateRoutedOrderShipment,
} from '../../../services/orders';
import {
  getProductSetupSnapshot,
  type CatalogCandidate,
} from '../../../services/productSetup';
import { tenantStorage } from '../../../services/tenantStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import {
  Badge,
  Button,
  Card,
  InputField,
  SelectField,
  TextareaField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

const routeStatuses = [
  { name: 'Queued', value: 'queued' },
  { name: 'In production', value: 'in_production' },
  { name: 'Shipped', value: 'shipped' },
];

const exceptionOptions = [
  { name: 'Artwork issue', value: 'artwork_issue' },
  { name: 'Partner delay', value: 'partner_delay' },
  { name: 'Address hold', value: 'address_hold' },
  { name: 'Reprint request', value: 'reprint_request' },
];

const shipmentOptions = [
  { name: 'Awaiting label', value: 'awaiting_label' },
  { name: 'Label ready', value: 'label_ready' },
  { name: 'In transit', value: 'in_transit' },
  { name: 'Delivered', value: 'delivered' },
  { name: 'Delivery issue', value: 'delivery_issue' },
];

const settlementOptions = [
  { name: 'Pending', value: 'pending' },
  { name: 'Reconciled', value: 'reconciled' },
  { name: 'Paid', value: 'paid' },
  { name: 'Disputed', value: 'disputed' },
];

const issueResolutionOptions = [
  { name: 'Monitor', value: 'monitor' },
  { name: 'Reprint', value: 'reprint' },
  { name: 'Refund', value: 'refund' },
  { name: 'Carrier claim', value: 'carrier_claim' },
  { name: 'Address retry', value: 'address_retry' },
];

type ShipmentDraft = {
  shipmentStatus: string;
  shipmentCarrier: string;
  shipmentTrackingNumber: string;
  shipmentTrackingUrl: string;
  shipmentNotes: string;
};

type SettlementDraft = {
  fulfillmentCost: string;
  shippingCost: string;
  settlementStatus: string;
  settlementNotes: string;
};

type IssueDraft = {
  issueCost: string;
  issueResolution: string;
  issueNotes: string;
};

function statusColor(status: string) {
  switch (status) {
    case 'shipped':
      return 'green';
    case 'in_production':
      return 'blue';
    default:
      return 'yellow';
  }
}

function exceptionColor(status: string) {
  switch (status) {
    case 'resolved':
      return 'green';
    case 'escalated':
      return 'red';
    case 'open':
      return 'yellow';
    default:
      return 'dark';
  }
}

function shipmentColor(status: string) {
  switch (status) {
    case 'delivered':
      return 'green';
    case 'in_transit':
      return 'blue';
    case 'delivery_issue':
      return 'red';
    case 'label_ready':
      return 'indigo';
    default:
      return 'dark';
  }
}

function settlementColor(status: string) {
  switch (status) {
    case 'paid':
      return 'green';
    case 'reconciled':
      return 'blue';
    case 'disputed':
      return 'red';
    default:
      return 'yellow';
  }
}

export default function TenantOrdersPage() {
  const params = useParams({ from: '/t/$tenantId/orders' });

  const [availableCandidates, setAvailableCandidates] = createSignal<
    CatalogCandidate[]
  >([]);
  const [orders, setOrders] = createSignal<RoutedOrder[]>([]);
  const [selectedCandidateId, setSelectedCandidateId] = createSignal('');
  const [customerName, setCustomerName] = createSignal('');
  const [quantity, setQuantity] = createSignal('1');
  const [selectedExceptionType, setSelectedExceptionType] =
    createSignal('artwork_issue');
  const [shipmentDrafts, setShipmentDrafts] = createSignal<
    Record<string, ShipmentDraft>
  >({});
  const [settlementDrafts, setSettlementDrafts] = createSignal<
    Record<string, SettlementDraft>
  >({});
  const [issueDrafts, setIssueDrafts] = createSignal<Record<string, IssueDraft>>(
    {}
  );
  const [message, setMessage] = createSignal('');
  const [error, setError] = createSignal('');

  const shipmentDraftFor = (order: RoutedOrder): ShipmentDraft =>
    shipmentDrafts()[order.id] || {
      shipmentStatus: order.shipmentStatus || 'awaiting_label',
      shipmentCarrier: order.shipmentCarrier || '',
      shipmentTrackingNumber: order.shipmentTrackingNumber || '',
      shipmentTrackingUrl: order.shipmentTrackingUrl || '',
      shipmentNotes: order.shipmentNotes || '',
    };

  const settlementDraftFor = (order: RoutedOrder): SettlementDraft =>
    settlementDrafts()[order.id] || {
      fulfillmentCost: order.fulfillmentCost || order.baseCostSnapshot || '$0.00',
      shippingCost: order.shippingCost || '$0.00',
      settlementStatus: order.settlementStatus || 'pending',
      settlementNotes: order.settlementNotes || '',
    };

  const issueDraftFor = (order: RoutedOrder): IssueDraft =>
    issueDrafts()[order.id] || {
      issueCost: order.issueCost || '$0.00',
      issueResolution: order.issueResolution || 'monitor',
      issueNotes: order.issueNotes || '',
    };

  const patchShipmentDraft = (orderId: string, patch: Partial<ShipmentDraft>) => {
    setShipmentDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          shipmentStatus: 'awaiting_label',
          shipmentCarrier: '',
          shipmentTrackingNumber: '',
          shipmentTrackingUrl: '',
          shipmentNotes: '',
        }),
        ...patch,
      },
    }));
  };

  const patchSettlementDraft = (
    orderId: string,
    patch: Partial<SettlementDraft>
  ) => {
    setSettlementDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          fulfillmentCost: '$0.00',
          shippingCost: '$0.00',
          settlementStatus: 'pending',
          settlementNotes: '',
        }),
        ...patch,
      },
    }));
  };

  const patchIssueDraft = (orderId: string, patch: Partial<IssueDraft>) => {
    setIssueDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          issueCost: '$0.00',
          issueResolution: 'monitor',
          issueNotes: '',
        }),
        ...patch,
      },
    }));
  };

  const loadCandidates = async () => {
    const result = await getProductSetupSnapshot();
    if (!result.success) {
      setError(result.message);
      setAvailableCandidates([]);
      return;
    }
    const published = result.data.candidates.filter(
      (candidate) => candidate.status === 'published_mock'
    );
    setAvailableCandidates(published);
    if (!selectedCandidateId() && published.length > 0) {
      setSelectedCandidateId(published[0].id);
    }
  };

  const loadOrders = async () => {
    const result = await getRoutedOrders();
    if (!result.success) {
      setError(result.message);
      setOrders([]);
      return;
    }
    setOrders(result.data.orders);
  };

  const createMockOrder = async (event: SubmitEvent) => {
    event.preventDefault();
    const candidate = availableCandidates().find(
      (item) => item.id === selectedCandidateId()
    );
    if (!candidate) {
      setMessage('Publish a mock product candidate first before routing orders.');
      return;
    }

    setError('');
    const result = await createRoutedOrder({
      candidateId: candidate.id,
      customerName: customerName().trim(),
      quantity: Math.max(1, Number.parseInt(quantity(), 10) || 1),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) => [result.data, ...current]);
    setCustomerName('');
    setQuantity('1');
    setMessage(`Created routed order ${result.data.id}.`);
  };

  const advanceOrder = async (orderId: string) => {
    setError('');
    const result = await advanceRoutedOrder(orderId);
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    setMessage(`Advanced order ${orderId} to the next routing stage.`);
  };

  const raiseException = async (orderId: string) => {
    setError('');
    const result = await openRoutedOrderException(
      orderId,
      selectedExceptionType()
    );
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    setMessage(`Raised ${selectedExceptionType().replaceAll('_', ' ')} on ${orderId}.`);
  };

  const updateExceptionStatus = async (orderId: string, nextStatus: string) => {
    setError('');
    const result = await updateRoutedOrderExceptionStatus(orderId, nextStatus);
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    setMessage(`Updated exception on ${orderId} to ${nextStatus}.`);
  };

  const saveShipment = async (order: RoutedOrder) => {
    setError('');
    const draft = shipmentDraftFor(order);
    const result = await updateRoutedOrderShipment(order.id, {
      shipmentStatus: draft.shipmentStatus,
      carrier: draft.shipmentCarrier.trim(),
      trackingNumber: draft.shipmentTrackingNumber.trim(),
      trackingUrl: draft.shipmentTrackingUrl.trim(),
      notes: draft.shipmentNotes.trim(),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setShipmentDrafts((current) => ({
      ...current,
      [order.id]: {
        shipmentStatus: result.data.shipmentStatus,
        shipmentCarrier: result.data.shipmentCarrier,
        shipmentTrackingNumber: result.data.shipmentTrackingNumber,
        shipmentTrackingUrl: result.data.shipmentTrackingUrl,
        shipmentNotes: result.data.shipmentNotes,
      },
    }));
    setMessage(`Updated manual shipment control on ${order.id}.`);
  };

  const saveSettlement = async (order: RoutedOrder) => {
    setError('');
    const draft = settlementDraftFor(order);
    const result = await updateRoutedOrderSettlement(order.id, {
      fulfillmentCost: draft.fulfillmentCost.trim(),
      shippingCost: draft.shippingCost.trim(),
      settlementStatus: draft.settlementStatus,
      notes: draft.settlementNotes.trim(),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setSettlementDrafts((current) => ({
      ...current,
      [order.id]: {
        fulfillmentCost: result.data.fulfillmentCost,
        shippingCost: result.data.shippingCost,
        settlementStatus: result.data.settlementStatus,
        settlementNotes: result.data.settlementNotes,
      },
    }));
    setMessage(`Updated settlement readiness on ${order.id}.`);
  };

  const saveIssueHandling = async (order: RoutedOrder) => {
    setError('');
    const draft = issueDraftFor(order);
    const result = await updateRoutedOrderIssueHandling(order.id, {
      issueCost: draft.issueCost.trim(),
      issueResolution: draft.issueResolution,
      notes: draft.issueNotes.trim(),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setIssueDrafts((current) => ({
      ...current,
      [order.id]: {
        issueCost: result.data.issueCost,
        issueResolution: result.data.issueResolution,
        issueNotes: result.data.issueNotes,
      },
    }));
    setMessage(`Updated issue cost handling on ${order.id}.`);
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    void loadCandidates();
    void loadOrders();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Order Routing Workspace"
          title={`POD routing board for store ${params().tenantId}`}
          copy="This routing workspace persists store-scoped POD orders in the backend. Published mock products can be routed through production, shipment, issue handling, and settlement readiness."
        />
      </Card>

      <Show when={message()}>
        <InfoAlert>{message()}</InfoAlert>
      </Show>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <InfoAlert>
        Orders and published product candidates now come from backend store data. Shipment and settlement control both stay manual on this board so the store team can manage POD execution directly.
      </InfoAlert>

      <div class="grid gap-6 lg:grid-cols-[0.96fr_1.04fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create routed order"
            subtitle="Use a published mock product candidate as the source, then send the order into the backend-backed POD routing flow."
          />

          <Show
            when={availableCandidates().length > 0}
            fallback={
              <EmptyBlock
                title="No published mock products yet"
                copy="Go to Product setup, promote a draft, and mock publish it from the backend-backed setup workflow before testing order routing."
              />
            }
          >
            <form class="space-y-4" onSubmit={createMockOrder}>
              <SelectField
                label="Published mock product"
                value={selectedCandidateId()}
                options={availableCandidates().map((candidate) => ({
                  name: `${candidate.title} · ${candidate.partner}`,
                  value: candidate.id,
                }))}
                onChange={(event) =>
                  setSelectedCandidateId(event.currentTarget.value)
                }
              />
              <div class="grid gap-4 md:grid-cols-2">
                <InputField
                  label="Customer name"
                  value={customerName()}
                  placeholder="Nguyen Minh"
                  onInput={(event) => setCustomerName(event.currentTarget.value)}
                />
                <InputField
                  label="Quantity"
                  value={quantity()}
                  placeholder="1"
                  onInput={(event) => setQuantity(event.currentTarget.value)}
                />
              </div>
              <SelectField
                label="Default exception scenario"
                value={selectedExceptionType()}
                options={exceptionOptions}
                onChange={(event) =>
                  setSelectedExceptionType(event.currentTarget.value)
                }
              />
              <Button type="submit" color="blue">
                Create routed order
              </Button>
            </form>
          </Show>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Routing board"
            subtitle="Watch each order move from intake to production, then manage shipment and settlement state directly inside the store-scoped POD workflow."
          />

          <Show
            when={orders().length > 0}
            fallback={
              <EmptyBlock
                title="No routed orders yet"
                copy="Create a routed order on the left to test store-side routing, manual shipment control, and settlement readiness."
              />
            }
          >
            <div class="space-y-3">
              <For each={orders()}>
                {(order) => (
                  <div class="rounded-2xl border border-gray-200 bg-white p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="text-base font-semibold text-gray-900">
                          {order.id}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          {order.productTitle} · {order.partner}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          customer {order.customerName} · qty {order.quantity} · total {order.total}
                        </p>
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge
                          content={order.status.replaceAll('_', ' ')}
                          color={statusColor(order.status)}
                        />
                        <Show when={order.exceptionStatus}>
                          <Badge
                            content={`${order.exceptionStatus} issue`}
                            color={exceptionColor(order.exceptionStatus)}
                          />
                        </Show>
                        <Badge
                          content={order.shipmentStatus.replaceAll('_', ' ')}
                          color={shipmentColor(order.shipmentStatus)}
                        />
                        <Badge
                          content={order.settlementStatus.replaceAll('_', ' ')}
                          color={settlementColor(order.settlementStatus)}
                        />
                        <Button
                          type="button"
                          size="xs"
                          color="green"
                          disabled={
                            order.status === 'shipped' ||
                            order.exceptionStatus === 'open' ||
                            order.exceptionStatus === 'escalated'
                          }
                          onClick={() => {
                            advanceOrder(order.id);
                          }}
                        >
                          Advance route
                        </Button>
                        <Button
                          type="button"
                          size="xs"
                          color="alternative"
                          disabled={
                            order.exceptionStatus === 'open' ||
                            order.exceptionStatus === 'resolved'
                          }
                          onClick={() => {
                            raiseException(order.id);
                          }}
                        >
                          Raise issue
                        </Button>
                        <Show when={order.exceptionStatus === 'open'}>
                          <Button
                            type="button"
                            size="xs"
                            color="blue"
                            onClick={() => {
                              updateExceptionStatus(order.id, 'escalated');
                            }}
                          >
                            Escalate
                          </Button>
                          <Button
                            type="button"
                            size="xs"
                            color="light"
                            onClick={() => {
                              updateExceptionStatus(order.id, 'resolved');
                            }}
                          >
                            Resolve
                          </Button>
                        </Show>
                      </div>
                    </div>

                    <Show when={order.exceptionType}>
                      <div class="mt-3 rounded-xl border border-amber-200 bg-amber-50 p-3">
                        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-amber-700">
                          Exception
                        </p>
                        <p class="mt-2 text-sm text-amber-900">
                          {order.exceptionType.replaceAll('_', ' ')} ·{' '}
                          {order.exceptionStatus || 'draft'}
                        </p>
                      </div>
                    </Show>

                    <Show
                      when={
                        order.exceptionType ||
                        order.shipmentStatus === 'delivery_issue'
                      }
                    >
                      <div class="mt-3 rounded-xl border border-rose-200 bg-rose-50 p-3">
                        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-rose-700">
                          Issue cost handling
                        </p>
                        <div class="mt-3 grid gap-4 md:grid-cols-2">
                          <InputField
                            label="Issue cost"
                            value={issueDraftFor(order).issueCost}
                            placeholder="$6.00"
                            onInput={(event) =>
                              patchIssueDraft(order.id, {
                                issueCost: event.currentTarget.value,
                              })
                            }
                          />
                          <SelectField
                            label="Resolution path"
                            value={issueDraftFor(order).issueResolution}
                            options={issueResolutionOptions}
                            onChange={(event) =>
                              patchIssueDraft(order.id, {
                                issueResolution: event.currentTarget.value,
                              })
                            }
                          />
                        </div>
                        <div class="mt-4">
                          <TextareaField
                            label="Issue notes"
                            value={issueDraftFor(order).issueNotes}
                            rows={3}
                            onInput={(event) =>
                              patchIssueDraft(order.id, {
                                issueNotes: event.currentTarget.value,
                              })
                            }
                          />
                        </div>
                        <div class="mt-3 flex flex-wrap items-center gap-2">
                          <Button
                            type="button"
                            size="xs"
                            color="red"
                            onClick={() => {
                              saveIssueHandling(order);
                            }}
                          >
                            Save issue handling
                          </Button>
                          <Badge content={`cost ${order.issueCost}`} color="red" />
                          <Badge
                            content={order.issueResolution.replaceAll('_', ' ')}
                            color="yellow"
                          />
                        </div>
                      </div>
                    </Show>

                    <div class="mt-3 rounded-xl border border-emerald-200 bg-emerald-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">
                        Settlement readiness
                      </p>
                      <div class="mt-3 grid gap-3 md:grid-cols-2">
                        <div class="rounded-xl border border-emerald-200 bg-white p-3">
                          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">
                            Base cost snapshot
                          </p>
                          <p class="mt-2 text-sm font-semibold text-gray-900">
                            {order.baseCostSnapshot}
                          </p>
                        </div>
                        <div class="rounded-xl border border-emerald-200 bg-white p-3">
                          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">
                            Realized margin
                          </p>
                          <p class="mt-2 text-sm font-semibold text-gray-900">
                            {order.realizedMargin}
                          </p>
                        </div>
                      </div>
                      <div class="mt-3 grid gap-4 md:grid-cols-2">
                        <InputField
                          label="Fulfillment cost"
                          value={settlementDraftFor(order).fulfillmentCost}
                          placeholder="$9.50"
                          onInput={(event) =>
                            patchSettlementDraft(order.id, {
                              fulfillmentCost: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Shipping cost"
                          value={settlementDraftFor(order).shippingCost}
                          placeholder="$4.25"
                          onInput={(event) =>
                            patchSettlementDraft(order.id, {
                              shippingCost: event.currentTarget.value,
                            })
                          }
                        />
                        <SelectField
                          label="Settlement status"
                          value={settlementDraftFor(order).settlementStatus}
                          options={settlementOptions}
                          onChange={(event) =>
                            patchSettlementDraft(order.id, {
                              settlementStatus: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-4">
                        <TextareaField
                          label="Settlement notes"
                          value={settlementDraftFor(order).settlementNotes}
                          rows={3}
                          onInput={(event) =>
                            patchSettlementDraft(order.id, {
                              settlementNotes: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-3 flex flex-wrap items-center gap-2">
                        <Button
                          type="button"
                          size="xs"
                          color="green"
                          onClick={() => {
                            saveSettlement(order);
                          }}
                        >
                          Save settlement state
                        </Button>
                        <Badge
                          content={`margin ${order.realizedMargin}`}
                          color="green"
                        />
                        <Badge
                          content={`settlement ${order.settlementStatus.replaceAll('_', ' ')}`}
                          color={settlementColor(order.settlementStatus)}
                        />
                      </div>
                    </div>

                    <div class="mt-3 rounded-xl border border-slate-200 bg-slate-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">
                        Manual shipment control
                      </p>
                      <div class="mt-3 grid gap-4 md:grid-cols-2">
                        <SelectField
                          label="Shipment status"
                          value={shipmentDraftFor(order).shipmentStatus}
                          options={shipmentOptions}
                          onChange={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentStatus: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Carrier"
                          value={shipmentDraftFor(order).shipmentCarrier}
                          placeholder="UPS"
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentCarrier: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Tracking number"
                          value={shipmentDraftFor(order).shipmentTrackingNumber}
                          placeholder="1Z999..."
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentTrackingNumber: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Tracking URL"
                          value={shipmentDraftFor(order).shipmentTrackingUrl}
                          placeholder="https://tracking.example/..."
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentTrackingUrl: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-4">
                        <TextareaField
                          label="Shipment notes"
                          value={shipmentDraftFor(order).shipmentNotes}
                          rows={3}
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentNotes: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-3 flex flex-wrap items-center gap-2">
                        <Button
                          type="button"
                          size="xs"
                          color="blue"
                          onClick={() => {
                            saveShipment(order);
                          }}
                        >
                          Save shipment state
                        </Button>
                        <Show when={order.shipmentCarrier || order.shipmentTrackingNumber}>
                          <Badge
                            content={`${order.shipmentCarrier || 'manual carrier'} ${order.shipmentTrackingNumber || ''}`.trim()}
                            color="indigo"
                          />
                        </Show>
                        <Show when={order.deliveredAt}>
                          <Badge content="Delivered confirmed" color="green" />
                        </Show>
                      </div>
                    </div>

                    <div class="mt-3 rounded-xl bg-gray-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
                        Timeline
                      </p>
                      <div class="mt-2 space-y-1 text-sm text-gray-600">
                        <For each={order.timeline}>
                          {(entry) => <p>{entry}</p>}
                        </For>
                      </div>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </Card>
      </div>

      <Card class="mt-6 space-y-4">
        <SectionTitle
          title="Routing, shipment, and settlement stages"
          subtitle="Production routing, shipment control, and settlement updates are all managed inside the workspace, so operators can run POD execution manually without relying on external fulfillment callbacks."
        />
        <div class="flex flex-wrap gap-2">
          <For each={routeStatuses}>
            {(stage) => (
              <Badge content={stage.name} color={statusColor(stage.value)} />
            )}
          </For>
          <Badge content="Open issue" color="yellow" />
          <Badge content="Escalated issue" color="red" />
          <Badge content="Resolved issue" color="green" />
          <For each={shipmentOptions}>
            {(stage) => (
              <Badge content={stage.name} color={shipmentColor(stage.value)} />
            )}
          </For>
          <For each={settlementOptions}>
            {(stage) => (
              <Badge
                content={`Settlement ${stage.name}`}
                color={settlementColor(stage.value)}
              />
            )}
          </For>
          <For each={issueResolutionOptions}>
            {(stage) => (
              <Badge
                content={`Issue ${stage.name}`}
                color={stage.value === 'reprint' || stage.value === 'refund' ? 'red' : 'yellow'}
              />
            )}
          </For>
        </div>
      </Card>
    </PageShell>
  );
}
