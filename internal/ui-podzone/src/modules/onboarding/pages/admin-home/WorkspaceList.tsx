import { For, Show } from 'solid-js'
import { tokenStorage } from '@/services/tokenStorage'
import { EmptyBlock, LoadingInline } from '@/solid/components/common/Feedback'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { useAdminHome } from './context'
import type { WorkspaceSummary } from './presentation'

export function WorkspaceList() {
  const vm = useAdminHome()

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="My workspaces"
        subtitle="Pick one store to enter. To switch later, return here and choose another store."
      />
      <Show when={vm.loadingTenants()}>
        <LoadingInline label="Loading workspaces..." />
      </Show>
      <Show
        when={!vm.loadingTenants() && vm.workspaceSummaries().length > 0}
        fallback={
          <EmptyBlock
            title="No workspaces yet"
            copy="Create your first tenant workspace to start managing stores, catalog, team access, and POD operations."
          />
        }
      >
        <div class="space-y-3">
          <For each={vm.workspaceSummaries()}>
            {(workspace) => <WorkspaceCard workspace={workspace} />}
          </For>
        </div>
      </Show>
    </Card>
  )
}

function WorkspaceCard(props: { workspace: WorkspaceSummary }) {
  const vm = useAdminHome()
  const workspace = props.workspace

  return (
    <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
      <div class="flex flex-wrap items-center gap-2">
        <Badge content={workspace.roleName} color="blue" />
        <Badge
          content={workspace.status}
          color={vm.membershipStatusColor(workspace.status)}
        />
        <Show when={workspace.tenantId === tokenStorage.getActiveTenantID()}>
          <Badge content="current" color="yellow" />
        </Show>
      </div>
      <p class="mt-3 font-semibold text-gray-950">{workspace.tenantId}</p>
      <p class="mt-1 text-sm text-gray-600">
        access role for user {workspace.userId}
      </p>
      <p class="mt-1 text-sm text-gray-600">
        {workspace.storeCount} stores · {workspace.activeStoreCount} active
      </p>
      <div class="mt-4 space-y-3">
        <StoreRequests workspace={workspace} />
        <Stores workspace={workspace} />
        <CreateStoreInline tenantId={workspace.tenantId} />
      </div>
    </div>
  )
}

function StoreRequests(props: { workspace: WorkspaceSummary }) {
  const vm = useAdminHome()

  return (
    <Show when={props.workspace.storeRequests.length > 0}>
      <For each={props.workspace.storeRequests}>
        {(request) => (
          <div class="rounded-md border border-gray-200 bg-white p-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-gray-950">
                  {request.name}
                </div>
                <div class="mt-1 text-xs text-gray-500">
                  {request.subdomain}
                </div>
              </div>
              <Badge
                content={vm.provisioningStatusLabel(request.status)}
                color={
                  request.status === 'ready'
                    ? 'green'
                    : request.status === 'failed'
                      ? 'red'
                      : 'yellow'
                }
              />
            </div>
            <div
              class="mt-3 grid grid-cols-4 gap-1"
              aria-label={`Provisioning progress: ${vm.provisioningStatusLabel(request.status)}`}
            >
              <For each={vm.provisioningSteps}>
                {(step, index) => (
                  <div>
                    <div
                      class={`h-1.5 rounded-full ${
                        index() <= vm.provisioningStepIndex(request.status)
                          ? request.status === 'failed'
                            ? 'bg-red-500'
                            : 'bg-gray-950'
                          : 'bg-gray-200'
                      }`}
                    />
                    <p class="mt-1 truncate text-[11px] text-gray-500">
                      {vm.provisioningStatusLabel(step)}
                    </p>
                  </div>
                )}
              </For>
            </div>
            <Show when={request.last_error}>
              <p class="mt-2 text-sm text-red-700">{request.last_error}</p>
            </Show>
            <Show when={request.status === 'failed'}>
              <div class="mt-3">
                <Button
                  size="xs"
                  color="alternative"
                  loading={vm.retryingStoreRequestId() === request.id}
                  disabled={vm.retryingStoreRequestId() === request.id}
                  onClick={() => {
                    void vm.retryStore(props.workspace.tenantId, request.id)
                  }}
                >
                  Retry provisioning
                </Button>
              </div>
            </Show>
          </div>
        )}
      </For>
    </Show>
  )
}

function Stores(props: { workspace: WorkspaceSummary }) {
  const vm = useAdminHome()

  return (
    <Show
      when={props.workspace.stores.length > 0}
      fallback={
        <EmptyBlock
          title="No stores in this workspace"
          copy="Create the first store here, then the backoffice will open in that store scope."
        />
      }
    >
      <For each={props.workspace.stores}>
        {(store) => (
          <div class="flex flex-col gap-3 rounded-md border border-gray-200 bg-white p-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="min-w-0">
              <div class="truncate text-sm font-semibold text-gray-950">
                {store.name}
              </div>
              <div class="mt-1 flex flex-wrap gap-2">
                <Badge
                  content={store.status || 'unknown'}
                  color={store.isActive ? 'green' : 'dark'}
                />
                <Badge content={store.id} color="dark" />
              </div>
            </div>
            <Button
              size="sm"
              onClick={() => {
                void vm.openStore(props.workspace.tenantId, store.id)
              }}
            >
              Open store
            </Button>
          </div>
        )}
      </For>
    </Show>
  )
}

function CreateStoreInline(props: { tenantId: string }) {
  const vm = useAdminHome()

  return (
    <div class="flex flex-col gap-3 rounded-md border border-dashed border-gray-300 bg-white p-3 sm:flex-row">
      <input
        class="block h-9 min-w-0 flex-1 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-900 outline-none transition focus:border-gray-950 focus:ring-2 focus:ring-gray-100"
        value={vm.storeNameByTenant()[props.tenantId] || ''}
        placeholder="New store name"
        onInput={(event) =>
          vm.setDraftStoreName(props.tenantId, event.currentTarget.value)
        }
      />
      <Button
        size="sm"
        color="alternative"
        loading={vm.creatingStoreTenantId() === props.tenantId}
        disabled={
          vm.creatingStoreTenantId() === props.tenantId ||
          !(vm.storeNameByTenant()[props.tenantId] || '').trim()
        }
        onClick={() => {
          void vm.submitCreateStore(props.tenantId)
        }}
      >
        Create store
      </Button>
    </div>
  )
}
