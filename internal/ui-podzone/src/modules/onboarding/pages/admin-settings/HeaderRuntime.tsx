import { Show } from 'solid-js'
import { GW_API_URL, TENANT_GQL_URL } from '@/services/baseurl'
import { tenantStorage } from '@/services/tenantStorage'
import { InfoAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { useAdminSettings } from './context'

export function HeaderRuntime() {
  const { platformRoles, workspaceAccess } = useAdminSettings()

  return (
    <>
      <section class="flex flex-col gap-4 border-b border-gray-200 pb-5 lg:flex-row lg:items-center lg:justify-between">
        <SectionTitle
          title="Environment overview"
          subtitle="Runtime targets, local session context, and available administration access."
        />
        <div class="flex flex-wrap gap-3">
          <Button
            href="/admin/iam"
            color={platformRoles.canManage() ? 'dark' : 'light'}
            size="sm"
            disabled={!platformRoles.canManage()}
          >
            Open IAM console
          </Button>
        </div>
        <Show when={!platformRoles.canManage()}>
          <span class="self-center text-sm text-gray-500">
            Platform IAM is unavailable for this session.
          </span>
        </Show>
      </section>

      <Show when={workspaceAccess.loadingAccess()}>
        <LoadingInline label="Checking workspace access permissions..." />
      </Show>
      <Show when={platformRoles.canManage()}>
        <InfoAlert>
          Platform administration controls are enabled for this session.
        </InfoAlert>
      </Show>

      <div class="grid gap-6 lg:grid-cols-2">
        <RuntimeEndpoints />
        <LocalSessionState />
      </div>
    </>
  )
}

function RuntimeEndpoints() {
  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Runtime endpoints"
        subtitle="Current frontend targets."
      />
      <div class="space-y-3 text-sm text-gray-600">
        <div class="rounded-lg bg-gray-50 p-4">
          <p class="font-semibold text-gray-900">Gateway API</p>
          <p class="mt-1 break-all">{GW_API_URL}</p>
        </div>
        <div class="rounded-lg bg-gray-50 p-4">
          <p class="font-semibold text-gray-900">Store GraphQL API</p>
          <p class="mt-1 break-all">{TENANT_GQL_URL}</p>
        </div>
      </div>
    </Card>
  )
}

function LocalSessionState() {
  const { user } = useAdminSettings()

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Local session state"
        subtitle="Storage-backed sign-in and navigation state."
      />
      <div class="flex flex-wrap gap-2">
        <Badge
          content={user.hasToken() ? 'token present' : 'no token'}
          color={user.hasToken() ? 'green' : 'red'}
        />
        <Badge
          content={
            user.activeTenantID()
              ? `current workspace ${user.activeTenantID()}`
              : 'current workspace not set'
          }
          color={user.activeTenantID() ? 'green' : 'dark'}
        />
        <Badge
          content={
            user.sessionID() ? `session ${user.sessionID()}` : 'session missing'
          }
          color={user.sessionID() ? 'indigo' : 'red'}
        />
        <Badge
          content={
            user.routeTenantID()
              ? `last opened workspace ${user.routeTenantID()}`
              : 'last opened workspace not set'
          }
          color={user.routeTenantID() ? 'indigo' : 'dark'}
        />
      </div>
      <Button
        color="alternative"
        onClick={() => {
          tenantStorage.clearTenantID()
          user.setRouteTenantID('')
          window.location.reload()
        }}
      >
        Clear last opened workspace
      </Button>
    </Card>
  )
}
