import { For, Show } from 'solid-js'
import { CollectionFilters } from '@/solid/components/common/CollectionFilters'
import { CollectionToolbar } from '@/solid/components/common/CollectionToolbar'
import {
  EmptyBlock,
  ErrorAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { useAdminSettings } from '../context'
import { auditFilterFields, auditSortOptions } from './audit.collection'

export function AuditPanel() {
  const { audit } = useAdminSettings()

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Audit trail"
        subtitle="Recent access, invite, session, and platform administration actions performed by this account."
      />
      <Button color="alternative" onClick={() => void audit.reload()}>
        Reload audit logs
      </Button>
      <CollectionToolbar
        search={audit.query.search || ''}
        searchPlaceholder="Action, resource, workspace, or status"
        sortBy={audit.query.sortBy || 'created_at'}
        sortDirection={audit.query.sortDirection || 'SORT_DIRECTION_DESC'}
        pageSize={audit.query.pageSize}
        sortOptions={auditSortOptions}
        onSearch={(search) => audit.updateQuery({ search })}
        onSortByChange={(sortBy) => audit.updateQuery({ sortBy })}
        onSortDirectionChange={(sortDirection) =>
          audit.updateQuery({ sortDirection })
        }
        onPageSizeChange={(pageSize) => audit.updateQuery({ pageSize })}
      />
      <CollectionFilters
        fields={auditFilterFields}
        filters={audit.query.filters || []}
        onChange={(filters) => audit.updateQuery({ filters })}
      />
      <Show when={audit.loading()}>
        <LoadingInline label="Loading audit logs..." />
      </Show>
      <Show when={audit.error()}>
        <ErrorAlert>{audit.error()}</ErrorAlert>
      </Show>
      <div class="min-h-48">
        <Show
          when={!audit.error() && audit.items().length > 0}
          fallback={
            <Show when={!audit.loading() && !audit.error()}>
              <EmptyBlock
                title="No audit logs yet"
                copy="Sensitive sign-in, access, and administration actions will appear here after they run."
              />
            </Show>
          }
        >
          <div class="max-h-[36rem] space-y-3 overflow-y-auto pr-1">
            <For each={audit.items()}>
              {(log) => (
                <div class="rounded-lg border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="font-semibold text-gray-900">
                        {log.action || 'unknown action'}
                      </p>
                      <p class="mt-1 text-sm text-gray-500">
                        {log.resourceType || 'resource'}{' '}
                        {log.resourceId || 'n/a'} · workspace{' '}
                        {log.tenantId || 'global'}
                      </p>
                      <Show when={log.payloadJson}>
                        <pre class="mt-3 overflow-x-auto rounded-lg bg-gray-50 p-3 text-xs text-gray-700">
                          {log.payloadJson}
                        </pre>
                      </Show>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge
                        content={log.status || 'unknown'}
                        color={log.status === 'success' ? 'green' : 'dark'}
                      />
                      <Badge
                        content={log.createdAt || 'time unknown'}
                        color="indigo"
                      />
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Show>
      </div>
      <Pagination
        page={audit.query.page}
        pageSize={audit.query.pageSize}
        total={audit.pageInfo().total}
        loading={audit.loading()}
        onPageChange={(page) => audit.updateQuery({ page })}
      />
    </Card>
  )
}
