import { For, Show } from 'solid-js'
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
import { CollectionControls } from '@/solid/components/common/CollectionControls'
import { useAdminIamGroup } from './context'

export function GroupAccessTables() {
  const group = useAdminIamGroup()

  return (
    <div class="grid gap-6 xl:grid-cols-2">
      <section class="min-w-0 space-y-3">
        <p class="text-sm font-semibold text-gray-900">Members</p>
        <CollectionControls
          query={group.groupMembersQuery}
          loading={group.groupMembersLoading}
          error={group.groupMembersError}
          searchPlaceholder="Search user ID"
          sortOptions={[
            { label: 'Added', value: 'createdAt' },
            { label: 'User ID', value: 'userId' },
          ]}
          filterFields={[
            {
              label: 'User ID',
              value: 'userId',
              operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
            },
          ]}
          updateQuery={group.updateGroupMembersQuery}
        />
        <Show
          when={group.groupMembers().length > 0}
          fallback={
            <EmptyBlock
              title="No group members"
              copy="Select a group and add users."
            />
          }
        >
          <DataTable>
            <TableHead>
              <TableRow>
                <TableHeaderCell>User ID</TableHeaderCell>
                <TableHeaderCell class="text-right">Action</TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <For each={group.groupMembers()}>
                {(userID) => (
                  <TableRow>
                    <TableCell class="font-medium text-gray-900">
                      {userID}
                    </TableCell>
                    <TableCell class="text-right">
                      <Button
                        size="xs"
                        color="red"
                        onClick={() => group.handleRemoveGroupMember(userID)}
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
            page={group.groupMembersPageInfo().page}
            pageSize={group.groupMembersPageInfo().pageSize}
            total={group.groupMembersPageInfo().total}
            loading={group.groupMembersLoading()}
            onPageChange={(page) => group.updateGroupMembersQuery({ page })}
          />
        </Show>
      </section>

      <section class="min-w-0 space-y-3">
        <p class="text-sm font-semibold text-gray-900">Attached policies</p>
        <CollectionControls
          query={group.groupPoliciesQuery}
          loading={group.groupPoliciesLoading}
          error={group.groupPoliciesError}
          searchPlaceholder="Search policy"
          sortOptions={[
            { label: 'Created', value: 'createdAt' },
            { label: 'Name', value: 'name' },
            { label: 'Scope', value: 'scope' },
          ]}
          filterFields={[
            {
              label: 'Scope',
              value: 'scope',
              operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
            },
            {
              label: 'Name',
              value: 'name',
              operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
            },
          ]}
          updateQuery={group.updateGroupPoliciesQuery}
        />
        <Show
          when={group.groupPolicies().length > 0}
          fallback={
            <EmptyBlock
              title="No group policies"
              copy="Attach a managed policy to this group."
            />
          }
        >
          <DataTable>
            <TableHead>
              <TableRow>
                <TableHeaderCell>Policy</TableHeaderCell>
                <TableHeaderCell>Scope</TableHeaderCell>
                <TableHeaderCell class="text-right">Action</TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <For each={group.groupPolicies()}>
                {(policy) => (
                  <TableRow>
                    <TableCell class="font-medium text-gray-900">
                      {policy.name}
                    </TableCell>
                    <TableCell>
                      <Badge content={policy.scope} color="blue" />
                    </TableCell>
                    <TableCell class="text-right">
                      <Button
                        size="xs"
                        color="red"
                        onClick={() =>
                          group.handleDetachGroupPolicy(policy.name)
                        }
                      >
                        Detach
                      </Button>
                    </TableCell>
                  </TableRow>
                )}
              </For>
            </TableBody>
          </DataTable>
          <Pagination
            page={group.groupPoliciesPageInfo().page}
            pageSize={group.groupPoliciesPageInfo().pageSize}
            total={group.groupPoliciesPageInfo().total}
            loading={group.groupPoliciesLoading()}
            onPageChange={(page) => group.updateGroupPoliciesQuery({ page })}
          />
        </Show>
      </section>
    </div>
  )
}

export function GroupInlinePoliciesTable() {
  const group = useAdminIamGroup()

  return (
    <div class="space-y-3">
      <CollectionControls
        query={group.groupInlinePoliciesQuery}
        loading={group.groupInlinePoliciesLoading}
        error={group.groupInlinePoliciesError}
        searchPlaceholder="Search inline policy"
        sortOptions={[
          { label: 'Created', value: 'createdAt' },
          { label: 'Updated', value: 'updatedAt' },
          { label: 'Name', value: 'name' },
        ]}
        filterFields={[
          {
            label: 'Name',
            value: 'name',
            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
          },
        ]}
        updateQuery={group.updateGroupInlinePoliciesQuery}
      />
      <Show
        when={group.groupInlinePolicies().length > 0}
        fallback={
          <EmptyBlock
            title="No group inline policies"
            copy="No inline policies are attached to this group."
          />
        }
      >
        <DataTable>
          <TableHead>
            <TableRow>
              <TableHeaderCell>Policy</TableHeaderCell>
              <TableHeaderCell>Description</TableHeaderCell>
              <TableHeaderCell>Statements</TableHeaderCell>
              <TableHeaderCell class="text-right">Action</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <For each={group.groupInlinePolicies()}>
              {(policy) => (
                <TableRow>
                  <TableCell class="font-medium text-gray-900">
                    {policy.name}
                  </TableCell>
                  <TableCell class="text-gray-600">
                    {policy.description || 'No description'}
                  </TableCell>
                  <TableCell class="text-gray-600">
                    {policy.statements?.length || 0}
                  </TableCell>
                  <TableCell class="text-right">
                    <Button
                      size="xs"
                      color="red"
                      onClick={() =>
                        group.handleDeleteGroupInlinePolicy(policy.name)
                      }
                    >
                      Delete
                    </Button>
                  </TableCell>
                </TableRow>
              )}
            </For>
          </TableBody>
        </DataTable>
        <Pagination
          page={group.groupInlinePoliciesPageInfo().page}
          pageSize={group.groupInlinePoliciesPageInfo().pageSize}
          total={group.groupInlinePoliciesPageInfo().total}
          loading={group.groupInlinePoliciesLoading()}
          onPageChange={(page) =>
            group.updateGroupInlinePoliciesQuery({ page })
          }
        />
      </Show>
    </div>
  )
}
