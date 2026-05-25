import { For, Show } from 'solid-js';
import type { RoutedOrder } from '../../../../services/orders';
import {
  Badge,
  Button,
  InputField,
  SelectField,
  TextareaField,
} from '../../../components/common/Primitives';
import { formatActivityActor, formatActivityTime } from './utils';

type ActivityFilter = 'all' | 'notes' | 'system' | 'shipment_note' | 'settlement_note' | 'issue_note';
type QueueSort = 'priority' | 'newest';
type BadgeColor = 'blue' | 'indigo' | 'green' | 'yellow' | 'pink' | 'dark' | 'red';

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

type QueueDraft = {
  operatorAssignee: string;
  shipmentSlaDueAt: string;
  issueSlaDueAt: string;
};

type RerouteDraft = {
  preferredPartner: string;
};

type SelectOption = {
  name: string;
  value: string;
};

type ActivityDetail = {
  key: string;
  value: string;
};

type OrderCardActions = {
  toggleSelected: (checked: boolean) => void;
  advanceOrder: (orderId: string) => void;
  raiseException: (orderId: string) => void;
  updateExceptionStatus: (orderId: string, nextStatus: string) => void;
  rerouteBlockedOrder: (order: RoutedOrder) => void;
  saveQueueControl: (order: RoutedOrder) => void;
  saveIssueHandling: (order: RoutedOrder) => void;
  saveSettlement: (order: RoutedOrder) => void;
  saveShipment: (order: RoutedOrder) => void;
  copyActivitySummary: (order: RoutedOrder) => Promise<void>;
  queueDraftFor: (order: RoutedOrder) => QueueDraft;
  patchQueueDraft: (orderId: string, patch: Partial<QueueDraft>) => void;
  issueDraftFor: (order: RoutedOrder) => IssueDraft;
  patchIssueDraft: (orderId: string, patch: Partial<IssueDraft>) => void;
  settlementDraftFor: (order: RoutedOrder) => SettlementDraft;
  patchSettlementDraft: (
    orderId: string,
    patch: Partial<SettlementDraft>
  ) => void;
  shipmentDraftFor: (order: RoutedOrder) => ShipmentDraft;
  patchShipmentDraft: (orderId: string, patch: Partial<ShipmentDraft>) => void;
  rerouteDraftFor: (order: RoutedOrder) => RerouteDraft;
  patchRerouteDraft: (orderId: string, patch: Partial<RerouteDraft>) => void;
};

type OrderCardHelpers = {
  queueSort: QueueSort;
  priorityScoreFor: (order: RoutedOrder) => number;
  statusColor: (status: string) => BadgeColor;
  exceptionColor: (status: string) => BadgeColor;
  shipmentColor: (status: string) => BadgeColor;
  settlementColor: (status: string) => BadgeColor;
  activityColor: (type: string) => BadgeColor;
  isOverdue: (value?: string) => boolean;
  filteredActivityLogFor: (order: RoutedOrder) => {
    type: string;
    actor: string;
    createdAt: string;
    message: string;
    details: ActivityDetail[];
  }[];
  hiddenSystemActivityCountFor: (order: RoutedOrder) => number;
};

type OrderCardUi = {
  activityFilter: ActivityFilter;
  setActivityFilter: (value: ActivityFilter) => void;
  hideSystemActivity: boolean;
  toggleHideSystemActivity: () => void;
  activityFilterOptions: SelectOption[];
  shipmentOptions: SelectOption[];
  settlementOptions: SelectOption[];
  issueResolutionOptions: SelectOption[];
};

type OrderCardProps = {
  order: RoutedOrder;
  selected: boolean;
  actions: OrderCardActions;
  helpers: OrderCardHelpers;
  ui: OrderCardUi;
};

export function OrderCard(props: OrderCardProps) {
  const { order, actions, helpers, ui } = props;

  return (
    <div class="rounded-2xl border border-gray-200 bg-white p-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-start gap-3">
          <label class="mt-1">
            <input
              type="checkbox"
              checked={props.selected}
              onChange={(event) => actions.toggleSelected(event.currentTarget.checked)}
            />
          </label>
          <div>
            <p class="text-base font-semibold text-gray-900">{order.id}</p>
            <p class="mt-1 text-sm text-gray-500">
              {order.productTitle} · {order.partner || 'partner pending'}
            </p>
            <p class="mt-1 text-sm text-gray-500">
              customer {order.customerName} · qty {order.quantity} · total {order.total}
            </p>
            <p class="mt-1 text-sm text-gray-500">
              owner {order.operatorAssignee || 'unassigned'}
            </p>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <Show when={helpers.queueSort === 'priority'}>
            <Badge
              content={`priority ${helpers.priorityScoreFor(order) + 1}`}
              color={helpers.priorityScoreFor(order) < 3 ? 'red' : 'dark'}
            />
          </Show>
          <Badge
            content={order.status.replaceAll('_', ' ')}
            color={helpers.statusColor(order.status)}
          />
          <Show when={order.exceptionStatus}>
            <Badge
              content={`${order.exceptionStatus} issue`}
              color={helpers.exceptionColor(order.exceptionStatus)}
            />
          </Show>
          <Show when={order.routingBlockCode}>
            <Badge
              content={`blocked ${order.routingBlockCode.replaceAll('_', ' ')}`}
              color="red"
            />
          </Show>
          <Badge
            content={order.shipmentStatus.replaceAll('_', ' ')}
            color={helpers.shipmentColor(order.shipmentStatus)}
          />
          <Badge
            content={order.settlementStatus.replaceAll('_', ' ')}
            color={helpers.settlementColor(order.settlementStatus)}
          />
          <Button
            type="button"
            size="xs"
            color="green"
            disabled={
              order.status === 'routing_blocked' ||
              order.status === 'shipped' ||
              order.exceptionStatus === 'open' ||
              order.exceptionStatus === 'escalated'
            }
            onClick={() => {
              actions.advanceOrder(order.id);
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
              actions.raiseException(order.id);
            }}
          >
            Raise issue
          </Button>
          <Show when={order.routingBlockReason}>
            <p class="w-full text-sm text-rose-700">
              Routing blocked: {order.routingBlockReason}
            </p>
          </Show>
          <Show when={order.exceptionStatus === 'open'}>
            <Button
              type="button"
              size="xs"
              color="blue"
              onClick={() => {
                actions.updateExceptionStatus(order.id, 'escalated');
              }}
            >
              Escalate
            </Button>
            <Button
              type="button"
              size="xs"
              color="light"
              onClick={() => {
                actions.updateExceptionStatus(order.id, 'resolved');
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
            {order.exceptionType.replaceAll('_', ' ')} · {order.exceptionStatus || 'draft'}
          </p>
        </div>
      </Show>

      <Show when={order.status === 'routing_blocked'}>
        <div class="mt-3 rounded-xl border border-rose-200 bg-rose-50 p-3">
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-rose-700">
            Manual reroute
          </p>
          <p class="mt-2 text-sm text-rose-900">
            Pick an eligible partner to clear the routing block and move the order back
            into the queued lane.
          </p>
          <div class="mt-3 flex flex-col gap-3 md:flex-row md:items-end">
            <div class="flex-1">
              <InputField
                label="Preferred partner"
                value={actions.rerouteDraftFor(order).preferredPartner}
                placeholder="partner code or name"
                onInput={(event) =>
                  actions.patchRerouteDraft(order.id, {
                    preferredPartner: event.currentTarget.value,
                  })
                }
              />
            </div>
            <Button
              type="button"
              color="red"
              onClick={() => {
                actions.rerouteBlockedOrder(order);
              }}
            >
              Force reroute
            </Button>
          </div>
        </div>
      </Show>

      <div class="mt-3 rounded-xl border border-sky-200 bg-sky-50 p-3">
        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-sky-700">
          Queue ownership
        </p>
        <div class="mt-3 grid gap-4 md:grid-cols-2">
          <InputField
            label="Operator assignee"
            value={actions.queueDraftFor(order).operatorAssignee}
            placeholder="linh.nguyen"
            onInput={(event) =>
              actions.patchQueueDraft(order.id, {
                operatorAssignee: event.currentTarget.value,
              })
            }
          />
          <InputField
            label="Shipment SLA due"
            type="datetime-local"
            value={actions.queueDraftFor(order).shipmentSlaDueAt}
            onInput={(event) =>
              actions.patchQueueDraft(order.id, {
                shipmentSlaDueAt: event.currentTarget.value,
              })
            }
          />
          <InputField
            label="Issue SLA due"
            type="datetime-local"
            value={actions.queueDraftFor(order).issueSlaDueAt}
            onInput={(event) =>
              actions.patchQueueDraft(order.id, {
                issueSlaDueAt: event.currentTarget.value,
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
              actions.saveQueueControl(order);
            }}
          >
            Save queue control
          </Button>
          <Badge
            content={`owner ${order.operatorAssignee || 'unassigned'}`}
            color="indigo"
          />
          <Show when={order.shipmentSlaDueAt}>
            <Badge
              content={`shipment SLA ${helpers.isOverdue(order.shipmentSlaDueAt) ? 'overdue' : 'set'}`}
              color={helpers.isOverdue(order.shipmentSlaDueAt) ? 'red' : 'blue'}
            />
          </Show>
          <Show when={order.issueSlaDueAt}>
            <Badge
              content={`issue SLA ${helpers.isOverdue(order.issueSlaDueAt) ? 'overdue' : 'set'}`}
              color={helpers.isOverdue(order.issueSlaDueAt) ? 'red' : 'blue'}
            />
          </Show>
        </div>
      </div>

      <Show when={order.exceptionType || order.shipmentStatus === 'delivery_issue'}>
        <div class="mt-3 rounded-xl border border-rose-200 bg-rose-50 p-3">
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-rose-700">
            Issue cost handling
          </p>
          <div class="mt-3 grid gap-4 md:grid-cols-2">
            <InputField
              label="Issue cost"
              value={actions.issueDraftFor(order).issueCost}
              placeholder="$6.00"
              onInput={(event) =>
                actions.patchIssueDraft(order.id, {
                  issueCost: event.currentTarget.value,
                })
              }
            />
            <SelectField
              label="Resolution path"
              value={actions.issueDraftFor(order).issueResolution}
              options={ui.issueResolutionOptions}
              onChange={(event) =>
                actions.patchIssueDraft(order.id, {
                  issueResolution: event.currentTarget.value,
                })
              }
            />
          </div>
          <div class="mt-4">
            <TextareaField
              label="Issue notes"
              value={actions.issueDraftFor(order).issueNotes}
              rows={3}
              onInput={(event) =>
                actions.patchIssueDraft(order.id, {
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
                actions.saveIssueHandling(order);
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
            value={actions.settlementDraftFor(order).fulfillmentCost}
            placeholder="$9.50"
            onInput={(event) =>
              actions.patchSettlementDraft(order.id, {
                fulfillmentCost: event.currentTarget.value,
              })
            }
          />
          <InputField
            label="Shipping cost"
            value={actions.settlementDraftFor(order).shippingCost}
            placeholder="$4.25"
            onInput={(event) =>
              actions.patchSettlementDraft(order.id, {
                shippingCost: event.currentTarget.value,
              })
            }
          />
          <SelectField
            label="Settlement status"
            value={actions.settlementDraftFor(order).settlementStatus}
            options={ui.settlementOptions}
            onChange={(event) =>
              actions.patchSettlementDraft(order.id, {
                settlementStatus: event.currentTarget.value,
              })
            }
          />
        </div>
        <div class="mt-4">
          <TextareaField
            label="Settlement notes"
            value={actions.settlementDraftFor(order).settlementNotes}
            rows={3}
            onInput={(event) =>
              actions.patchSettlementDraft(order.id, {
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
              actions.saveSettlement(order);
            }}
          >
            Save settlement state
          </Button>
          <Badge content={`margin ${order.realizedMargin}`} color="green" />
          <Badge
            content={`settlement ${order.settlementStatus.replaceAll('_', ' ')}`}
            color={helpers.settlementColor(order.settlementStatus)}
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
            value={actions.shipmentDraftFor(order).shipmentStatus}
            options={ui.shipmentOptions}
            onChange={(event) =>
              actions.patchShipmentDraft(order.id, {
                shipmentStatus: event.currentTarget.value,
              })
            }
          />
          <InputField
            label="Carrier"
            value={actions.shipmentDraftFor(order).shipmentCarrier}
            placeholder="UPS"
            onInput={(event) =>
              actions.patchShipmentDraft(order.id, {
                shipmentCarrier: event.currentTarget.value,
              })
            }
          />
          <InputField
            label="Tracking number"
            value={actions.shipmentDraftFor(order).shipmentTrackingNumber}
            placeholder="1Z999..."
            onInput={(event) =>
              actions.patchShipmentDraft(order.id, {
                shipmentTrackingNumber: event.currentTarget.value,
              })
            }
          />
          <InputField
            label="Tracking URL"
            value={actions.shipmentDraftFor(order).shipmentTrackingUrl}
            placeholder="https://tracking.example/..."
            onInput={(event) =>
              actions.patchShipmentDraft(order.id, {
                shipmentTrackingUrl: event.currentTarget.value,
              })
            }
          />
        </div>
        <div class="mt-4">
          <TextareaField
            label="Shipment notes"
            value={actions.shipmentDraftFor(order).shipmentNotes}
            rows={3}
            onInput={(event) =>
              actions.patchShipmentDraft(order.id, {
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
              actions.saveShipment(order);
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
          <For each={order.timeline}>{(entry) => <p>{entry}</p>}</For>
        </div>
      </div>

      <div class="mt-3 rounded-xl border border-slate-200 bg-white p-3">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">
            Activity log
          </p>
          <div class="flex flex-wrap items-center gap-2">
            <div class="min-w-[11rem]">
              <SelectField
                label=""
                value={ui.activityFilter}
                options={ui.activityFilterOptions}
                onChange={(event) =>
                  ui.setActivityFilter(event.currentTarget.value as ActivityFilter)
                }
              />
            </div>
            <Show when={ui.activityFilter === 'all'}>
              <Button
                type="button"
                size="xs"
                color={ui.hideSystemActivity ? 'dark' : 'light'}
                onClick={ui.toggleHideSystemActivity}
              >
                {ui.hideSystemActivity ? 'Show system' : 'Hide system'}
              </Button>
            </Show>
            <Button
              type="button"
              size="xs"
              color="light"
              onClick={() => {
                void actions.copyActivitySummary(order);
              }}
            >
              Copy summary
            </Button>
          </div>
        </div>
        <div class="mt-3 space-y-3">
          <Show
            when={helpers.filteredActivityLogFor(order).length > 0}
            fallback={
              <div class="rounded-xl border border-dashed border-slate-200 bg-slate-50 p-3 text-sm text-slate-500">
                <Show
                  when={helpers.hiddenSystemActivityCountFor(order) > 0}
                  fallback={'No activity matches the current filter.'}
                >
                  {helpers.hiddenSystemActivityCountFor(order)} system updates are hidden.
                </Show>
              </div>
            }
          >
            <For each={helpers.filteredActivityLogFor(order)}>
              {(activity) => (
                <div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge
                        content={activity.type.replaceAll('_', ' ')}
                        color={helpers.activityColor(activity.type)}
                      />
                      <p class="text-xs font-medium text-slate-500">
                        {formatActivityActor(activity.actor)}
                      </p>
                    </div>
                    <p class="text-xs text-slate-500">
                      {formatActivityTime(activity.createdAt)}
                    </p>
                  </div>
                  <p class="mt-2 text-sm text-slate-700">{activity.message}</p>
                  <Show when={activity.details.length}>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <For each={activity.details}>
                        {(detail) => (
                          <span class="rounded-full bg-white px-2 py-1 text-[11px] font-medium text-slate-600 ring-1 ring-slate-200">
                            {detail.key.replaceAll('_', ' ')}: {detail.value}
                          </span>
                        )}
                      </For>
                    </div>
                  </Show>
                </div>
              )}
            </For>
          </Show>
        </div>
      </div>
    </div>
  );
}
