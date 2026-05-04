import { useParams } from '@tanstack/solid-router';
import { For, Show, createEffect, createSignal } from 'solid-js';
import {
  getProductSetupSnapshot,
  type CatalogCandidate,
} from '../../../services/productSetup';
import { tenantStorage } from '../../../services/tenantStorage';
import { EmptyBlock, ErrorAlert, InfoAlert } from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import {
  Badge,
  Button,
  Card,
  InputField,
  SelectField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

type MockOrder = {
  id: string;
  candidateId: string;
  productTitle: string;
  partner: string;
  quantity: number;
  total: string;
  customerName: string;
  status: string;
  timeline: string[];
  exceptionType: string;
  exceptionStatus: string;
};

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

function ordersStorageKey(tenantId: string) {
  return `podzone:mock-orders:${tenantId}`;
}

function parseMoney(value: string) {
  const cleaned = value.replace(/[^0-9.]/g, '');
  const parsed = Number.parseFloat(cleaned);
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatMoney(value: number) {
  return `$${value.toFixed(2)}`;
}

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

function nextRouteStatus(status: string) {
  if (status === 'queued') return 'in_production';
  if (status === 'in_production') return 'shipped';
  return 'shipped';
}

function timelineEntry(status: string, partner: string) {
  switch (status) {
    case 'in_production':
      return `Sent to ${partner} for POD production`;
    case 'shipped':
      return `Marked as shipped from ${partner}`;
    default:
      return `Queued for ${partner}`;
  }
}

export default function TenantOrdersPage() {
  const params = useParams({ from: '/t/$tenantId/orders' });

  const [availableCandidates, setAvailableCandidates] = createSignal<
    CatalogCandidate[]
  >([]);
  const [orders, setOrders] = createSignal<MockOrder[]>([]);
  const [selectedCandidateId, setSelectedCandidateId] = createSignal('');
  const [customerName, setCustomerName] = createSignal('');
  const [quantity, setQuantity] = createSignal('1');
  const [selectedExceptionType, setSelectedExceptionType] =
    createSignal('artwork_issue');
  const [message, setMessage] = createSignal('');
  const [error, setError] = createSignal('');

  const persistOrders = (nextOrders: MockOrder[]) => {
    localStorage.setItem(
      ordersStorageKey(params().tenantId),
      JSON.stringify(nextOrders)
    );
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

  const loadOrders = () => {
    const raw = localStorage.getItem(ordersStorageKey(params().tenantId));
    if (!raw) {
      const seeded: MockOrder[] = [];
      setOrders(seeded);
      persistOrders(seeded);
      return;
    }

    try {
      setOrders(JSON.parse(raw) as MockOrder[]);
    } catch {
      setOrders([]);
    }
  };

  const createMockOrder = (event: SubmitEvent) => {
    event.preventDefault();
    const candidate = availableCandidates().find(
      (item) => item.id === selectedCandidateId()
    );
    if (!candidate) {
      setMessage('Publish a mock product candidate first before routing orders.');
      return;
    }

    const qty = Math.max(1, Number.parseInt(quantity(), 10) || 1);
    const total = formatMoney(parseMoney(candidate.retailPrice) * qty);
    const nextOrder: MockOrder = {
      id: `ORD-${Date.now().toString().slice(-6)}`,
      candidateId: candidate.id,
      productTitle: candidate.title,
      partner: candidate.partner,
      quantity: qty,
      total,
      customerName: customerName().trim() || 'Sample customer',
      status: 'queued',
      timeline: [
        `Order created for ${candidate.title}`,
        timelineEntry('queued', candidate.partner),
      ],
      exceptionType: '',
      exceptionStatus: '',
    };

    const nextOrders = [nextOrder, ...orders()];
    setOrders(nextOrders);
    persistOrders(nextOrders);
    setCustomerName('');
    setQuantity('1');
    setMessage(`Created routed mock order ${nextOrder.id}.`);
  };

  const advanceOrder = (orderId: string) => {
    const nextOrders = orders().map((order) => {
      if (order.id !== orderId) return order;
      const nextStatus = nextRouteStatus(order.status);
      return {
        ...order,
        status: nextStatus,
        timeline: [...order.timeline, timelineEntry(nextStatus, order.partner)],
      };
    });
    setOrders(nextOrders);
    persistOrders(nextOrders);
    setMessage(`Advanced order ${orderId} to the next routing stage.`);
  };

  const raiseException = (orderId: string) => {
    const nextOrders = orders().map((order) => {
      if (order.id !== orderId) return order;
      if (order.exceptionStatus === 'open') return order;
      return {
        ...order,
        exceptionType: selectedExceptionType(),
        exceptionStatus: 'open',
        timeline: [
          ...order.timeline,
          `Exception opened: ${selectedExceptionType().replaceAll('_', ' ')}`,
        ],
      };
    });
    setOrders(nextOrders);
    persistOrders(nextOrders);
    setMessage(`Raised ${selectedExceptionType().replaceAll('_', ' ')} on ${orderId}.`);
  };

  const updateExceptionStatus = (orderId: string, nextStatus: string) => {
    const nextOrders = orders().map((order) => {
      if (order.id !== orderId) return order;
      if (!order.exceptionType) return order;
      return {
        ...order,
        exceptionStatus: nextStatus,
        timeline: [
          ...order.timeline,
          `Exception ${nextStatus}: ${order.exceptionType.replaceAll('_', ' ')}`,
        ],
      };
    });
    setOrders(nextOrders);
    persistOrders(nextOrders);
    setMessage(`Updated exception on ${orderId} to ${nextStatus}.`);
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    void loadCandidates();
    loadOrders();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Order Routing Prototype"
          title={`Mock order routing for store ${params().tenantId}`}
          copy="This is a browser-local routing workspace for POD operations. It uses published prototype products and moves them through early fulfillment states without relying on external systems."
        />
      </Card>

      <Show when={message()}>
        <InfoAlert>{message()}</InfoAlert>
      </Show>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <InfoAlert>
        This screen stores prototype orders in the browser only. It now reads published product candidates from the backend-backed Product setup workflow.
      </InfoAlert>

      <div class="grid gap-6 lg:grid-cols-[0.96fr_1.04fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create prototype routed order"
            subtitle="Use a published prototype product candidate as the source, then send the order into the local routing flow."
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
            subtitle="Watch each order move from intake to production and shipment in the local POD workflow."
          />

          <Show
            when={orders().length > 0}
            fallback={
              <EmptyBlock
                title="No routed orders yet"
                copy="Create a mock order on the left to test store-side routing and POD partner handoff."
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
          title="Routing stages"
          subtitle="The current mock flow now includes exception handling so the team can validate operator language and decisions before deeper architecture work."
        />
        <div class="flex flex-wrap gap-2">
          <For each={routeStatuses}>
            {(stage) => (
              <Badge
                content={stage.name}
                color={statusColor(stage.value)}
              />
            )}
          </For>
          <Badge content="Open issue" color="yellow" />
          <Badge content="Escalated issue" color="red" />
          <Badge content="Resolved issue" color="green" />
        </div>
      </Card>
    </PageShell>
  );
}
