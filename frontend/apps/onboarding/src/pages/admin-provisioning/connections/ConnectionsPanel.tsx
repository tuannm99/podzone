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
import { EmptyBlock, ErrorAlert } from '@podzone/shared/ui/components/common/Feedback'
import { Pagination } from '@podzone/shared/ui/components/common/Pagination'
import { Badge, Button } from '@podzone/shared/ui/components/common/Primitives'
import { useAdminProvisioning } from '../context'
import { ConnectionEditor } from './ConnectionEditor'

function formatTime(value: string) {
    return value ? new Date(value).toLocaleString() : 'Never'
}

export function ConnectionsPanel() {
    const { shell, connections } = useAdminProvisioning()
    const routeValue = (value?: string) => value || 'Not published'

    return (
        <section class="space-y-5 rounded-lg border border-gray-200 bg-white p-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <h2 class="text-base font-semibold text-gray-950">Tenant connection registry</h2>
                    <p class="mt-1 text-sm text-gray-500">
                        Runtime routes and secret references published for{' '}
                        {shell.selectedTenantId() || 'the selected workspace'}.
                    </p>
                </div>
                <Button size="sm" disabled={!shell.workspaceReady()} onClick={connections.openCreate}>
                    Add connection
                </Button>
            </div>
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
                <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                    <div>
                        <div class="flex flex-wrap items-center gap-2">
                            <h3 class="text-sm font-semibold text-gray-950">Placement route health</h3>
                            <Show when={connections.placementStatus()}>
                                {(status) => (
                                    <>
                                        <Badge
                                            content={
                                                status().allocation_ready ? 'allocation ready' : 'allocation missing'
                                            }
                                            color={status().allocation_ready ? 'green' : 'red'}
                                        />
                                        <Badge
                                            content={status().route_ready ? 'route published' : 'route missing'}
                                            color={status().route_ready ? 'green' : 'yellow'}
                                        />
                                        <Badge
                                            content={status().in_sync ? 'in sync' : 'repair needed'}
                                            color={status().in_sync ? 'green' : 'red'}
                                        />
                                    </>
                                )}
                            </Show>
                        </div>
                        <p class="mt-1 text-sm text-gray-500">
                            Compares onboarding allocation source of truth with the runtime KV route consumed by tenant
                            database resolution.
                        </p>
                    </div>
                    <div class="flex gap-2">
                        <Button
                            size="xs"
                            color="alternative"
                            loading={connections.placementLoading()}
                            disabled={!shell.workspaceReady()}
                            onClick={() => void connections.loadPlacementStatus()}
                        >
                            Check route
                        </Button>
                        <Button
                            size="xs"
                            disabled={!shell.workspaceReady()}
                            loading={connections.placementRepairing()}
                            onClick={() => void connections.reconcile()}
                        >
                            Republish route
                        </Button>
                    </div>
                </div>
                <Show when={connections.placementError()}>
                    <div class="mt-3">
                        <ErrorAlert>{connections.placementError()}</ErrorAlert>
                    </div>
                </Show>
                <Show when={connections.placementStatus()}>
                    {(status) => (
                        <div class="mt-4 grid gap-3 md:grid-cols-2">
                            <div class="rounded border border-gray-200 bg-white p-3">
                                <p class="text-xs font-semibold uppercase text-gray-500">Allocation</p>
                                <p class="mt-2 text-sm text-gray-700">
                                    {routeValue(status().allocation?.cluster_name)} ·{' '}
                                    {routeValue(status().allocation?.db_name)}
                                </p>
                                <p class="mt-1 text-xs text-gray-500">
                                    mode {routeValue(status().allocation?.mode)} · schema{' '}
                                    {routeValue(status().allocation?.schema_name)}
                                </p>
                            </div>
                            <div class="rounded border border-gray-200 bg-white p-3">
                                <p class="text-xs font-semibold uppercase text-gray-500">Runtime KV route</p>
                                <p class="mt-2 text-sm text-gray-700">
                                    {routeValue(status().route?.cluster_name)} · {routeValue(status().route?.db_name)}
                                </p>
                                <p class="mt-1 text-xs text-gray-500">
                                    mode {routeValue(status().route?.mode)} · schema{' '}
                                    {routeValue(status().route?.schema_name)}
                                </p>
                            </div>
                            <Show when={status().reason}>
                                <p class="md:col-span-2 text-sm text-amber-700">{status().reason}</p>
                            </Show>
                        </div>
                    )}
                </Show>
            </div>
            <CollectionControls
                query={connections.connections.query}
                loading={connections.connections.loading}
                error={connections.connections.error}
                searchPlaceholder="Name, type, endpoint, status"
                sortOptions={[
                    { label: 'Updated', value: 'updatedAt' },
                    { label: 'Name', value: 'name' },
                    { label: 'Type', value: 'infraType' },
                    { label: 'Status', value: 'status' },
                ]}
                filterFields={[
                    {
                        label: 'Type',
                        value: 'infraType',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
                    {
                        label: 'Status',
                        value: 'status',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
                ]}
                updateQuery={connections.connections.updateQuery}
            />
            <Show when={connections.error()}>
                <ErrorAlert>{connections.error()}</ErrorAlert>
            </Show>
            <Show
                when={connections.connections.items().length}
                fallback={
                    <EmptyBlock
                        title="No tenant connections"
                        copy="Provisioning will publish database and runtime routes here."
                    />
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>Connection</TableHeaderCell>
                            <TableHeaderCell>Endpoint</TableHeaderCell>
                            <TableHeaderCell>Secret</TableHeaderCell>
                            <TableHeaderCell>Updated</TableHeaderCell>
                            <TableHeaderCell />
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={connections.connections.items()}>
                            {(connection) => (
                                <TableRow>
                                    <TableCell>
                                        <p class="font-medium text-gray-950">{connection.name}</p>
                                        <div class="mt-1 flex gap-2">
                                            <Badge content={connection.infra_type} color="blue" />
                                            <Badge
                                                content={connection.status}
                                                color={connection.status === 'active' ? 'green' : 'yellow'}
                                            />
                                        </div>
                                    </TableCell>
                                    <TableCell class="max-w-72 break-all text-gray-600">
                                        {connection.endpoint}
                                    </TableCell>
                                    <TableCell class="max-w-56 break-all text-gray-600">
                                        {connection.secret_ref || 'Not configured'}
                                    </TableCell>
                                    <TableCell class="whitespace-nowrap text-gray-500">
                                        {formatTime(connection.updated_at)}
                                    </TableCell>
                                    <TableCell class="text-right">
                                        <div class="flex justify-end gap-2">
                                            <Button
                                                size="xs"
                                                color="alternative"
                                                onClick={() => connections.openEdit(connection)}
                                            >
                                                Edit
                                            </Button>
                                            <Button
                                                size="xs"
                                                color="red"
                                                loading={connections.deletingName() === connection.name}
                                                onClick={() => void connections.remove(connection)}
                                            >
                                                Delete
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={connections.connections.pageInfo().page}
                    pageSize={connections.connections.pageInfo().pageSize}
                    total={connections.connections.pageInfo().total}
                    loading={connections.connections.loading()}
                    onPageChange={(page) => connections.connections.updateQuery({ page })}
                />
            </Show>
            <ConnectionEditor
                open={connections.creating() || Boolean(connections.editor())}
                connection={connections.editor()}
                vm={connections}
            />
        </section>
    )
}
