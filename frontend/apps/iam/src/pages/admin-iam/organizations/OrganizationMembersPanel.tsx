import { For, Show, type Accessor } from 'solid-js'
import type { OrganizationMembership } from '@podzone/shared/services/iam'
import type { CollectionQuery, PageInfo } from '@podzone/shared/services/collection'
import { CollectionFilters } from '@podzone/shared/ui/components/common/CollectionFilters'
import { CollectionToolbar } from '@podzone/shared/ui/components/common/CollectionToolbar'
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
import { SectionTitle } from '@podzone/shared/ui/components/common/SectionTitle'
import { SearchSelectField, type SearchSelectOption } from '@podzone/shared/ui/components/common/SearchSelectField'
import { createFormStore, FormSelectField, required } from '@podzone/shared/ui/forms'

type MemberForm = {
    userId: string
    roleName: string
}

type OrganizationMembersPanelProps = {
    identityForUser: (userID: number | string) => {
        label: string
        description: string
    }
    organizationId: Accessor<string>
    members: Accessor<OrganizationMembership[]>
    query: CollectionQuery
    pageInfo: Accessor<PageInfo>
    loading: Accessor<boolean>
    error: Accessor<string>
    updateQuery: (patch: Partial<CollectionQuery>) => void
    addMember: (userID: number, roleName: string) => Promise<void>
    removeMember: (userID: string) => Promise<void>
    userOptions: Accessor<SearchSelectOption[]>
    usersLoading: Accessor<boolean>
    usersError: Accessor<string>
    searchUsers: (search: string) => void
}

export function OrganizationMembersPanel(props: OrganizationMembersPanelProps) {
    const form = createFormStore<MemberForm>({
        initialValues: {
            userId: '',
            roleName: 'organization_viewer',
        },
        validators: {
            userId: [required('Choose a user.')],
            roleName: [required('Select an organization role.')],
        },
    })

    const submit = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!form.validate()) return
        form.setSubmitting(true)
        try {
            await props.addMember(Number.parseInt(form.values.userId, 10), form.values.roleName)
        } finally {
            form.setSubmitting(false)
        }
    }

    return (
        <section class="space-y-4 border-t border-gray-200 pt-5">
            <SectionTitle
                title="Organization members"
                subtitle="Delegate organization administration without sharing root access."
            />
            <Show when={props.organizationId()}>
                <form class="grid gap-3 md:grid-cols-[1fr_1fr_auto]" onSubmit={submit}>
                    <SearchSelectField
                        label="User"
                        value={form.values.userId}
                        options={props.userOptions()}
                        loading={props.usersLoading()}
                        error={form.error('userId') || props.usersError()}
                        onSearch={props.searchUsers}
                        onChange={(value) => form.setValue('userId', value)}
                        placeholder="Search name, username, or email"
                    />
                    <FormSelectField
                        form={form}
                        name="roleName"
                        label="Organization role"
                        options={[
                            { name: 'Organization admin', value: 'organization_admin' },
                            { name: 'Organization viewer', value: 'organization_viewer' },
                        ]}
                    />
                    <div class="self-end">
                        <Button type="submit" size="sm" disabled={form.isSubmitting()}>
                            Add member
                        </Button>
                    </div>
                </form>
                <CollectionToolbar
                    search={props.query.search || ''}
                    searchPlaceholder="Search user, role, or status"
                    sortBy={props.query.sortBy || 'createdAt'}
                    sortDirection={props.query.sortDirection || 'SORT_DIRECTION_DESC'}
                    pageSize={props.query.pageSize}
                    sortOptions={[
                        { label: 'Created', value: 'createdAt' },
                        { label: 'Updated', value: 'updatedAt' },
                        { label: 'User ID', value: 'userId' },
                        { label: 'Role', value: 'roleName' },
                        { label: 'Status', value: 'status' },
                    ]}
                    onSearch={(search) => props.updateQuery({ search })}
                    onSortByChange={(sortBy) => props.updateQuery({ sortBy })}
                    onSortDirectionChange={(sortDirection) => props.updateQuery({ sortDirection })}
                    onPageSizeChange={(pageSize) => props.updateQuery({ pageSize })}
                />
                <CollectionFilters
                    fields={[
                        {
                            label: 'User ID',
                            value: 'userId',
                            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                        },
                        {
                            label: 'Role',
                            value: 'roleName',
                            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
                        },
                        {
                            label: 'Status',
                            value: 'status',
                            operators: ['FILTER_OPERATOR_EQ'],
                        },
                    ]}
                    filters={props.query.filters || []}
                    onChange={(filters) => props.updateQuery({ filters })}
                />
            </Show>

            <Show when={props.error()}>
                <ErrorAlert>{props.error()}</ErrorAlert>
            </Show>
            <Show when={props.loading()}>
                <LoadingInline label="Loading organization members..." />
            </Show>
            <Show
                when={props.members().length > 0}
                fallback={
                    <EmptyBlock
                        title="No organization members"
                        copy="Add an administrator or viewer to this organization."
                    />
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
                        <For each={props.members()}>
                            {(member) => {
                                const identity = () => props.identityForUser(member.userId)
                                return (
                                    <TableRow>
                                        <TableCell>
                                            <span class="block font-semibold text-gray-900">{identity().label}</span>
                                            <span class="block text-xs text-gray-500">{identity().description}</span>
                                        </TableCell>
                                        <TableCell>{member.roleName}</TableCell>
                                        <TableCell>
                                            <Badge
                                                content={member.status}
                                                color={member.status === 'active' ? 'green' : 'dark'}
                                            />
                                        </TableCell>
                                        <TableCell class="text-right">
                                            <Button
                                                size="xs"
                                                color="red"
                                                disabled={member.roleName === 'organization_root'}
                                                onClick={() => props.removeMember(member.userId)}
                                            >
                                                Remove
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                )
                            }}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={props.pageInfo().page}
                    pageSize={props.pageInfo().pageSize}
                    total={props.pageInfo().total}
                    loading={props.loading()}
                    onPageChange={(page) => props.updateQuery({ page })}
                />
            </Show>
        </section>
    )
}
