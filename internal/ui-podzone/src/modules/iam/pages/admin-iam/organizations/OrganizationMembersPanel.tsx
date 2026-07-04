import { For, Show, type Accessor } from 'solid-js'
import type { OrganizationMembership } from '@/services/iam'
import type { CollectionQuery, PageInfo } from '@/services/collection'
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
import {
  EmptyBlock,
  ErrorAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import {
  createFormStore,
  FormInputField,
  FormSelectField,
  numberValue,
  required,
} from '@/solid/forms'

type MemberForm = {
  userId: string
  roleName: string
}

type OrganizationMembersPanelProps = {
  organizationId: Accessor<string>
  members: Accessor<OrganizationMembership[]>
  query: CollectionQuery
  pageInfo: Accessor<PageInfo>
  loading: Accessor<boolean>
  error: Accessor<string>
  updateQuery: (patch: Partial<CollectionQuery>) => void
  addMember: (userID: number, roleName: string) => Promise<void>
  removeMember: (userID: string) => Promise<void>
}

export function OrganizationMembersPanel(props: OrganizationMembersPanelProps) {
  const form = createFormStore<MemberForm>({
    initialValues: {
      userId: '',
      roleName: 'organization_viewer',
    },
    validators: {
      userId: [
        required('Enter a user ID.'),
        numberValue('Enter a valid user ID.'),
      ],
      roleName: [required('Select an organization role.')],
    },
  })

  const submit = async (event: SubmitEvent) => {
    event.preventDefault()
    if (!form.validate()) return
    form.setSubmitting(true)
    try {
      await props.addMember(
        Number.parseInt(form.values.userId, 10),
        form.values.roleName
      )
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
          <FormInputField form={form} name="userId" label="User ID" />
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
          onSortDirectionChange={(sortDirection) =>
            props.updateQuery({ sortDirection })
          }
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
              {(member) => (
                <TableRow>
                  <TableCell class="font-semibold text-gray-900">
                    {member.userId}
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
              )}
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
