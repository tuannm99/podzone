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
import {
  EmptyBlock,
  ErrorAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { useAdminIamPolicy } from './context'

export function PoliciesCollection() {
  const policy = useAdminIamPolicy()

  return (
    <div class="space-y-4">
      <CollectionToolbar
        search={policy.query.search || ''}
        searchPlaceholder="Search policy name, scope, or description"
        sortBy={policy.query.sortBy || 'createdAt'}
        sortDirection={policy.query.sortDirection || 'SORT_DIRECTION_DESC'}
        pageSize={policy.query.pageSize}
        sortOptions={[
          { label: 'Created', value: 'createdAt' },
          { label: 'Updated', value: 'updatedAt' },
          { label: 'Name', value: 'name' },
          { label: 'Scope', value: 'scope' },
        ]}
        onSearch={(search) => policy.updateQuery({ search })}
        onSortByChange={(sortBy) => policy.updateQuery({ sortBy })}
        onSortDirectionChange={(sortDirection) =>
          policy.updateQuery({ sortDirection })
        }
        onPageSizeChange={(pageSize) => policy.updateQuery({ pageSize })}
      />
      <CollectionFilters
        fields={[
          {
            label: 'Scope',
            value: 'scope',
            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
          },
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
            label: 'System policy',
            value: 'isSystem',
            operators: ['FILTER_OPERATOR_EQ'],
          },
        ]}
        filters={policy.query.filters || []}
        onChange={(filters) => policy.updateQuery({ filters })}
      />
      <Show when={policy.error()}>
        <ErrorAlert>{policy.error()}</ErrorAlert>
      </Show>
      <Show when={policy.loading()}>
        <LoadingInline label="Loading policies..." />
      </Show>
      <Show
        when={policy.policies().length > 0}
        fallback={
          <EmptyBlock
            title="No policies"
            copy="No managed policies match the current collection query."
          />
        }
      >
        <DataTable>
          <TableHead>
            <TableRow>
              <TableHeaderCell>Policy</TableHeaderCell>
              <TableHeaderCell>Scope</TableHeaderCell>
              <TableHeaderCell>Default version</TableHeaderCell>
              <TableHeaderCell class="text-right">Action</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <For each={policy.policies()}>
              {(item) => (
                <TableRow>
                  <TableCell>
                    <p class="font-semibold text-gray-900">{item.name}</p>
                    <p class="mt-1 text-xs text-gray-500">
                      {item.description || 'No description'}
                    </p>
                  </TableCell>
                  <TableCell>
                    <div class="flex flex-wrap gap-2">
                      <Badge content={item.scope} color="blue" />
                      <Show when={item.isSystem}>
                        <Badge content="system" color="dark" />
                      </Show>
                    </div>
                  </TableCell>
                  <TableCell>{item.defaultVersion || 'v1'}</TableCell>
                  <TableCell class="text-right">
                    <Button
                      type="button"
                      size="xs"
                      color={
                        item.name === policy.selectedPolicyName()
                          ? 'dark'
                          : 'alternative'
                      }
                      onClick={() => policy.setSelectedPolicyName(item.name)}
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
          page={policy.pageInfo().page}
          pageSize={policy.pageInfo().pageSize}
          total={policy.pageInfo().total}
          loading={policy.loading()}
          onPageChange={(page) => policy.updateQuery({ page })}
        />
      </Show>
    </div>
  )
}
