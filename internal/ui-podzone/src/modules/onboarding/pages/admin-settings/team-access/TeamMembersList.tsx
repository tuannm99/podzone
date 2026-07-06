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
import { EmptyBlock, InfoAlert } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { useAdminSettings } from '../context'
import { membershipStatusColor } from '../presentation'

export function TeamMembersList() {
    const { teamAccess } = useAdminSettings()
    const { access } = teamAccess

    return (
        <section class="space-y-3">
            <CollectionControls
                query={access.members.query}
                loading={access.members.loading}
                error={access.members.error}
                searchPlaceholder="Search user, role, or status"
                sortOptions={[
                    { label: 'Added', value: 'createdAt' },
                    { label: 'User ID', value: 'userId' },
                    { label: 'Role', value: 'roleName' },
                    { label: 'Status', value: 'status' },
                ]}
                filterFields={[
                    {
                        label: 'User ID',
                        value: 'userId',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
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
                updateQuery={access.members.updateQuery}
            />
            <Show when={!access.loadingAccess() && !access.canRead()}>
                <EmptyBlock
                    title="No workspace access"
                    copy="You do not currently have permission to inspect team access for this workspace."
                />
            </Show>
            <Show when={access.canRead() && !access.loadingAccess() && !access.canManage()}>
                <InfoAlert>
                    You can inspect this workspace, but only authorized workspace owners or admins can manage team
                    access.
                </InfoAlert>
            </Show>
            <Show
                when={access.canRead() && access.members.items().length > 0}
                fallback={
                    <Show when={access.canRead() && !access.members.loading()}>
                        <EmptyBlock
                            title="No team members"
                            copy="No workspace members match the current collection query."
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
                        <For each={access.members.items()}>
                            {(membership) => (
                                <TableRow>
                                    <TableCell>
                                        <p class="font-medium text-gray-900">{membership.userId}</p>
                                        <p class="text-xs text-gray-500">{membership.tenantId}</p>
                                    </TableCell>
                                    <TableCell>
                                        <Badge content={membership.roleName} color="blue" />
                                    </TableCell>
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
                                            disabled={!access.canManage()}
                                            onClick={() =>
                                                void teamAccess.remove(membership.tenantId, membership.userId)
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
                    page={access.members.pageInfo().page}
                    pageSize={access.members.pageInfo().pageSize}
                    total={access.members.pageInfo().total}
                    loading={access.members.loading()}
                    onPageChange={(page) => access.members.updateQuery({ page })}
                />
            </Show>
        </section>
    )
}
