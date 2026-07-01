import { For, Show, type Accessor } from 'solid-js'
import type { OrganizationInfo, PolicyInfo } from '@/services/iam'
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

export type OrganizationsCollectionProps = {
  organizations: Accessor<OrganizationInfo[]>
  query: CollectionQuery
  pageInfo: Accessor<PageInfo>
  loading: Accessor<boolean>
  error: Accessor<string>
  updateQuery: (patch: Partial<CollectionQuery>) => void
  selectedOrgId: Accessor<string>
  orgTenantId: Accessor<string>
  orgPolicies: Accessor<PolicyInfo[]>
  handleDetachTenantFromOrg: (tenantID: string) => void
  handleDetachScp: (policyName: string) => void
}

export function OrganizationsCollection(props: OrganizationsCollectionProps) {
  return (
    <div class="space-y-4">
      <CollectionToolbar
        search={props.query.search || ''}
        searchPlaceholder="Search name, slug, or organization ID"
        sortBy={props.query.sortBy || 'createdAt'}
        sortDirection={props.query.sortDirection || 'SORT_DIRECTION_DESC'}
        pageSize={props.query.pageSize}
        sortOptions={[
          { label: 'Created', value: 'createdAt' },
          { label: 'Updated', value: 'updatedAt' },
          { label: 'Name', value: 'name' },
          { label: 'Slug', value: 'slug' },
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
            label: 'Name',
            value: 'name',
            operators: [
              'FILTER_OPERATOR_EQ',
              'FILTER_OPERATOR_CONTAINS',
              'FILTER_OPERATOR_STARTS_WITH',
            ],
          },
          {
            label: 'Slug',
            value: 'slug',
            operators: [
              'FILTER_OPERATOR_EQ',
              'FILTER_OPERATOR_CONTAINS',
              'FILTER_OPERATOR_STARTS_WITH',
            ],
          },
          {
            label: 'Organization ID',
            value: 'id',
            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
          },
        ]}
        filters={props.query.filters || []}
        onChange={(filters) => props.updateQuery({ filters })}
      />
      <Show when={props.error()}>
        <ErrorAlert>{props.error()}</ErrorAlert>
      </Show>
      <Show when={props.loading()}>
        <LoadingInline label="Loading organizations..." />
      </Show>
      <Show
        when={props.organizations().length > 0}
        fallback={
          <EmptyBlock
            title="No organizations"
            copy="No organizations match the current collection query."
          />
        }
      >
        <DataTable>
          <TableHead>
            <TableRow>
              <TableHeaderCell>Organization</TableHeaderCell>
              <TableHeaderCell>Slug</TableHeaderCell>
              <TableHeaderCell>Status</TableHeaderCell>
              <TableHeaderCell class="text-right">Actions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <For each={props.organizations()}>
              {(organization) => (
                <TableRow>
                  <TableCell class="font-semibold text-gray-900">
                    {organization.name}
                  </TableCell>
                  <TableCell class="text-gray-600">
                    {organization.slug}
                  </TableCell>
                  <TableCell>
                    <Badge
                      content={
                        organization.id === props.selectedOrgId()
                          ? 'selected'
                          : 'organization'
                      }
                      color={
                        organization.id === props.selectedOrgId()
                          ? 'blue'
                          : 'dark'
                      }
                    />
                  </TableCell>
                  <TableCell>
                    <div class="flex flex-wrap justify-end gap-2">
                      <Show
                        when={
                          organization.id === props.selectedOrgId() &&
                          props.orgTenantId().trim()
                        }
                      >
                        <Button
                          size="xs"
                          color="light"
                          onClick={() =>
                            props.handleDetachTenantFromOrg(
                              props.orgTenantId().trim()
                            )
                          }
                        >
                          Detach selected workspace
                        </Button>
                      </Show>
                      <For
                        each={
                          organization.id === props.selectedOrgId()
                            ? props.orgPolicies()
                            : []
                        }
                      >
                        {(policy) => (
                          <Button
                            size="xs"
                            color="alternative"
                            onClick={() => props.handleDetachScp(policy.name)}
                          >
                            Detach SCP {policy.name}
                          </Button>
                        )}
                      </For>
                    </div>
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
    </div>
  )
}
