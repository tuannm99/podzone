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
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { useAdminHome } from './context'

export function ProvisioningRequestsPanel() {
  const vm = useAdminHome()

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Provisioning requests"
        subtitle="Track store infrastructure planning, approval, provisioning, and finalization."
      />
      <Show
        when={vm.selectedWorkspace()}
        fallback={
          <EmptyBlock
            title="No workspace selected"
            copy="Choose a workspace to inspect its provisioning pipeline."
          />
        }
      >
        <CollectionControls
          query={vm.storeRequests.query}
          loading={vm.storeRequests.loading}
          error={vm.storeRequestsError}
          searchPlaceholder="Search request, subdomain, requester, or status"
          sortOptions={[
            { label: 'Updated', value: 'updatedAt' },
            { label: 'Created', value: 'createdAt' },
            { label: 'Name', value: 'name' },
            { label: 'Status', value: 'status' },
          ]}
          filterFields={[
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
              label: 'Subdomain',
              value: 'subdomain',
              operators: [
                'FILTER_OPERATOR_EQ',
                'FILTER_OPERATOR_CONTAINS',
                'FILTER_OPERATOR_STARTS_WITH',
              ],
            },
            {
              label: 'Status',
              value: 'status',
              operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
            },
          ]}
          updateQuery={vm.storeRequests.updateQuery}
        />
        <Show
          when={vm.storeRequests.items().length > 0}
          fallback={
            <Show when={!vm.storeRequests.loading()}>
              <EmptyBlock
                title="No provisioning requests"
                copy="No store requests match the current collection query."
              />
            </Show>
          }
        >
          <DataTable>
            <TableHead>
              <TableRow>
                <TableHeaderCell>Request</TableHeaderCell>
                <TableHeaderCell>Status</TableHeaderCell>
                <TableHeaderCell>Requester</TableHeaderCell>
                <TableHeaderCell>Updated</TableHeaderCell>
                <TableHeaderCell class="text-right">Action</TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <For each={vm.storeRequests.items()}>
                {(request) => (
                  <TableRow>
                    <TableCell>
                      <p class="font-medium text-gray-900">{request.name}</p>
                      <p class="text-xs text-gray-500">{request.subdomain}</p>
                      <Show when={request.last_error}>
                        <p class="mt-1 text-xs text-red-700">
                          {request.last_error}
                        </p>
                      </Show>
                    </TableCell>
                    <TableCell>
                      <Badge
                        content={vm.provisioningStatusLabel(request.status)}
                        color={
                          request.status === 'ready'
                            ? 'green'
                            : request.status.startsWith('failed')
                              ? 'red'
                              : 'yellow'
                        }
                      />
                    </TableCell>
                    <TableCell>{request.requested_by}</TableCell>
                    <TableCell>{request.updated_at}</TableCell>
                    <TableCell class="text-right">
                      <Button
                        size="xs"
                        color="alternative"
                        loading={vm.retryingStoreRequestId() === request.id}
                        disabled={
                          vm.retryingStoreRequestId() === request.id ||
                          !['failed', 'failed_retryable'].includes(
                            request.status
                          )
                        }
                        onClick={() =>
                          void vm.retryStore(
                            vm.selectedWorkspaceId(),
                            request.id
                          )
                        }
                      >
                        Retry
                      </Button>
                    </TableCell>
                  </TableRow>
                )}
              </For>
            </TableBody>
          </DataTable>
          <Pagination
            page={vm.storeRequests.pageInfo().page}
            pageSize={vm.storeRequests.pageInfo().pageSize}
            total={vm.storeRequests.pageInfo().total}
            loading={vm.storeRequests.loading()}
            onPageChange={(page) => vm.storeRequests.updateQuery({ page })}
          />
        </Show>
      </Show>
    </Card>
  )
}
