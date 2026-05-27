import { For, Show } from 'solid-js';
import { Badge } from '../../../components/common/Primitives';
import { useTenantOrdersInsights } from './context';
import { anomalyFlagsFor, formatAnomalyLabel, formatBlockLabel, parseMoneyValue } from './utils';

export function OrdersInsightsPanel() {
  const insights = useTenantOrdersInsights();

  return (
    <>
      <div class="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-4">
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <p class="text-sm font-semibold text-rose-900">
              Routing blocked insights
            </p>
            <p class="mt-1 text-sm text-rose-700">
              {insights.blockedOrders().length} blocked order(s) are waiting for
              operator intervention.
            </p>
            <div class="mt-3 flex flex-wrap gap-2">
              <Show
                when={insights.blockedReasonSummary().length > 0}
                fallback={
                  <p class="text-sm text-rose-700">
                    No blocked reason codes in the queue right now.
                  </p>
                }
              >
                <For each={insights.blockedReasonSummary()}>
                  {(item) => (
                    <Badge
                      content={`${formatBlockLabel(item.code)} · ${item.count}`}
                      color="red"
                    />
                  )}
                </For>
              </Show>
            </div>
          </div>
          <div>
            <p class="text-sm font-semibold text-slate-900">
              Forced reroute frequency
            </p>
            <p class="mt-1 text-sm text-slate-600">
              Spot which partners are being manually chosen to clear blocked
              orders.
            </p>
            <div class="mt-3 flex flex-wrap gap-2">
              <Show
                when={insights.forcedRerouteSummary().length > 0}
                fallback={
                  <p class="text-sm text-slate-600">
                    No forced reroutes recorded yet.
                  </p>
                }
              >
                <For each={insights.forcedRerouteSummary().slice(0, 6)}>
                  {(item) => (
                    <Badge
                      content={`${item.partner} · ${item.count}`}
                      color="indigo"
                    />
                  )}
                </For>
              </Show>
            </div>
          </div>
        </div>
      </div>
      <div class="mt-4 grid gap-4 xl:grid-cols-[0.95fr_1.05fr]">
        <div class="rounded-lg border border-emerald-200 bg-emerald-50 p-4">
          <p class="text-sm font-semibold text-emerald-900">
            Settlement reconciliation queue
          </p>
          <p class="mt-1 text-sm text-emerald-700">
            Review pending, disputed, and anomalous orders before partner payout
            closes.
          </p>
          <div class="mt-3 space-y-3">
            <Show
              when={insights.reconciliationOrders().length > 0}
              fallback={
                <p class="text-sm text-emerald-700">
                  No finance review orders are waiting right now.
                </p>
              }
            >
              <For each={insights.reconciliationOrders().slice(0, 6)}>
                {(order) => (
                  <div class="rounded-md border border-emerald-100 bg-white p-3">
                    <div class="flex flex-wrap items-center justify-between gap-2">
                      <div>
                        <p class="text-sm font-semibold text-slate-900">
                          {order.id}
                        </p>
                        <p class="text-xs text-slate-500">
                          {order.productTitle} · {order.partner || 'partner pending'}
                        </p>
                      </div>
                      <div class="flex flex-wrap gap-2">
                        <Badge
                          content={order.settlementStatus.replaceAll('_', ' ')}
                          color={order.settlementStatus === 'disputed' ? 'red' : 'yellow'}
                        />
                        <Badge
                          content={`margin ${order.realizedMargin}`}
                          color={
                            (parseMoneyValue(order.realizedMargin) || 0) < 0
                              ? 'red'
                              : 'green'
                          }
                        />
                      </div>
                    </div>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <For each={anomalyFlagsFor(order)}>
                        {(flag) => (
                          <Badge
                            content={formatAnomalyLabel(flag)}
                            color="red"
                          />
                        )}
                      </For>
                    </div>
                  </div>
                )}
              </For>
            </Show>
          </div>
        </div>
        <div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
          <p class="text-sm font-semibold text-slate-900">
            Partner finance snapshot
          </p>
          <p class="mt-1 text-sm text-slate-600">
            Track settlement pressure, realized margin, and reroute load per
            partner.
          </p>
          <div class="mt-3 space-y-3">
            <Show
              when={insights.partnerFinanceSummary().length > 0}
              fallback={
                <p class="text-sm text-slate-600">
                  No partner finance data yet.
                </p>
              }
            >
              <For each={insights.partnerFinanceSummary().slice(0, 6)}>
                {(item) => (
                  <div class="rounded-md border border-slate-200 bg-white p-3">
                    <div class="flex flex-wrap items-center justify-between gap-2">
                      <p class="text-sm font-semibold text-slate-900">
                        {item.partner}
                      </p>
                      <Badge
                        content={`margin $${item.realizedMargin.toFixed(2)}`}
                        color={item.realizedMargin < 0 ? 'red' : 'green'}
                      />
                    </div>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <Badge content={`${item.orders} orders`} color="dark" />
                      <Badge content={`${item.pending} pending`} color="yellow" />
                      <Badge content={`${item.disputed} disputed`} color="red" />
                      <Badge content={`${item.paid} paid`} color="green" />
                      <Show when={item.blocked > 0}>
                        <Badge content={`${item.blocked} blocked`} color="red" />
                      </Show>
                      <Show when={item.forcedReroutes > 0}>
                        <Badge
                          content={`${item.forcedReroutes} forced reroutes`}
                          color="indigo"
                        />
                      </Show>
                    </div>
                  </div>
                )}
              </For>
            </Show>
          </div>
        </div>
      </div>
    </>
  );
}

