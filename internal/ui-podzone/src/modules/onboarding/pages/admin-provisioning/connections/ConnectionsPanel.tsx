import { For, Show } from 'solid-js'
import { CollectionControls } from '@/solid/components/common/CollectionControls'
import {
    DataTable,
    TableBody,
    TableCell,
    TableHead,
    TableHeaderCell,
    TableRow,
} from '@/solid/components/common/DataTable'
import { EmptyBlock, ErrorAlert } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { useAdminProvisioning } from '../context'
import { ConnectionEditor } from './ConnectionEditor'

function formatTime(value: string) {
    return value ? new Date(value).toLocaleString() : 'Never'
}

export function ConnectionsPanel() {
    const { shell, connections } = useAdminProvisioning()

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
