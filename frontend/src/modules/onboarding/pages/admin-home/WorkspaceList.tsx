import { For, Show } from 'solid-js'
import { useAuthContext } from '@/modules/shell/auth-context'
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
                    <For each={vm.workspaceSummaries()}>{(workspace) => <WorkspaceCard workspace={workspace} />}</For>
                </div>
            </Show>
        </Card>
    )
}

function WorkspaceCard(props: { workspace: WorkspaceSummary }) {
    const vm = useAdminHome()
    const auth = useAuthContext()
    const workspace = props.workspace

    return (
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <div class="flex flex-wrap items-center gap-2">
                <Badge content={workspace.roleName} color="blue" />
                <Badge content={workspace.status} color={vm.membershipStatusColor(workspace.status)} />
                <Show when={workspace.tenantId === auth.getActiveTenantId()}>
                    <Badge content="current" color="yellow" />
                </Show>
            </div>
            <p class="mt-3 font-semibold text-gray-950">{workspace.tenantId}</p>
            <p class="mt-1 text-sm text-gray-600">access role for user {workspace.userId}</p>
            <p class="mt-1 text-sm text-gray-600">
                {workspace.storeCount} stores · {workspace.activeStoreCount} active
            </p>
            <div class="mt-4 space-y-3">
                <Button size="sm" color="alternative" onClick={() => vm.setSelectedWorkspaceId(workspace.tenantId)}>
                    Choose stores
                </Button>
                <CreateStoreInline tenantId={workspace.tenantId} />
            </div>
        </div>
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
                onInput={(event) => vm.setDraftStoreName(props.tenantId, event.currentTarget.value)}
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
