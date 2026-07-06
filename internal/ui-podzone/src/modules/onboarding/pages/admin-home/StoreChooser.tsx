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
import { EmptyBlock } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button, Card, SelectField } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { CreateFirstStore } from './CreateFirstStore'
import { useAdminHome } from './context'

export function StoreChooser() {
    const vm = useAdminHome()
    const hasCollectionQuery = () => Boolean(vm.stores.query.search?.trim()) || Boolean(vm.stores.query.filters?.length)

    return (
        <Card class="space-y-4">
            <SectionTitle
                title="Choose store"
                subtitle="Pick a workspace, inspect its stores, then enter one store-scoped workspace."
            />
            <div class="grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
                <SelectField
                    label="Workspace"
                    value={vm.selectedWorkspaceId()}
                    options={vm.selectedWorkspaceOptions()}
                    onChange={(event) => vm.setSelectedWorkspaceId(event.currentTarget.value)}
                />
                <div class="flex flex-wrap items-end justify-between gap-3">
                    <p class="text-sm text-gray-600">{vm.currentSelectionLabel()}</p>
                    <Button
                        disabled={!vm.selectedWorkspaceId() || !vm.selectedStoreId()}
                        loading={vm.switchingTenant()}
                        onClick={() => void vm.openStore(vm.selectedWorkspaceId(), vm.selectedStoreId())}
                    >
                        Open selected store
                    </Button>
                </div>
            </div>

            <Show when={vm.selectedWorkspace()}>
                <CollectionControls
                    query={vm.stores.query}
                    loading={vm.stores.loading}
                    error={vm.storesError}
                    searchPlaceholder="Search store name, ID, or status"
                    sortOptions={[
                        { label: 'Created', value: 'createdAt' },
                        { label: 'Updated', value: 'updatedAt' },
                        { label: 'Name', value: 'name' },
                        { label: 'Status', value: 'status' },
                    ]}
                    filterFields={[
                        {
                            label: 'Name',
                            value: 'name',
                            operators: [
                                'FILTER_OPERATOR_EQ',
                                'FILTER_OPERATOR_CONTAINS',
                                'FILTER_OPERATOR_STARTS_WITH',
                            ],
                        },
                        {
                            label: 'Status',
                            value: 'status',
                            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                        },
                    ]}
                    updateQuery={vm.stores.updateQuery}
                />
                <Show
                    when={vm.stores.items().length > 0}
                    fallback={
                        <Show when={!vm.stores.loading()}>
                            <Show
                                when={!hasCollectionQuery()}
                                fallback={
                                    <EmptyBlock
                                        title="No stores match"
                                        copy="Adjust search or filters to inspect another store."
                                    />
                                }
                            >
                                <CreateFirstStore />
                            </Show>
                        </Show>
                    }
                >
                    <DataTable>
                        <TableHead>
                            <TableRow>
                                <TableHeaderCell class="w-12">Select</TableHeaderCell>
                                <TableHeaderCell>Store</TableHeaderCell>
                                <TableHeaderCell>Status</TableHeaderCell>
                                <TableHeaderCell>Updated</TableHeaderCell>
                                <TableHeaderCell class="text-right">Action</TableHeaderCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            <For each={vm.stores.items()}>
                                {(store) => (
                                    <TableRow>
                                        <TableCell>
                                            <input
                                                type="radio"
                                                name="selected-store"
                                                aria-label={`Select ${store.name}`}
                                                checked={vm.selectedStoreId() === store.id}
                                                onChange={() => vm.setSelectedStoreId(store.id)}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <p class="font-medium text-gray-900">{store.name}</p>
                                            <p class="text-xs text-gray-500">{store.id}</p>
                                        </TableCell>
                                        <TableCell>
                                            <Badge
                                                content={store.status || 'unknown'}
                                                color={store.isActive ? 'green' : 'dark'}
                                            />
                                        </TableCell>
                                        <TableCell>{store.updatedAt || 'Unknown'}</TableCell>
                                        <TableCell class="text-right">
                                            <Button
                                                size="xs"
                                                onClick={() => void vm.openStore(vm.selectedWorkspaceId(), store.id)}
                                            >
                                                Open
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                )}
                            </For>
                        </TableBody>
                    </DataTable>
                    <Pagination
                        page={vm.stores.pageInfo().page}
                        pageSize={vm.stores.pageInfo().pageSize}
                        total={vm.stores.pageInfo().total}
                        loading={vm.stores.loading()}
                        onPageChange={(page) => vm.stores.updateQuery({ page })}
                    />
                </Show>
            </Show>
            <Show when={!vm.selectedWorkspace()}>
                <EmptyBlock
                    title="No workspace selected"
                    copy="Choose or create a workspace before selecting a store."
                />
            </Show>
        </Card>
    )
}
