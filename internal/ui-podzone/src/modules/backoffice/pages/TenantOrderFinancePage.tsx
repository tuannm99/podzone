import { useParams } from '@tanstack/solid-router'
import { For, Show, createEffect, createSignal } from 'solid-js'
import { getRoutedOrders, type RoutedOrder } from '@/services/orders'
import { tenantStorage } from '@/services/tenantStorage'
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
} from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { StatCard } from '@/solid/components/dashboard/StatCard'
import { useTenantWorkspace } from '@/solid/workspace/context'

function parseMoneyValue(value: string) {
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

function formatMoney(value: number) {
  return `$${value.toFixed(2)}`
}

function formatFlag(value: string) {
  return value.replaceAll('_', ' ')
}

function anomalyFlagsFor(order: RoutedOrder) {
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

function hasFinanceAttention(order: RoutedOrder) {
  return (
    order.settlementStatus === 'pending' ||
    order.settlementStatus === 'disputed' ||
    anomalyFlagsFor(order).length > 0
  )
}

function settlementColor(status: string) {
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

type PartnerFinanceRow = {
  partner: string
  orders: number
  pending: number
  disputed: number
  paid: number
  blocked: number
  forcedReroutes: number
  realizedMargin: number
}

function buildFinanceQueueHref(tenantID: string, storeID: string) {
  const params = new URLSearchParams({
    queueView: 'finance_review',
    queueSort: 'priority',
  })
  if (storeID) {
    params.set('storeId', storeID)
  }
  return `/t/${tenantID}/orders?${params.toString()}`
}

export default function TenantOrderFinancePage() {
  const params = useParams({ from: '/t/$tenantId/orders/finance' })
  const workspace = useTenantWorkspace()

  const [orders, setOrders] = createSignal<RoutedOrder[]>([])
  const [message, setMessage] = createSignal('')
  const [error, setError] = createSignal('')
  const currentStoreId = () => workspace?.currentStoreId() || ''
  const currentStore = () => workspace?.currentStore()
  const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
  const storeLabel = () =>
    currentStore()?.name || currentStoreId() || 'selected store'

  const loadOrders = async () => {
    const result = await getRoutedOrders()
    if (!result.success) {
      setError(result.message)
      setOrders([])
      return
    }
    setError('')
    setOrders(result.data.orders)
  }

  const financeOrders = () =>
    [...orders()].filter(hasFinanceAttention).sort((a, b) => {
      const aDisputed = a.settlementStatus === 'disputed' ? 0 : 1
      const bDisputed = b.settlementStatus === 'disputed' ? 0 : 1
      if (aDisputed !== bDisputed) {
        return aDisputed - bDisputed
      }
      return (
        anomalyFlagsFor(b).length - anomalyFlagsFor(a).length ||
        new Date(a.createdAt || 0).getTime() -
          new Date(b.createdAt || 0).getTime()
      )
    })

  const partnerFinanceSummary = () => {
    const summary = new Map<string, PartnerFinanceRow>()
    for (const order of orders()) {
      const partner = order.partner || 'partner pending'
      const current = summary.get(partner) || {
        partner,
        orders: 0,
        pending: 0,
        disputed: 0,
        paid: 0,
        blocked: 0,
        forcedReroutes: 0,
        realizedMargin: 0,
      }
      current.orders += 1
      if (order.settlementStatus === 'pending') {
        current.pending += 1
      }
      if (order.settlementStatus === 'disputed') {
        current.disputed += 1
      }
      if (order.settlementStatus === 'paid') {
        current.paid += 1
      }
      if (order.status === 'routing_blocked') {
        current.blocked += 1
      }
      const margin = parseMoneyValue(order.realizedMargin)
      if (margin !== null) {
        current.realizedMargin += margin
      }
      for (const activity of order.activityLog) {
        const manualReroute = activity.details.some(
          (detail) => detail.key === 'manual_reroute' && detail.value === 'true'
        )
        if (manualReroute) {
          current.forcedReroutes += 1
        }
      }
      summary.set(partner, current)
    }
    return [...summary.values()].sort((a, b) => {
      return (
        b.disputed - a.disputed ||
        b.pending - a.pending ||
        a.partner.localeCompare(b.partner)
      )
    })
  }

  const totalRealizedMargin = () =>
    formatMoney(
      orders().reduce(
        (sum, order) => sum + (parseMoneyValue(order.realizedMargin) || 0),
        0
      )
    )

  const issueExposure = () =>
    formatMoney(
      orders().reduce(
        (sum, order) => sum + (parseMoneyValue(order.issueCost) || 0),
        0
      )
    )

  const copySummary = async () => {
    const lines = [
      `Finance review for ${params().tenantId}`,
      `Store: ${storeLabel()} (${currentStoreId() || 'pending'})`,
      `Orders needing attention: ${financeOrders().length}`,
      `Realized margin: ${totalRealizedMargin()}`,
      `Issue exposure: ${issueExposure()}`,
      '',
      ...partnerFinanceSummary()
        .slice(0, 8)
        .map((item) =>
          [
            item.partner,
            `orders=${item.orders}`,
            `pending=${item.pending}`,
            `disputed=${item.disputed}`,
            `paid=${item.paid}`,
            `blocked=${item.blocked}`,
            `forced_reroutes=${item.forcedReroutes}`,
            `margin=${formatMoney(item.realizedMargin)}`,
          ].join(' ')
        ),
    ].join('\n')
    try {
      await navigator.clipboard.writeText(lines)
      setMessage(`Copied finance summary for ${storeLabel()}.`)
    } catch {
      setError('Could not copy finance summary to clipboard.')
    }
  }

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId)
    if (!workspaceReady()) {
      return
    }
    void loadOrders()
  })

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Settlement Finance"
          title={`Finance review for ${storeLabel()}`}
          copy="Track pending and disputed settlements, margin anomalies, issue cost exposure, and partner payout pressure from a single operations lane."
        />
      </Card>

      <Show when={message()}>
        <InfoAlert>{message()}</InfoAlert>
      </Show>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <Show when={!workspaceReady()}>
        <EmptyBlock
          title="Choose a store first"
          copy="Use the workspace store switcher before loading store-scoped finance review data."
        />
      </Show>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Needs review" value={String(financeOrders().length)} />
        <StatCard label="Realized margin" value={totalRealizedMargin()} />
        <StatCard label="Issue exposure" value={issueExposure()} />
        <StatCard
          label="Disputed settlements"
          value={String(
            orders().filter((order) => order.settlementStatus === 'disputed')
              .length
          )}
        />
      </div>

      <Card class="space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <SectionTitle
            title="Reconciliation queue"
            subtitle="Orders here need follow-up before partner payout closes or margin leaks further."
          />
          <div class="flex flex-wrap gap-2">
            <Button
              type="button"
              size="xs"
              color="light"
              onClick={() => {
                void copySummary()
              }}
            >
              Copy summary
            </Button>
            <Button
              type="button"
              size="xs"
              color="alternative"
              href={buildFinanceQueueHref(params().tenantId, currentStoreId())}
            >
              Open queue board
            </Button>
          </div>
        </div>
        <Show
          when={financeOrders().length > 0}
          fallback={
            <EmptyBlock
              title="No finance review orders"
              copy="This store currently has no pending, disputed, or anomalous routed orders."
            />
          }
        >
          <div class="space-y-3">
            <For each={financeOrders()}>
              {(order) => (
                <div class="rounded-lg border border-slate-200 bg-white p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-slate-900">
                        {order.id}
                      </p>
                      <p class="text-xs text-slate-500">
                        {order.productTitle} ·{' '}
                        {order.partner || 'partner pending'}
                      </p>
                    </div>
                    <div class="flex flex-wrap gap-2">
                      <Badge
                        content={order.settlementStatus.replaceAll('_', ' ')}
                        color={settlementColor(order.settlementStatus)}
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
                  <div class="mt-3 flex flex-wrap gap-2">
                    <For each={anomalyFlagsFor(order)}>
                      {(flag) => (
                        <Badge content={formatFlag(flag)} color="red" />
                      )}
                    </For>
                  </div>
                  <Show when={order.settlementNotes}>
                    <p class="mt-3 text-sm text-slate-600">
                      {order.settlementNotes}
                    </p>
                  </Show>
                </div>
              )}
            </For>
          </div>
        </Show>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Partner finance snapshot"
          subtitle="Use this to spot who is accumulating payout pressure, blocked work, or too many forced reroutes."
        />
        <Show
          when={partnerFinanceSummary().length > 0}
          fallback={
            <EmptyBlock
              title="No partner finance data yet"
              copy="Create routed orders and update settlements to start building partner finance history."
            />
          }
        >
          <div class="grid gap-3 xl:grid-cols-2">
            <For each={partnerFinanceSummary()}>
              {(item) => (
                <div class="rounded-lg border border-slate-200 bg-white p-4">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <p class="text-sm font-semibold text-slate-900">
                      {item.partner}
                    </p>
                    <Badge
                      content={`margin ${formatMoney(item.realizedMargin)}`}
                      color={item.realizedMargin < 0 ? 'red' : 'green'}
                    />
                  </div>
                  <div class="mt-3 flex flex-wrap gap-2">
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
          </div>
        </Show>
      </Card>
    </PageShell>
  )
}
