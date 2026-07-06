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

export function InvitesTable() {
    const { invites } = useAdminSettings()
    const { access } = invites

    return (
        <section class="space-y-3">
            <CollectionControls
                query={access.invites.query}
                loading={access.invites.loading}
                error={access.invites.error}
                searchPlaceholder="Search email, role, or status"
                sortOptions={[
                    { label: 'Created', value: 'createdAt' },
                    { label: 'Expires', value: 'expiresAt' },
                    { label: 'Email', value: 'email' },
                    { label: 'Role', value: 'roleName' },
                    { label: 'Status', value: 'status' },
                ]}
                filterFields={[
                    {
                        label: 'Email',
                        value: 'email',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS', 'FILTER_OPERATOR_STARTS_WITH'],
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
                updateQuery={access.invites.updateQuery}
            />
            <Show
                when={access.invites.items().length > 0}
                fallback={
                    <Show when={!access.invites.loading()}>
                        <EmptyBlock
                            title="No workspace invites"
                            copy="No invites match the current collection query."
                        />
                    </Show>
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>Invite</TableHeaderCell>
                            <TableHeaderCell>Role</TableHeaderCell>
                            <TableHeaderCell>Status</TableHeaderCell>
                            <TableHeaderCell>Expires</TableHeaderCell>
                            <TableHeaderCell class="text-right">Action</TableHeaderCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={access.invites.items()}>
                            {(invite) => (
                                <TableRow>
                                    <TableCell>
                                        <p class="font-medium text-gray-900">{invite.email}</p>
                                        <p class="text-xs text-gray-500">{invite.tenantId}</p>
                                    </TableCell>
                                    <TableCell>{invite.roleName}</TableCell>
                                    <TableCell>
                                        <Badge content={invite.status} color={membershipStatusColor(invite.status)} />
                                    </TableCell>
                                    <TableCell>{invite.expiresAt || 'Unknown'}</TableCell>
                                    <TableCell class="text-right">
                                        <Button
                                            color="red"
                                            size="xs"
                                            disabled={!access.canManage() || invite.status !== 'pending'}
                                            onClick={() =>
                                                void invites.revoke(invite.id, invite.tenantId, invite.email)
                                            }
                                        >
                                            Revoke
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={access.invites.pageInfo().page}
                    pageSize={access.invites.pageInfo().pageSize}
                    total={access.invites.pageInfo().total}
                    loading={access.invites.loading()}
                    onPageChange={(page) => access.invites.updateQuery({ page })}
                />
            </Show>
        </section>
    )
}
