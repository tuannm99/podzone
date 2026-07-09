import { For, Show, createEffect, createMemo, createSignal } from 'solid-js'
import type { Accessor, Setter } from 'solid-js'
import type { PageInfo } from '@/services/collection'
import type { RoutedOrder } from '@/services/orders'
import { EmptyBlock, ErrorAlert, InfoAlert } from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Badge, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { BulkOpsPanel } from './panels/BulkOpsPanel'
import { TenantOrdersBoardProvider } from './board-context'
import type { TenantOrdersBoardContextValue } from './board-context'
import { CreateRoutedOrderPanel } from './panels/CreateRoutedOrderPanel'
import { TenantOrdersComposerProvider } from './composer-context'
import type { TenantOrdersComposerContextValue } from './composer-context'
import { TenantOrdersInsightsProvider } from './insights-context'
import type { TenantOrdersInsightsContextValue } from './insights-context'
import { OrderCard } from './OrderCard'
import { OrdersQueueTable } from './panels/OrdersQueueTable'
import { OrdersInsightsPanel } from './panels/OrdersInsightsPanel'
import { QueueToolbarPanel } from './panels/QueueToolbarPanel'
import { StoreActivityFeedPanel } from './panels/StoreActivityFeedPanel'
import {
    activityColor,
    activityFilterOptions,
    exceptionColor,
    isOverdue,
    issueResolutionOptions,
    routeStatuses,
    settlementColor,
    settlementOptions,
    shipmentColor,
    shipmentOptions,
    statusColor,
} from './shared/presentation'
import type { ActivityFilter } from './order-card/types'
import type { useOrderActions } from './hooks/useOrderActions'
import type { useOrderDrafts } from './hooks/useOrderDrafts'
import type { useOrderInsights } from './hooks/useOrderInsights'
import type { useOrderStorage } from './hooks/useOrderStorage'

type TenantOrdersViewProps = {
    storeLabel: Accessor<string>
    workspaceReady: Accessor<boolean>
    message: Accessor<string>
    error: Accessor<string>
    orders: Accessor<RoutedOrder[]>
    queueOrders: Accessor<RoutedOrder[]>
    queuePageInfo: Accessor<PageInfo>
    queuePage: Accessor<number>
    setQueuePage: (page: number) => void
    activityFilter: Accessor<ActivityFilter>
    setActivityFilter: Setter<ActivityFilter>
    hideSystemActivity: Accessor<boolean>
    setHideSystemActivity: Setter<boolean>
    composerContextValue: TenantOrdersComposerContextValue
    boardContextValue: TenantOrdersBoardContextValue
    insightsContextValue: TenantOrdersInsightsContextValue
    storage: ReturnType<typeof useOrderStorage>
    actions: ReturnType<typeof useOrderActions>
    drafts: ReturnType<typeof useOrderDrafts>
    insights: ReturnType<typeof useOrderInsights>
}

export function TenantOrdersView(props: TenantOrdersViewProps) {
    const [detailOrderID, setDetailOrderID] = createSignal('')
    const detailOrder = createMemo(() => props.queueOrders().find((order) => order.id === detailOrderID()))

    createEffect(() => {
        const orders = props.queueOrders()
        if (!orders.some((order) => order.id === detailOrderID())) {
            setDetailOrderID(orders[0]?.id || '')
        }
    })

    return (
        <PageShell>
            <Card class="space-y-4">
                <SectionLead
                    eyebrow="Order Routing Workspace"
                    title={`POD routing board for ${props.storeLabel()}`}
                    copy="This routing workspace persists store-scoped POD orders in the backend. Published mock products can be routed through production, shipment, issue handling, and settlement readiness."
                />
            </Card>

            <Show when={props.message()}>
                <InfoAlert>{props.message()}</InfoAlert>
            </Show>

            <Show when={props.error()}>
                <ErrorAlert>{props.error()}</ErrorAlert>
            </Show>

            <Show when={!props.workspaceReady()}>
                <EmptyBlock
                    title="Choose a store first"
                    copy="Use the workspace store switcher before loading store-scoped routing and fulfillment data."
                />
            </Show>

            <InfoAlert>
                Orders and published product candidates now come from backend store data. Shipment and settlement
                control both stay manual on this board so the store team can manage POD execution directly.
            </InfoAlert>

            <div class="grid gap-6 lg:grid-cols-[0.96fr_1.04fr]">
                <Card class="space-y-4">
                    <TenantOrdersComposerProvider value={props.composerContextValue}>
                        <CreateRoutedOrderPanel />
                    </TenantOrdersComposerProvider>
                </Card>

                <Card class="space-y-4">
                    <SectionTitle
                        title="Routing board"
                        subtitle="Watch each order move from intake to production, then manage shipment and settlement state directly inside the store-scoped POD workflow."
                    />

                    <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
                        <TenantOrdersBoardProvider value={props.boardContextValue}>
                            <QueueToolbarPanel />
                            <BulkOpsPanel />
                        </TenantOrdersBoardProvider>
                        <TenantOrdersInsightsProvider value={props.insightsContextValue}>
                            <OrdersInsightsPanel />
                        </TenantOrdersInsightsProvider>
                    </div>

                    <Show
                        when={props.queueOrders().length > 0}
                        fallback={
                            <EmptyBlock
                                title={
                                    props.orders().length > 0 ? 'No orders in this queue view' : 'No routed orders yet'
                                }
                                copy={
                                    props.orders().length > 0
                                        ? 'Adjust the queue view or operator lens to inspect a different operational slice.'
                                        : 'Create a routed order on the left to test store-side routing, manual shipment control, and settlement readiness.'
                                }
                            />
                        }
                    >
                        <div class="space-y-3">
                            <TenantOrdersInsightsProvider value={props.insightsContextValue}>
                                <StoreActivityFeedPanel />
                            </TenantOrdersInsightsProvider>
                            <OrdersQueueTable
                                orders={props.queueOrders}
                                pageInfo={props.queuePageInfo}
                                page={props.queuePage}
                                onPageChange={props.setQueuePage}
                                detailOrderID={detailOrderID}
                                setDetailOrderID={setDetailOrderID}
                                isSelected={props.storage.isSelected}
                                toggleSelected={props.storage.toggleOrderSelection}
                                priorityScoreFor={props.insights.priorityScoreFor}
                                statusColor={statusColor}
                                exceptionColor={exceptionColor}
                                settlementColor={settlementColor}
                            />
                            <Show when={detailOrder()}>
                                {(order) => (
                                    <div class="border-t border-gray-200 pt-4">
                                        <p class="mb-3 text-sm font-semibold text-gray-900">Order detail</p>
                                        <OrderCard
                                            order={order()}
                                            selected={props.storage.isSelected(order().id)}
                                            actions={{
                                                toggleSelected: (checked) =>
                                                    props.storage.toggleOrderSelection(order().id, checked),
                                                ...props.actions,
                                                copyActivitySummary: props.insights.copyActivitySummary,
                                                queueDraftFor: props.drafts.queueDraftFor,
                                                patchQueueDraft: props.drafts.patchQueueDraft,
                                                issueDraftFor: props.drafts.issueDraftFor,
                                                patchIssueDraft: props.drafts.patchIssueDraft,
                                                settlementDraftFor: props.drafts.settlementDraftFor,
                                                patchSettlementDraft: props.drafts.patchSettlementDraft,
                                                shipmentDraftFor: props.drafts.shipmentDraftFor,
                                                patchShipmentDraft: props.drafts.patchShipmentDraft,
                                                rerouteDraftFor: props.drafts.rerouteDraftFor,
                                                patchRerouteDraft: props.drafts.patchRerouteDraft,
                                            }}
                                            helpers={{
                                                queueSort: props.boardContextValue.activeQueueSort(),
                                                priorityScoreFor: props.insights.priorityScoreFor,
                                                statusColor,
                                                exceptionColor,
                                                shipmentColor,
                                                settlementColor,
                                                activityColor,
                                                isOverdue,
                                                filteredActivityLogFor: props.insights.filteredActivityLogFor,
                                                hiddenSystemActivityCountFor:
                                                    props.insights.hiddenSystemActivityCountFor,
                                            }}
                                            ui={{
                                                activityFilter: props.activityFilter(),
                                                setActivityFilter: props.setActivityFilter,
                                                hideSystemActivity: props.hideSystemActivity(),
                                                toggleHideSystemActivity: () =>
                                                    props.setHideSystemActivity((current) => !current),
                                                activityFilterOptions,
                                                shipmentOptions,
                                                settlementOptions,
                                                issueResolutionOptions,
                                            }}
                                        />
                                    </div>
                                )}
                            </Show>
                        </div>
                    </Show>
                </Card>
            </div>

            <Card class="mt-6 space-y-4">
                <SectionTitle
                    title="Routing, shipment, and settlement stages"
                    subtitle="Production routing, queue ownership, shipment control, and settlement updates are all managed inside the workspace, so operators can run POD execution manually without relying on external fulfillment callbacks."
                />
                <div class="flex flex-wrap gap-2">
                    <For each={routeStatuses}>
                        {(stage) => <Badge content={stage.name} color={statusColor(stage.value)} />}
                    </For>
                    <Badge content="Open issue" color="yellow" />
                    <Badge content="Escalated issue" color="red" />
                    <Badge content="Resolved issue" color="green" />
                    <For each={shipmentOptions}>
                        {(stage) => <Badge content={stage.name} color={shipmentColor(stage.value)} />}
                    </For>
                    <For each={settlementOptions}>
                        {(stage) => <Badge content={`Settlement ${stage.name}`} color={settlementColor(stage.value)} />}
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
    )
}
