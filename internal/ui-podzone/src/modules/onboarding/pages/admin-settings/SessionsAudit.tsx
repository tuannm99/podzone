import { For, Show } from 'solid-js'
import {
  EmptyBlock,
  ErrorAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { createClientPagination } from '@/solid/pagination'
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
  const pagination = createClientPagination(vm.sessions, 5)

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
      <Show when={vm.loadingSessions()}>
        <LoadingInline label="Loading sessions..." />
      </Show>
      <Show when={vm.sessionReadError()}>
        <ErrorAlert>{vm.sessionReadError()}</ErrorAlert>
      </Show>
      <Show
        when={
          !vm.loadingSessions() &&
          !vm.sessionReadError() &&
          vm.sessions().length > 0
        }
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
          <For each={pagination.pageItems()}>
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
      <Pagination
        page={pagination.page()}
        pageSize={pagination.pageSize}
        total={pagination.total()}
        onPageChange={pagination.setPage}
      />
    </Card>
  )
}

function AuditPanel() {
  const vm = useAdminSettings()
  const pagination = createClientPagination(vm.auditLogs, 6)

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
      <Show when={vm.loadingAuditLogs()}>
        <LoadingInline label="Loading audit logs..." />
      </Show>
      <Show when={vm.auditReadError()}>
        <ErrorAlert>{vm.auditReadError()}</ErrorAlert>
      </Show>
      <Show
        when={
          !vm.loadingAuditLogs() &&
          !vm.auditReadError() &&
          vm.auditLogs().length > 0
        }
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
          <For each={pagination.pageItems()}>
            {(log) => (
              <div class="rounded-lg border border-gray-200 p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p class="font-semibold text-gray-900">
                      {log.action || 'unknown action'}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                      {log.resourceType || 'resource'} {log.resourceId || 'n/a'}{' '}
                      · workspace {log.tenantId || 'global'}
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
      <Pagination
        page={pagination.page()}
        pageSize={pagination.pageSize}
        total={pagination.total()}
        onPageChange={pagination.setPage}
      />
    </Card>
  )
}
