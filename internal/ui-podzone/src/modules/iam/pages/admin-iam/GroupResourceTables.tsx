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
import { createClientPagination } from '@/solid/pagination'
import { useAdminIamGroup } from './group-context'

export function GroupAccessTables() {
  const group = useAdminIamGroup()
  const membersPage = createClientPagination(group.groupMembers, 8)
  const policiesPage = createClientPagination(group.groupPolicies, 8)

  return (
    <div class="grid gap-6 xl:grid-cols-2">
      <section class="min-w-0 space-y-3">
        <p class="text-sm font-semibold text-gray-900">Members</p>
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
              <For each={membersPage.pageItems()}>
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
            page={membersPage.page()}
            pageSize={membersPage.pageSize}
            total={membersPage.total()}
            onPageChange={membersPage.setPage}
          />
        </Show>
      </section>

      <section class="min-w-0 space-y-3">
        <p class="text-sm font-semibold text-gray-900">Attached policies</p>
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
              <For each={policiesPage.pageItems()}>
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
            page={policiesPage.page()}
            pageSize={policiesPage.pageSize}
            total={policiesPage.total()}
            onPageChange={policiesPage.setPage}
          />
        </Show>
      </section>
    </div>
  )
}

export function GroupInlinePoliciesTable() {
  const group = useAdminIamGroup()
  const policiesPage = createClientPagination(group.groupInlinePolicies, 6)

  return (
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
          <For each={policiesPage.pageItems()}>
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
        page={policiesPage.page()}
        pageSize={policiesPage.pageSize}
        total={policiesPage.total()}
        onPageChange={policiesPage.setPage}
      />
    </Show>
  )
}
