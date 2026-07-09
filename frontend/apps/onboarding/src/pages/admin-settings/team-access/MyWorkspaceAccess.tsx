import { For, Show } from 'solid-js'
import { EmptyBlock, LoadingInline } from '@podzone/shared/ui/components/common/Feedback'
import { Badge, Button, Card } from '@podzone/shared/ui/components/common/Primitives'
import { SectionTitle } from '@podzone/shared/ui/components/common/SectionTitle'
import { useAdminSettings } from '../context'

export function MyWorkspaceAccess() {
    const { teamAccess } = useAdminSettings()
    const { access } = teamAccess

    return (
        <Card class="space-y-4">
            <SectionTitle title="My workspace access" subtitle="Workspaces this account can access right now." />
            <Show when={access.loadingMemberships()}>
                <LoadingInline label="Loading workspace access..." />
            </Show>
            <Show
                when={!access.loadingMemberships() && access.memberships().length > 0}
                fallback={
                    <EmptyBlock
                        title="No workspace access yet"
                        copy="Create or join a workspace to see your working spaces here."
                    />
                }
            >
                <div class="space-y-3">
                    <For each={access.memberships()}>
                        {(membership) => (
                            <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
                                <div class="flex flex-wrap items-center gap-2">
                                    <Badge content={membership.roleName} color="blue" />
                                    <Badge content={membership.status} color="green" />
                                </div>
                                <p class="mt-3 font-semibold text-gray-900">{membership.tenantId}</p>
                                <p class="mt-1 text-sm text-gray-500">user {membership.userId}</p>
                                <div class="mt-3">
                                    <Button
                                        size="sm"
                                        color="alternative"
                                        onClick={() => void teamAccess.reload(membership.tenantId)}
                                    >
                                        Open team access
                                    </Button>
                                </div>
                            </div>
                        )}
                    </For>
                </div>
            </Show>
        </Card>
    )
}
