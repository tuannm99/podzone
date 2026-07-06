import { For, Show } from 'solid-js'
import { CollectionFilters } from '@/solid/components/common/CollectionFilters'
import { CollectionToolbar } from '@/solid/components/common/CollectionToolbar'
import {
    DataTable,
    TableBody,
    TableCell,
    TableHead,
    TableHeaderCell,
    TableRow,
} from '@/solid/components/common/DataTable'
import { EmptyBlock, ErrorAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button, SelectField } from '@/solid/components/common/Primitives'
import { useAdminIamGroup } from './context'

export function GroupsCollection() {
    const group = useAdminIamGroup()

    return (
        <div class="space-y-4">
            <div class="grid gap-3 md:grid-cols-2">
                <SelectField
                    label="Collection scope"
                    value={group.groupScope()}
                    options={group.groupScopeOptions()}
                    onChange={(event) => {
                        group.setGroupScope(event.currentTarget.value)
                        group.updateQuery({ page: 1 })
                    }}
                />
                <Show when={group.groupScope() === 'tenant'}>
                    <SelectField
                        label="Collection tenant"
                        value={group.groupTenantId()}
                        options={group.tenantOptions()}
                        onChange={(event) => {
                            group.setGroupTenantId(event.currentTarget.value)
                            group.updateQuery({ page: 1 })
                        }}
                    />
                </Show>
            </div>
            <CollectionToolbar
                search={group.query.search || ''}
                searchPlaceholder="Search group name, scope, tenant, or description"
                sortBy={group.query.sortBy || 'createdAt'}
                sortDirection={group.query.sortDirection || 'SORT_DIRECTION_DESC'}
                pageSize={group.query.pageSize}
                sortOptions={[
                    { label: 'Created', value: 'createdAt' },
                    { label: 'Updated', value: 'updatedAt' },
                    { label: 'Name', value: 'name' },
                    { label: 'Scope', value: 'scope' },
                ]}
                onSearch={(search) => group.updateQuery({ search })}
                onSortByChange={(sortBy) => group.updateQuery({ sortBy })}
                onSortDirectionChange={(sortDirection) => group.updateQuery({ sortDirection })}
                onPageSizeChange={(pageSize) => group.updateQuery({ pageSize })}
            />
            <CollectionFilters
                fields={[
                    {
                        label: 'Name',
                        value: 'name',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS', 'FILTER_OPERATOR_STARTS_WITH'],
                    },
                    {
                        label: 'Tenant ID',
                        value: 'tenantId',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
                    {
                        label: 'System group',
                        value: 'isSystem',
                        operators: ['FILTER_OPERATOR_EQ'],
                    },
                ]}
                filters={group.query.filters || []}
                onChange={(filters) => group.updateQuery({ filters })}
            />
            <Show when={group.error()}>
                <ErrorAlert>{group.error()}</ErrorAlert>
            </Show>
            <Show when={group.loading()}>
                <LoadingInline label="Loading groups..." />
            </Show>
            <Show
                when={group.groups().length > 0}
                fallback={
                    <EmptyBlock title="No groups" copy="No groups match the selected scope and collection query." />
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>Group</TableHeaderCell>
                            <TableHeaderCell>Scope</TableHeaderCell>
                            <TableHeaderCell>Tenant</TableHeaderCell>
                            <TableHeaderCell class="text-right">Action</TableHeaderCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={group.groups()}>
                            {(item) => (
                                <TableRow>
                                    <TableCell>
                                        <p class="font-semibold text-gray-900">{item.name}</p>
                                        <p class="mt-1 text-xs text-gray-500">{item.description || 'No description'}</p>
                                    </TableCell>
                                    <TableCell>
                                        <div class="flex flex-wrap gap-2">
                                            <Badge content={item.scope} color="blue" />
                                            <Show when={item.isSystem}>
                                                <Badge content="system" color="dark" />
                                            </Show>
                                        </div>
                                    </TableCell>
                                    <TableCell>{item.tenantId || 'Platform'}</TableCell>
                                    <TableCell class="text-right">
                                        <Button
                                            type="button"
                                            size="xs"
                                            color={String(item.id) === group.selectedGroupId() ? 'dark' : 'alternative'}
                                            onClick={() => group.setSelectedGroupId(String(item.id || ''))}
                                        >
                                            Inspect
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={group.pageInfo().page}
                    pageSize={group.pageInfo().pageSize}
                    total={group.pageInfo().total}
                    loading={group.loading()}
                    onPageChange={(page) => group.updateQuery({ page })}
                />
            </Show>
        </div>
    )
}
