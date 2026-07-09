import { useParams } from '@tanstack/solid-router'
import { For, Show } from 'solid-js'
import { EmptyBlock, ErrorAlert, InfoAlert } from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { StatCard } from '@/solid/components/dashboard/StatCard'
import { useTenantWorkspace } from '@/solid/context/workspace-context'
import { createOrderFinanceViewModel } from './createOrderFinanceViewModel'
import {
    anomalyFlagsFor,
    buildFinanceQueueHref,
    formatFlag,
    formatMoney,
    parseMoneyValue,
    settlementColor,
} from './presentation'

export function TenantOrderFinanceView() {
    const params = useParams({ from: '/t/$tenantId/orders/finance' })
    const workspace = useTenantWorkspace()

    const currentStoreId = () => workspace?.currentStoreId() || ''
    const currentStore = () => workspace?.currentStore()
    const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
    const storeLabel = () => currentStore()?.name || currentStoreId() || 'selected store'
    const {
        orders,
        message,
        error,
        financeOrders,
        partnerFinanceSummary,
        totalRealizedMargin,
        issueExposure,
        copySummary,
    } = createOrderFinanceViewModel({
        tenantID: () => params().tenantId,
        storeID: currentStoreId,
        storeLabel,
        workspaceReady,
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
                    value={String(orders().filter((order) => order.settlementStatus === 'disputed').length)}
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
                                            <p class="text-sm font-semibold text-slate-900">{order.id}</p>
                                            <p class="text-xs text-slate-500">
                                                {order.productTitle} · {order.partner || 'partner pending'}
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
                                                    (parseMoneyValue(order.realizedMargin) || 0) < 0 ? 'red' : 'green'
                                                }
                                            />
                                        </div>
                                    </div>
                                    <div class="mt-3 flex flex-wrap gap-2">
                                        <For each={anomalyFlagsFor(order)}>
                                            {(flag) => <Badge content={formatFlag(flag)} color="red" />}
                                        </For>
                                    </div>
                                    <Show when={order.settlementNotes}>
                                        <p class="mt-3 text-sm text-slate-600">{order.settlementNotes}</p>
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
                                        <p class="text-sm font-semibold text-slate-900">{item.partner}</p>
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
                                            <Badge content={`${item.forcedReroutes} forced reroutes`} color="indigo" />
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
