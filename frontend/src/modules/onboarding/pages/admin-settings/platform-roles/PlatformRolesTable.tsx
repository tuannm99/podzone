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
import { Badge, Button } from '@/solid/components/common/Primitives'
import { useAdminSettings } from '../context'
import { membershipStatusColor } from '../presentation'

export function PlatformRolesTable() {
    const { platformRoles } = useAdminSettings()

    return (
        <section class="space-y-3">
            <CollectionControls
                query={platformRoles.query}
                loading={platformRoles.loading}
                error={platformRoles.collectionError}
                searchPlaceholder="Search role or status"
                sortOptions={[
                    { label: 'Assigned', value: 'createdAt' },
                    { label: 'Role', value: 'roleName' },
                    { label: 'Status', value: 'status' },
                ]}
                filterFields={[
                    {
                        label: 'Role',
                        value: 'roleName',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
                    {
                        label: 'Status',
                        value: 'status',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
                ]}
                updateQuery={platformRoles.updateQuery}
            />
            <Show
                when={platformRoles.items().length > 0}
                fallback={
                    <Show when={!platformRoles.loading()}>
                        <EmptyBlock
                            title="No admin roles"
                            copy="No platform roles match the current user and collection query."
                        />
                    </Show>
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>User</TableHeaderCell>
                            <TableHeaderCell>Role</TableHeaderCell>
                            <TableHeaderCell>Status</TableHeaderCell>
                            <TableHeaderCell class="text-right">Action</TableHeaderCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={platformRoles.items()}>
                            {(membership) => (
                                <TableRow>
                                    <TableCell class="font-medium text-gray-900">{membership.userId}</TableCell>
                                    <TableCell>{membership.roleName}</TableCell>
                                    <TableCell>
                                        <Badge
                                            content={membership.status}
                                            color={membershipStatusColor(membership.status)}
                                        />
                                    </TableCell>
                                    <TableCell class="text-right">
                                        <Button
                                            color="red"
                                            size="xs"
                                            disabled={!platformRoles.canManage()}
                                            onClick={() =>
                                                void platformRoles.remove(membership.userId, membership.roleName)
                                            }
                                        >
                                            Remove
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={platformRoles.pageInfo().page}
                    pageSize={platformRoles.pageInfo().pageSize}
                    total={platformRoles.pageInfo().total}
                    loading={platformRoles.loading()}
                    onPageChange={(page) => platformRoles.updateQuery({ page })}
                />
            </Show>
        </section>
    )
}
