import { For, Show } from 'solid-js'
import {
  EmptyBlock,
  ErrorAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { CollectionToolbar } from '@/solid/components/common/CollectionToolbar'
import { CollectionFilters } from '@/solid/components/common/CollectionFilters'
import { sessionStatusColor } from './presentation'
import { useAdminSettings } from './context'

export function SessionsAudit() {
  return (
    <>
      <SessionsPanel />
      <AuditPanel />
    </>
  )
}

function SessionsPanel() {
  const vm = useAdminSettings()

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Sessions"
        subtitle="Review active sign-ins and revoke sessions that should no longer access your workspaces."
      />
      <div class="flex flex-wrap gap-3">
        <Badge content={`current ${vm.currentSessionCount()}`} color="yellow" />
        <Badge content={`other ${vm.otherSessionCount()}`} color="indigo" />
        <Button color="alternative" onClick={() => void vm.loadSessions()}>
          Reload sessions
        </Button>
      </div>
      <CollectionToolbar
        search={vm.sessionQuery.search || ''}
        searchPlaceholder="Session, workspace, status, or role"
        sortBy={vm.sessionQuery.sortBy || 'created_at'}
        sortDirection={vm.sessionQuery.sortDirection || 'SORT_DIRECTION_DESC'}
        pageSize={vm.sessionQuery.pageSize}
        sortOptions={[
          { label: 'Created time', value: 'created_at' },
          { label: 'Expiry time', value: 'expires_at' },
          { label: 'Status', value: 'status' },
          { label: 'Workspace', value: 'tenant_id' },
        ]}
        onSearch={(search) => vm.updateSessionQuery({ search })}
        onSortByChange={(sortBy) => vm.updateSessionQuery({ sortBy })}
        onSortDirectionChange={(sortDirection) =>
          vm.updateSessionQuery({ sortDirection })
        }
        onPageSizeChange={(pageSize) => vm.updateSessionQuery({ pageSize })}
      />
      <CollectionFilters
        fields={[
          {
            label: 'Session id',
            value: 'id',
            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
          },
          {
            label: 'Workspace',
            value: 'active_tenant_id',
            operators: [
              'FILTER_OPERATOR_EQ',
              'FILTER_OPERATOR_CONTAINS',
              'FILTER_OPERATOR_IN',
            ],
          },
          {
            label: 'Status',
            value: 'status',
            operators: [
              'FILTER_OPERATOR_EQ',
              'FILTER_OPERATOR_NEQ',
              'FILTER_OPERATOR_IN',
            ],
          },
          {
            label: 'Created time',
            value: 'created_at',
            operators: [
              'FILTER_OPERATOR_GT',
              'FILTER_OPERATOR_GTE',
              'FILTER_OPERATOR_LT',
              'FILTER_OPERATOR_LTE',
            ],
          },
        ]}
        filters={vm.sessionQuery.filters || []}
        onChange={(filters) => vm.updateSessionQuery({ filters })}
      />
      <Show when={vm.loadingSessions()}>
        <LoadingInline label="Loading sessions..." />
      </Show>
      <Show when={vm.sessionReadError()}>
        <ErrorAlert>{vm.sessionReadError()}</ErrorAlert>
      </Show>
      <div class="min-h-48">
        <Show
          when={!vm.sessionReadError() && vm.sessions().length > 0}
          fallback={
            <Show when={!vm.loadingSessions() && !vm.sessionReadError()}>
              <EmptyBlock
                title="No sessions loaded"
                copy="Signed-in sessions will appear here once this account starts using the backoffice."
              />
            </Show>
          }
        >
          <div class="max-h-[28rem] space-y-3 overflow-y-auto pr-1">
            <For each={vm.sessions()}>
              {(session) => (
                <div class="rounded-lg border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="break-all font-semibold text-gray-900">
                        {session.id}
                      </p>
                      <p class="mt-1 text-sm text-gray-500">
                        workspace {session.activeTenantId || 'not selected'} ·
                        expires {session.expiresAt || 'unknown'}
                      </p>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge
                        content={session.status || 'unknown'}
                        color={sessionStatusColor(session.status)}
                      />
                      <Show when={session.id === vm.sessionID()}>
                        <Badge content="current" color="yellow" />
                      </Show>
                      <Button
                        color="red"
                        size="xs"
                        onClick={() => void vm.handleRevokeSession(session.id)}
                      >
                        Revoke
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Show>
      </div>
      <Pagination
        page={vm.sessionQuery.page}
        pageSize={vm.sessionQuery.pageSize}
        total={vm.sessionPageInfo().total}
        loading={vm.loadingSessions()}
        onPageChange={(page) => vm.updateSessionQuery({ page })}
      />
    </Card>
  )
}

function AuditPanel() {
  const vm = useAdminSettings()

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Audit trail"
        subtitle="Recent access, invite, session, and platform administration actions performed by this account."
      />
      <div class="flex flex-wrap gap-3">
        <Button color="alternative" onClick={() => void vm.loadAuditLogs()}>
          Reload audit logs
        </Button>
      </div>
      <CollectionToolbar
        search={vm.auditQuery.search || ''}
        searchPlaceholder="Action, resource, workspace, or status"
        sortBy={vm.auditQuery.sortBy || 'created_at'}
        sortDirection={vm.auditQuery.sortDirection || 'SORT_DIRECTION_DESC'}
        pageSize={vm.auditQuery.pageSize}
        sortOptions={[
          { label: 'Created time', value: 'created_at' },
          { label: 'Action', value: 'action' },
          { label: 'Resource type', value: 'resource_type' },
          { label: 'Status', value: 'status' },
          { label: 'Workspace', value: 'tenant_id' },
        ]}
        onSearch={(search) => vm.updateAuditQuery({ search })}
        onSortByChange={(sortBy) => vm.updateAuditQuery({ sortBy })}
        onSortDirectionChange={(sortDirection) =>
          vm.updateAuditQuery({ sortDirection })
        }
        onPageSizeChange={(pageSize) => vm.updateAuditQuery({ pageSize })}
      />
      <CollectionFilters
        fields={[
          {
            label: 'Action',
            value: 'action',
            operators: [
              'FILTER_OPERATOR_EQ',
              'FILTER_OPERATOR_CONTAINS',
              'FILTER_OPERATOR_STARTS_WITH',
              'FILTER_OPERATOR_IN',
            ],
          },
          {
            label: 'Resource type',
            value: 'resource_type',
            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
          },
          {
            label: 'Resource id',
            value: 'resource_id',
            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
          },
          {
            label: 'Workspace',
            value: 'tenant_id',
            operators: [
              'FILTER_OPERATOR_EQ',
              'FILTER_OPERATOR_CONTAINS',
              'FILTER_OPERATOR_IN',
            ],
          },
          {
            label: 'Status',
            value: 'status',
            operators: [
              'FILTER_OPERATOR_EQ',
              'FILTER_OPERATOR_NEQ',
              'FILTER_OPERATOR_IN',
            ],
          },
          {
            label: 'Created time',
            value: 'created_at',
            operators: [
              'FILTER_OPERATOR_GT',
              'FILTER_OPERATOR_GTE',
              'FILTER_OPERATOR_LT',
              'FILTER_OPERATOR_LTE',
            ],
          },
        ]}
        filters={vm.auditQuery.filters || []}
        onChange={(filters) => vm.updateAuditQuery({ filters })}
      />
      <Show when={vm.loadingAuditLogs()}>
        <LoadingInline label="Loading audit logs..." />
      </Show>
      <Show when={vm.auditReadError()}>
        <ErrorAlert>{vm.auditReadError()}</ErrorAlert>
      </Show>
      <div class="min-h-48">
        <Show
          when={!vm.auditReadError() && vm.auditLogs().length > 0}
          fallback={
            <Show when={!vm.loadingAuditLogs() && !vm.auditReadError()}>
              <EmptyBlock
                title="No audit logs yet"
                copy="Sensitive sign-in, access, and administration actions will appear here after they run."
              />
            </Show>
          }
        >
          <div class="max-h-[36rem] space-y-3 overflow-y-auto pr-1">
            <For each={vm.auditLogs()}>
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
        page={vm.auditQuery.page}
        pageSize={vm.auditQuery.pageSize}
        total={vm.auditPageInfo().total}
        loading={vm.loadingAuditLogs()}
        onPageChange={(page) => vm.updateAuditQuery({ page })}
      />
    </Card>
  )
}
