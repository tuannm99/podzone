import { For, Show } from 'solid-js'
import { CollectionControls } from '@podzone/shared/ui/components/common/CollectionControls'
import {
    DataTable,
    TableBody,
    TableCell,
    TableHead,
    TableHeaderCell,
    TableRow,
} from '@podzone/shared/ui/components/common/DataTable'
import { EmptyBlock, ErrorAlert, LoadingInline } from '@podzone/shared/ui/components/common/Feedback'
import { Pagination } from '@podzone/shared/ui/components/common/Pagination'
import { Badge, Button } from '@podzone/shared/ui/components/common/Primitives'
import { classes } from '@podzone/shared/ui/shared/utils'
import { useAdminProvisioning } from '../context'
import { PipelineStages } from './PipelineStages'

function timestamp(value: string) {
    return value ? new Date(value).toLocaleString() : 'Pending'
}

function statusColor(status: string) {
    if (status === 'ready') return 'green' as const
    if (status.startsWith('failed') || status === 'rejected') return 'red' as const
    return 'yellow' as const
}

export function PipelinePanel() {
    const { shell, pipeline } = useAdminProvisioning()

    return (
        <section class="grid min-h-[620px] overflow-hidden rounded-lg border border-gray-200 bg-white lg:grid-cols-[minmax(380px,0.8fr)_minmax(0,1.2fr)]">
            <div class="space-y-4 border-b border-gray-200 p-5 lg:border-b-0 lg:border-r">
                <div>
                    <h2 class="text-base font-semibold text-gray-950">Provisioning runs</h2>
                    <p class="mt-1 text-sm text-gray-500">
                        Store requests for {shell.selectedTenantId() || 'the workspace'}.
                    </p>
                </div>
                <CollectionControls
                    query={pipeline.requests.query}
                    loading={pipeline.requests.loading}
                    error={pipeline.requests.error}
                    searchPlaceholder="Request, store, subdomain, status"
                    sortOptions={[
                        { label: 'Updated', value: 'updatedAt' },
                        { label: 'Created', value: 'createdAt' },
                        { label: 'Status', value: 'status' },
                    ]}
                    filterFields={[
                        {
                            label: 'Status',
                            value: 'status',
                            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                        },
                        {
                            label: 'Owner',
                            value: 'requestedBy',
                            operators: ['FILTER_OPERATOR_EQ'],
                        },
                    ]}
                    updateQuery={pipeline.requests.updateQuery}
                />
                <Show
                    when={pipeline.requests.items().length > 0}
                    fallback={
                        <EmptyBlock
                            title="No provisioning runs"
                            copy="No store request matches this workspace and query."
                        />
                    }
                >
                    <div class="max-h-[430px] overflow-auto">
                        <DataTable>
                            <TableHead>
                                <TableRow>
                                    <TableHeaderCell>Store</TableHeaderCell>
                                    <TableHeaderCell>Status</TableHeaderCell>
                                    <TableHeaderCell />
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                <For each={pipeline.requests.items()}>
                                    {(request) => (
                                        <TableRow
                                            class={classes(request.id === pipeline.selectedRequestId() && 'bg-gray-50')}
                                        >
                                            <TableCell>
                                                <p class="font-medium text-gray-950">{request.name}</p>
                                                <p class="text-xs text-gray-500">{request.subdomain}</p>
                                            </TableCell>
                                            <TableCell>
                                                <Badge
                                                    content={request.status.replaceAll('_', ' ')}
                                                    color={statusColor(request.status)}
                                                />
                                            </TableCell>
                                            <TableCell class="text-right">
                                                <Button
                                                    size="xs"
                                                    color="alternative"
                                                    onClick={() => pipeline.setSelectedRequestId(request.id)}
                                                >
                                                    Inspect
                                                </Button>
                                            </TableCell>
                                        </TableRow>
                                    )}
                                </For>
                            </TableBody>
                        </DataTable>
                    </div>
                    <Pagination
                        page={pipeline.requests.pageInfo().page}
                        pageSize={pipeline.requests.pageInfo().pageSize}
                        total={pipeline.requests.pageInfo().total}
                        loading={pipeline.requests.loading()}
                        onPageChange={(page) => pipeline.requests.updateQuery({ page })}
                    />
                </Show>
            </div>

            <div class="min-w-0 space-y-6 p-5">
                <Show
                    when={pipeline.selectedRequest()}
                    fallback={
                        <EmptyBlock
                            title="Select a provisioning run"
                            copy="Choose a request to inspect stages and transition history."
                        />
                    }
                >
                    {(request) => (
                        <>
                            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                                <div>
                                    <p class="text-xs font-semibold uppercase text-gray-500">{request().id}</p>
                                    <h2 class="mt-1 text-xl font-semibold text-gray-950">{request().name}</h2>
                                    <p class="mt-1 text-sm text-gray-500">
                                        Owner {request().owner_id || request().requested_by} · Updated{' '}
                                        {timestamp(request().updated_at)}
                                    </p>
                                </div>
                                <Button
                                    size="sm"
                                    color="alternative"
                                    loading={pipeline.retryingId() === request().id}
                                    disabled={!['failed', 'failed_retryable'].includes(request().status)}
                                    onClick={() => void pipeline.retry(request().id)}
                                >
                                    Retry
                                </Button>
                            </div>

                            <PipelineStages request={request()} transitions={pipeline.transitions.items()} />

                            <Show when={pipeline.error()}>
                                <ErrorAlert>{pipeline.error()}</ErrorAlert>
                            </Show>
                            <Show when={pipeline.transitions.loading()}>
                                <LoadingInline label="Loading pipeline history..." />
                            </Show>

                            <div class="space-y-3">
                                <h3 class="text-sm font-semibold text-gray-950">Stage history</h3>
                                <div class="max-h-[330px] overflow-auto border-t border-gray-200">
                                    <For
                                        each={pipeline.transitions.items()}
                                        fallback={
                                            <p class="py-6 text-sm text-gray-500">No transition history recorded.</p>
                                        }
                                    >
                                        {(transition) => (
                                            <div class="grid gap-2 border-b border-gray-100 py-3 text-sm sm:grid-cols-[150px_150px_1fr]">
                                                <span class="text-xs text-gray-500">
                                                    {timestamp(transition.created_at)}
                                                </span>
                                                <span class="font-medium text-gray-900">
                                                    {transition.step ||
                                                        `${transition.from || 'new'} → ${transition.to}`}
                                                </span>
                                                <span class={transition.error_code ? 'text-red-700' : 'text-gray-600'}>
                                                    {transition.reason || transition.error_code}
                                                </span>
                                            </div>
                                        )}
                                    </For>
                                </div>
                            </div>
                        </>
                    )}
                </Show>
            </div>
        </section>
    )
}
