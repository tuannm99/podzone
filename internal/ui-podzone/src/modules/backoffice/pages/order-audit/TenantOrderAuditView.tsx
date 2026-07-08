import { useParams } from '@tanstack/solid-router'
import { For, Show } from 'solid-js'
import { EmptyBlock, ErrorAlert, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Badge, Button, Card, InputField, SelectField } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { useTenantWorkspace } from '@/solid/workspace/context'
import { createOrderAuditViewModel } from './createOrderAuditViewModel'
import {
    activityColor,
    activityFilterOptions,
    buildOrdersHref,
    formatActivityActor,
    formatActivityTime,
    timeWindowOptions,
    type ActivityFilter,
    type TimeWindow,
} from './presentation'

export function TenantOrderAuditView() {
    const params = useParams({ from: '/t/$tenantId/orders/audit' })
    const workspace = useTenantWorkspace()

    const currentStoreId = () => workspace?.currentStoreId() || ''
    const currentStore = () => workspace?.currentStore()
    const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
    const storeLabel = () => currentStore()?.name || currentStoreId() || 'selected store'
    const {
        nextCursor,
        total,
        activityFilter,
        setActivityFilter,
        hideSystemActivity,
        setHideSystemActivity,
        timeWindow,
        setTimeWindow,
        actorFilter,
        setActorFilter,
        orderFilter,
        setOrderFilter,
        partnerFilter,
        setPartnerFilter,
        assigneeFilter,
        setAssigneeFilter,
        message,
        error,
        loading,
        loadingMore,
        applyFilters,
        auditFeed,
        copyFeed,
        loadMore,
    } = createOrderAuditViewModel({
        tenantID: () => params().tenantId,
        storeID: currentStoreId,
        storeLabel,
        workspaceReady,
    })

    return (
        <PageShell>
            <Card class="space-y-4">
                <SectionLead
                    eyebrow="Store Audit"
                    title={`Audit history for ${storeLabel()}`}
                    copy="Review store-wide routed order activity across shipment, settlement, issue handling, and queue ownership updates without opening each order card."
                />
            </Card>

            <Show when={message()}>
                <InfoAlert>{message()}</InfoAlert>
            </Show>

            <Show when={error()}>
                <ErrorAlert>{error()}</ErrorAlert>
            </Show>
            <Show when={loading()}>
                <LoadingInline label="Loading audit activity..." />
            </Show>

            <Show when={!workspaceReady()}>
                <EmptyBlock
                    title="Choose a store first"
                    copy="Use the workspace store switcher before loading the store audit feed."
                />
            </Show>

            <Card class="space-y-4">
                <SectionTitle
                    title="Audit filters"
                    subtitle="Focus the store-wide activity feed by type, actor, order, partner, assignee, and recent time window."
                />
                <div class="grid gap-4 md:grid-cols-3">
                    <SelectField
                        label="Activity type"
                        value={activityFilter()}
                        options={activityFilterOptions.map((option) => ({
                            name: option.name,
                            value: option.value,
                        }))}
                        onChange={(event) => setActivityFilter(event.currentTarget.value as ActivityFilter)}
                    />
                    <SelectField
                        label="Time window"
                        value={timeWindow()}
                        options={timeWindowOptions.map((option) => ({
                            name: option.name,
                            value: option.value,
                        }))}
                        onChange={(event) => setTimeWindow(event.currentTarget.value as TimeWindow)}
                    />
                    <InputField
                        label="Actor filter"
                        value={actorFilter()}
                        placeholder="user:12"
                        onInput={(event) => setActorFilter(event.currentTarget.value)}
                    />
                    <InputField
                        label="Order filter"
                        value={orderFilter()}
                        placeholder="ORD-1234ABCD"
                        onInput={(event) => setOrderFilter(event.currentTarget.value)}
                    />
                    <InputField
                        label="Partner filter"
                        value={partnerFilter()}
                        placeholder="Print Partner A"
                        onInput={(event) => setPartnerFilter(event.currentTarget.value)}
                    />
                    <InputField
                        label="Assignee filter"
                        value={assigneeFilter()}
                        placeholder="ops.lead"
                        onInput={(event) => setAssigneeFilter(event.currentTarget.value)}
                    />
                </div>
                <div class="flex flex-wrap items-center gap-2">
                    <Show when={activityFilter() === 'all'}>
                        <Button
                            type="button"
                            size="xs"
                            color={hideSystemActivity() ? 'dark' : 'light'}
                            onClick={() => setHideSystemActivity((current) => !current)}
                        >
                            {hideSystemActivity() ? 'Show system' : 'Hide system'}
                        </Button>
                    </Show>
                    <Button
                        type="button"
                        size="xs"
                        color="alternative"
                        loading={loading()}
                        onClick={() => applyFilters()}
                    >
                        Apply filters
                    </Button>
                    <Button
                        type="button"
                        size="xs"
                        color="light"
                        href={buildOrdersHref(params().tenantId, currentStoreId())}
                    >
                        Back to orders board
                    </Button>
                    <Button
                        type="button"
                        size="xs"
                        color="blue"
                        onClick={() => {
                            void copyFeed()
                        }}
                    >
                        Copy audit feed
                    </Button>
                </div>
            </Card>

            <Card class="space-y-4">
                <SectionTitle
                    title="Audit feed"
                    subtitle="Newest matching activity first across all routed orders in this store."
                />
                <Show
                    when={auditFeed().length > 0}
                    fallback={
                        <EmptyBlock
                            title="No audit activity matched"
                            copy="Try widening the time window, removing the actor filter, or including system activity."
                        />
                    }
                >
                    <div class="flex flex-wrap items-center justify-between gap-2 text-sm text-slate-500">
                        <p>
                            Showing {auditFeed().length} of {total()} matching activity entries.
                        </p>
                        <Show when={!!nextCursor()}>
                            <Button
                                type="button"
                                size="xs"
                                color="alternative"
                                loading={loadingMore()}
                                disabled={loadingMore()}
                                onClick={() => {
                                    void loadMore()
                                }}
                            >
                                Load more
                            </Button>
                        </Show>
                    </div>
                    <div class="space-y-3">
                        <For each={auditFeed()}>
                            {(entry) => (
                                <div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
                                    <div class="flex flex-wrap items-center justify-between gap-2">
                                        <div class="flex flex-wrap items-center gap-2">
                                            <Badge
                                                content={entry.activity.type.replaceAll('_', ' ')}
                                                color={activityColor(entry.activity.type)}
                                            />
                                            <p class="text-sm font-semibold text-slate-900">{entry.orderId}</p>
                                            <p class="text-sm text-slate-500">{entry.productTitle}</p>
                                            <p class="text-sm text-slate-500">{entry.partner}</p>
                                        </div>
                                        <p class="text-xs text-slate-500">
                                            {formatActivityTime(entry.activity.createdAt)}
                                        </p>
                                    </div>
                                    <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-slate-500">
                                        <span>{formatActivityActor(entry.activity.actor)}</span>
                                        <span>owner {entry.operatorAssignee || 'unassigned'}</span>
                                    </div>
                                    <p class="mt-3 text-sm text-slate-700">{entry.activity.message}</p>
                                    <Show when={entry.activity.details.length}>
                                        <div class="mt-2 flex flex-wrap gap-2">
                                            <For each={entry.activity.details}>
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
                    </div>
                </Show>
            </Card>
        </PageShell>
    )
}
