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
import { useAdminIamPrincipal } from './principal-context'

export function PrincipalManagedPoliciesTable() {
  const principal = useAdminIamPrincipal()
  const policiesPage = createClientPagination(
    principal.currentManagedPolicies,
    6
  )

  return (
    <Show
      when={principal.currentManagedPolicies().length > 0}
      fallback={
        <EmptyBlock
          title="No direct policies"
          copy="No managed policies are attached to this principal."
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
                      principal.handleDetachPrincipalManagedPolicy(policy.name)
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
  )
}

export function PrincipalInlinePoliciesTable() {
  const principal = useAdminIamPrincipal()
  const policiesPage = createClientPagination(
    principal.currentInlinePolicies,
    6
  )

  return (
    <Show
      when={principal.currentInlinePolicies().length > 0}
      fallback={
        <EmptyBlock
          title="No inline policies"
          copy="No inline policies are attached to this principal."
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
                      principal.handleDeletePrincipalInlinePolicy(policy.name)
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
